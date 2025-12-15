package state

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"linkedin-automation-poc/internal/log"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite
type SQLiteStore struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	logger := log.Module("state")

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{
		db:     db,
		logger: logger,
	}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info("database initialized", zap.String("path", dbPath))

	return store, nil
}

// initSchema creates database tables
func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT UNIQUE NOT NULL,
		name TEXT,
		headline TEXT,
		location TEXT,
		connection_degree TEXT,
		first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_interaction TIMESTAMP,
		interaction_count INTEGER DEFAULT 0,
		is_connected BOOLEAN DEFAULT 0,
		notes TEXT
	);
	
	CREATE TABLE IF NOT EXISTS actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT NOT NULL,
		action_type TEXT NOT NULL,
		success BOOLEAN DEFAULT 1,
		message TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT,
		FOREIGN KEY (profile_url) REFERENCES profiles(profile_url)
	);
	
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		cookies BLOB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP,
		is_valid BOOLEAN DEFAULT 1
	);
	
	CREATE INDEX IF NOT EXISTS idx_profiles_url ON profiles(profile_url);
	CREATE INDEX IF NOT EXISTS idx_actions_profile ON actions(profile_url);
	CREATE INDEX IF NOT EXISTS idx_actions_timestamp ON actions(timestamp);
	CREATE INDEX IF NOT EXISTS idx_actions_type ON actions(action_type);
	CREATE INDEX IF NOT EXISTS idx_sessions_id ON sessions(session_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveProfile saves or updates a profile
func (s *SQLiteStore) SaveProfile(profile Profile) error {
	query := `
	INSERT INTO profiles (profile_url, name, headline, location, connection_degree, is_connected, notes)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(profile_url) DO UPDATE SET
		name = excluded.name,
		headline = excluded.headline,
		location = excluded.location,
		connection_degree = excluded.connection_degree,
		is_connected = excluded.is_connected,
		notes = excluded.notes
	`

	_, err := s.db.Exec(query,
		profile.ProfileURL,
		profile.Name,
		profile.Headline,
		profile.Location,
		profile.ConnectionDegree,
		profile.IsConnected,
		profile.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	s.logger.Debug("profile saved", zap.String("url", profile.ProfileURL))
	return nil
}

// GetProfile retrieves a profile by URL
func (s *SQLiteStore) GetProfile(url string) (*Profile, error) {
	query := `
	SELECT id, profile_url, name, headline, location, connection_degree,
		   first_seen, last_interaction, interaction_count, is_connected, notes
	FROM profiles WHERE profile_url = ?
	`

	var profile Profile
	var lastInteraction sql.NullTime

	err := s.db.QueryRow(query, url).Scan(
		&profile.ID,
		&profile.ProfileURL,
		&profile.Name,
		&profile.Headline,
		&profile.Location,
		&profile.ConnectionDegree,
		&profile.FirstSeen,
		&lastInteraction,
		&profile.InteractionCount,
		&profile.IsConnected,
		&profile.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if lastInteraction.Valid {
		profile.LastInteraction = lastInteraction.Time
	}

	return &profile, nil
}

// HasProfile checks if profile exists
func (s *SQLiteStore) HasProfile(url string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM profiles WHERE profile_url = ?", url).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListProfiles lists recent profiles
func (s *SQLiteStore) ListProfiles(limit int) ([]Profile, error) {
	query := `
	SELECT id, profile_url, name, headline, location, connection_degree,
		   first_seen, last_interaction, interaction_count, is_connected, notes
	FROM profiles
	ORDER BY last_interaction DESC, first_seen DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var profile Profile
		var lastInteraction sql.NullTime

		err := rows.Scan(
			&profile.ID,
			&profile.ProfileURL,
			&profile.Name,
			&profile.Headline,
			&profile.Location,
			&profile.ConnectionDegree,
			&profile.FirstSeen,
			&lastInteraction,
			&profile.InteractionCount,
			&profile.IsConnected,
			&profile.Notes,
		)
		if err != nil {
			continue
		}

		if lastInteraction.Valid {
			profile.LastInteraction = lastInteraction.Time
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// SaveAction records an action
func (s *SQLiteStore) SaveAction(action Action) error {
	query := `
	INSERT INTO actions (profile_url, action_type, success, message, session_id)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		action.ProfileURL,
		action.ActionType,
		action.Success,
		action.Message,
		action.SessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to save action: %w", err)
	}

	// Update profile interaction count and time
	updateQuery := `
	UPDATE profiles
	SET last_interaction = CURRENT_TIMESTAMP,
		interaction_count = interaction_count + 1
	WHERE profile_url = ?
	`
	s.db.Exec(updateQuery, action.ProfileURL)

	s.logger.Debug("action saved",
		zap.String("type", string(action.ActionType)),
		zap.String("profile", action.ProfileURL),
	)

	return nil
}

// GetActionsByDate retrieves actions for a specific date
func (s *SQLiteStore) GetActionsByDate(date time.Time) ([]Action, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
	SELECT id, profile_url, action_type, success, message, timestamp, session_id
	FROM actions
	WHERE timestamp >= ? AND timestamp < ?
	ORDER BY timestamp DESC
	`

	rows, err := s.db.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}
	defer rows.Close()

	var actions []Action
	for rows.Next() {
		var action Action
		var message sql.NullString
		var sessionID sql.NullString

		err := rows.Scan(
			&action.ID,
			&action.ProfileURL,
			&action.ActionType,
			&action.Success,
			&message,
			&action.Timestamp,
			&sessionID,
		)
		if err != nil {
			continue
		}

		if message.Valid {
			action.Message = message.String
		}
		if sessionID.Valid {
			action.SessionID = sessionID.String
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// GetDailyCount returns count of actions for a specific day
func (s *SQLiteStore) GetDailyCount(actionType ActionType, date time.Time) (int, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
	SELECT COUNT(*)
	FROM actions
	WHERE action_type = ? AND timestamp >= ? AND timestamp < ? AND success = 1
	`

	var count int
	err := s.db.QueryRow(query, actionType, startOfDay, endOfDay).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SaveSession saves session data
func (s *SQLiteStore) SaveSession(session Session) error {
	query := `
	INSERT INTO sessions (session_id, cookies, expires_at, is_valid)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(session_id) DO UPDATE SET
		cookies = excluded.cookies,
		expires_at = excluded.expires_at,
		is_valid = excluded.is_valid
	`

	_, err := s.db.Exec(query,
		session.SessionID,
		session.Cookies,
		session.ExpiresAt,
		session.IsValid,
	)

	return err
}

// GetSession retrieves session by ID
func (s *SQLiteStore) GetSession(sessionID string) (*Session, error) {
	query := `
	SELECT id, session_id, cookies, created_at, expires_at, is_valid
	FROM sessions WHERE session_id = ?
	`

	var session Session
	err := s.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.Cookies,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsValid,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session
func (s *SQLiteStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}

// GetStats returns aggregated statistics
func (s *SQLiteStore) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Total profiles
	s.db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&stats.TotalProfiles)

	// Total actions by type
	s.db.QueryRow("SELECT COUNT(*) FROM actions WHERE action_type = ? AND success = 1", ActionTypeConnection).Scan(&stats.TotalConnections)
	s.db.QueryRow("SELECT COUNT(*) FROM actions WHERE action_type = ? AND success = 1", ActionTypeMessage).Scan(&stats.TotalMessages)
	s.db.QueryRow("SELECT COUNT(*) FROM actions WHERE action_type = ? AND success = 1", ActionTypeSearch).Scan(&stats.TotalSearches)

	// Today's counts
	today := time.Now()
	stats.ConnectionsToday, _ = s.GetDailyCount(ActionTypeConnection, today)
	stats.MessagesToday, _ = s.GetDailyCount(ActionTypeMessage, today)
	stats.SearchesToday, _ = s.GetDailyCount(ActionTypeSearch, today)

	// Last action time
	var lastAction sql.NullTime
	s.db.QueryRow("SELECT MAX(timestamp) FROM actions").Scan(&lastAction)
	if lastAction.Valid {
		stats.LastActionTime = lastAction.Time
	}

	// Profiles interacted with
	s.db.QueryRow("SELECT COUNT(*) FROM profiles WHERE interaction_count > 0").Scan(&stats.ProfilesInteracted)

	return stats, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	s.logger.Info("closing database")
	return s.db.Close()
}
