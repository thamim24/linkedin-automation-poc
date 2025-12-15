package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"linkedin-automation-poc/internal/log"

	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

// Session manages browser session persistence
type Session struct {
	sessionID string
	cookies   []*proto.NetworkCookie
	logger    *zap.Logger
}

// SessionData represents serializable session data
type SessionData struct {
	SessionID string                 `json:"session_id"`
	Cookies   []*proto.NetworkCookie `json:"cookies"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// NewSession creates a new session manager
func NewSession(sessionID string) *Session {
	return &Session{
		sessionID: sessionID,
		logger:    log.Session(sessionID),
	}
}

// Save extracts and stores session cookies
func (s *Session) Save(cookies []*proto.NetworkCookie) error {
	if len(cookies) == 0 {
		return fmt.Errorf("no cookies to save")
	}

	s.cookies = cookies
	s.logger.Info("session saved", zap.Int("cookie_count", len(cookies)))

	return nil
}

// GetCookies returns stored cookies
func (s *Session) GetCookies() []*proto.NetworkCookie {
	return s.cookies
}

// IsValid checks if the session is still valid
func (s *Session) IsValid() bool {
	if len(s.cookies) == 0 {
		return false
	}

	// Check if any critical cookies have expired
	now := time.Now()

	for _, cookie := range s.cookies {
		// Look for LinkedIn session cookies
		if cookie.Name == "li_at" || cookie.Name == "JSESSIONID" {
			if cookie.Expires > 0 {
				expiryTime := time.Unix(int64(cookie.Expires), 0)
				if now.After(expiryTime) {
					s.logger.Warn("session cookie expired",
						zap.String("cookie", cookie.Name),
						zap.Time("expired_at", expiryTime),
					)
					return false
				}
			}
		}
	}

	return true
}

// ToJSON serializes session data to JSON
func (s *Session) ToJSON() ([]byte, error) {
	data := SessionData{
		SessionID: s.sessionID,
		Cookies:   s.cookies,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	return json.Marshal(data)
}

// FromJSON deserializes session data from JSON
func (s *Session) FromJSON(data []byte) error {
	var sessionData SessionData

	if err := json.Unmarshal(data, &sessionData); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	s.sessionID = sessionData.SessionID
	s.cookies = sessionData.Cookies

	// Check if session has expired
	if time.Now().After(sessionData.ExpiresAt) {
		return fmt.Errorf("session has expired")
	}

	s.logger.Info("session loaded from JSON",
		zap.String("session_id", s.sessionID),
		zap.Int("cookie_count", len(s.cookies)),
	)

	return nil
}

// GetSessionID returns the session ID
func (s *Session) GetSessionID() string {
	return s.sessionID
}

// Clear clears the session data
func (s *Session) Clear() {
	s.cookies = nil
	s.logger.Info("session cleared")
}

// HasAuthCookie checks if session has authentication cookie
func (s *Session) HasAuthCookie() bool {
	for _, cookie := range s.cookies {
		if cookie.Name == "li_at" {
			return true
		}
	}
	return false
}

// GetCookieValue retrieves a specific cookie value
func (s *Session) GetCookieValue(name string) string {
	for _, cookie := range s.cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// UpdateCookies updates the stored cookies
func (s *Session) UpdateCookies(cookies []*proto.NetworkCookie) {
	s.cookies = cookies
	s.logger.Debug("cookies updated", zap.Int("count", len(cookies)))
}

// GetExpiryTime returns when the session expires
func (s *Session) GetExpiryTime() time.Time {
	var earliestExpiry time.Time

	for _, cookie := range s.cookies {
		if cookie.Expires > 0 {
			expiryTime := time.Unix(int64(cookie.Expires), 0)
			if earliestExpiry.IsZero() || expiryTime.Before(earliestExpiry) {
				earliestExpiry = expiryTime
			}
		}
	}

	return earliestExpiry
}
