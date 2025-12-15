package interaction

import (
	"fmt"
	"sync"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/log"

	"go.uber.org/zap"
)

// Limits manages rate limiting and daily caps
type Limits struct {
	config           config.LimitsConfig
	connectionCount  int
	messageCount     int
	searchCount      int
	lastConnection   time.Time
	lastMessage      time.Time
	consecutiveCount int
	lastReset        time.Time
	mu               sync.Mutex
	logger           *zap.Logger
}

// NewLimits creates a new limits manager
func NewLimits(cfg config.LimitsConfig) *Limits {
	return &Limits{
		config:          cfg,
		connectionCount: 0,
		messageCount:    0,
		searchCount:     0,
		lastReset:       time.Now(),
		logger:          log.Module("limits"),
	}
}

// CanSendConnection checks if a connection request can be sent
func (l *Limits) CanSendConnection() (bool, string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check daily limit
	if l.connectionCount >= l.config.DailyConnections {
		return false, fmt.Sprintf("daily connection limit reached (%d/%d)",
			l.connectionCount, l.config.DailyConnections)
	}

	// Check cooldown
	if !l.lastConnection.IsZero() {
		elapsed := time.Since(l.lastConnection)
		if elapsed < l.config.ConnectionCooldown {
			remaining := l.config.ConnectionCooldown - elapsed
			return false, fmt.Sprintf("connection cooldown active (wait %s)",
				remaining.Round(time.Second))
		}
	}

	// Check consecutive actions
	if l.consecutiveCount >= l.config.MaxConsecutiveActions {
		return false, fmt.Sprintf("max consecutive actions reached (%d), take a break",
			l.config.MaxConsecutiveActions)
	}

	return true, "ok"
}

// CanSendMessage checks if a message can be sent
func (l *Limits) CanSendMessage() (bool, string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check daily limit
	if l.messageCount >= l.config.DailyMessages {
		return false, fmt.Sprintf("daily message limit reached (%d/%d)",
			l.messageCount, l.config.DailyMessages)
	}

	// Check cooldown
	if !l.lastMessage.IsZero() {
		elapsed := time.Since(l.lastMessage)
		if elapsed < l.config.MessageCooldown {
			remaining := l.config.MessageCooldown - elapsed
			return false, fmt.Sprintf("message cooldown active (wait %s)",
				remaining.Round(time.Second))
		}
	}

	return true, "ok"
}

// CanSearch checks if a search can be performed
func (l *Limits) CanSearch() (bool, string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check daily limit
	if l.searchCount >= l.config.DailySearches {
		return false, fmt.Sprintf("daily search limit reached (%d/%d)",
			l.searchCount, l.config.DailySearches)
	}

	return true, "ok"
}

// RecordConnection records that a connection was sent
func (l *Limits) RecordConnection() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.connectionCount++
	l.lastConnection = time.Now()
	l.consecutiveCount++

	l.logger.Info("connection recorded",
		zap.Int("count", l.connectionCount),
		zap.Int("daily_limit", l.config.DailyConnections),
		zap.Int("consecutive", l.consecutiveCount),
	)
}

// RecordMessage records that a message was sent
func (l *Limits) RecordMessage() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.messageCount++
	l.lastMessage = time.Now()

	l.logger.Info("message recorded",
		zap.Int("count", l.messageCount),
		zap.Int("daily_limit", l.config.DailyMessages),
	)
}

// RecordSearch records that a search was performed
func (l *Limits) RecordSearch() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.searchCount++

	l.logger.Debug("search recorded",
		zap.Int("count", l.searchCount),
		zap.Int("daily_limit", l.config.DailySearches),
	)
}

// ResetConsecutiveCount resets the consecutive action counter
func (l *Limits) ResetConsecutiveCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.consecutiveCount = 0
	l.logger.Debug("consecutive counter reset")
}

// ResetDaily resets daily counters (should be called at day rollover)
func (l *Limits) ResetDaily() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.connectionCount = 0
	l.messageCount = 0
	l.searchCount = 0
	l.consecutiveCount = 0
	l.lastReset = time.Now()

	l.logger.Info("daily limits reset")
}

// ShouldResetDaily checks if daily reset is needed
func (l *Limits) ShouldResetDaily() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	lastResetDay := l.lastReset.Day()
	currentDay := now.Day()

	return lastResetDay != currentDay
}

// GetConnectionCount returns current connection count
func (l *Limits) GetConnectionCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.connectionCount
}

// GetMessageCount returns current message count
func (l *Limits) GetMessageCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.messageCount
}

// GetSearchCount returns current search count
func (l *Limits) GetSearchCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.searchCount
}

// GetRemainingConnections returns remaining connections for today
func (l *Limits) GetRemainingConnections() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	remaining := l.config.DailyConnections - l.connectionCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingMessages returns remaining messages for today
func (l *Limits) GetRemainingMessages() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	remaining := l.config.DailyMessages - l.messageCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// WaitForConnectionCooldown waits for connection cooldown to expire
func (l *Limits) WaitForConnectionCooldown() {
	l.mu.Lock()
	lastConn := l.lastConnection
	cooldown := l.config.ConnectionCooldown
	l.mu.Unlock()

	if lastConn.IsZero() {
		return
	}

	elapsed := time.Since(lastConn)
	if elapsed < cooldown {
		waitTime := cooldown - elapsed
		l.logger.Info("waiting for connection cooldown",
			zap.Duration("wait_time", waitTime))
		time.Sleep(waitTime)
	}
}

// WaitForMessageCooldown waits for message cooldown to expire
func (l *Limits) WaitForMessageCooldown() {
	l.mu.Lock()
	lastMsg := l.lastMessage
	cooldown := l.config.MessageCooldown
	l.mu.Unlock()

	if lastMsg.IsZero() {
		return
	}

	elapsed := time.Since(lastMsg)
	if elapsed < cooldown {
		waitTime := cooldown - elapsed
		l.logger.Info("waiting for message cooldown",
			zap.Duration("wait_time", waitTime))
		time.Sleep(waitTime)
	}
}

// GetStats returns current limit statistics
func (l *Limits) GetStats() map[string]interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	return map[string]interface{}{
		"connections": map[string]int{
			"sent":      l.connectionCount,
			"limit":     l.config.DailyConnections,
			"remaining": l.config.DailyConnections - l.connectionCount,
		},
		"messages": map[string]int{
			"sent":      l.messageCount,
			"limit":     l.config.DailyMessages,
			"remaining": l.config.DailyMessages - l.messageCount,
		},
		"searches": map[string]int{
			"performed": l.searchCount,
			"limit":     l.config.DailySearches,
			"remaining": l.config.DailySearches - l.searchCount,
		},
		"consecutive_actions": l.consecutiveCount,
		"last_reset":          l.lastReset,
	}
}
