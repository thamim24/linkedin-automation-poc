package state

import (
	"time"
)

// Store defines the interface for state persistence
type Store interface {
	// Profile operations
	SaveProfile(profile Profile) error
	GetProfile(url string) (*Profile, error)
	HasProfile(url string) (bool, error)
	ListProfiles(limit int) ([]Profile, error)

	// Action operations
	SaveAction(action Action) error
	GetActionsByDate(date time.Time) ([]Action, error)
	GetDailyCount(actionType ActionType, date time.Time) (int, error)

	// Session operations
	SaveSession(session Session) error
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error

	// Statistics
	GetStats() (*Stats, error)

	// Cleanup
	Close() error
}

// Profile represents a LinkedIn profile record
type Profile struct {
	ID               int64
	ProfileURL       string
	Name             string
	Headline         string
	Location         string
	ConnectionDegree string
	FirstSeen        time.Time
	LastInteraction  time.Time
	InteractionCount int
	IsConnected      bool
	Notes            string
}

// ActionType represents types of actions
type ActionType string

const (
	ActionTypeConnection ActionType = "connection"
	ActionTypeMessage    ActionType = "message"
	ActionTypeSearch     ActionType = "search"
	ActionTypeView       ActionType = "view"
)

// Action represents an interaction record
type Action struct {
	ID         int64
	ProfileURL string
	ActionType ActionType
	Success    bool
	Message    string
	Timestamp  time.Time
	SessionID  string
}

// Session represents a browser session
type Session struct {
	ID        int64
	SessionID string
	Cookies   []byte // JSON serialized cookies
	CreatedAt time.Time
	ExpiresAt time.Time
	IsValid   bool
}

// Stats represents aggregated statistics
type Stats struct {
	TotalProfiles      int
	TotalConnections   int
	TotalMessages      int
	TotalSearches      int
	ConnectionsToday   int
	MessagesToday      int
	SearchesToday      int
	LastActionTime     time.Time
	ProfilesInteracted int
}
