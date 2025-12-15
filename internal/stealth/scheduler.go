package stealth

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/log"

	"go.uber.org/zap"
)

// Scheduler manages when actions should be performed
type Scheduler struct {
	config        config.ScheduleConfig
	actionCount   int
	lastBreakTime time.Time
	logger        *zap.Logger
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg config.ScheduleConfig) *Scheduler {
	return &Scheduler{
		config:        cfg,
		actionCount:   0,
		lastBreakTime: time.Now(),
		logger:        log.Module("scheduler"),
	}
}

// ShouldRun checks if the bot should be running now
func (s *Scheduler) ShouldRun() (bool, string) {
	now := time.Now()

	// Check if it's a valid weekday
	if !s.isValidWeekday(now) {
		return false, fmt.Sprintf("not a configured weekday: %s", now.Weekday())
	}

	// Check business hours
	if s.config.BusinessHoursOnly {
		if !s.isBusinessHours(now) {
			return false, fmt.Sprintf("outside business hours (%d:00-%d:00)",
				s.config.StartHour, s.config.EndHour)
		}
	}

	// Check if we need a break
	if s.needsBreak() {
		return false, fmt.Sprintf("break required after %d actions", s.config.BreakAfter)
	}

	return true, "ok to run"
}

// isValidWeekday checks if today is a configured weekday
func (s *Scheduler) isValidWeekday(t time.Time) bool {
	if len(s.config.Weekdays) == 0 {
		return true // No restriction
	}

	currentDay := t.Weekday().String()

	for _, day := range s.config.Weekdays {
		if day == currentDay {
			return true
		}
	}

	return false
}

// isBusinessHours checks if current time is within business hours
func (s *Scheduler) isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	return hour >= s.config.StartHour && hour < s.config.EndHour
}

// needsBreak checks if a break is needed
func (s *Scheduler) needsBreak() bool {
	if s.config.BreakAfter <= 0 {
		return false
	}

	return s.actionCount >= s.config.BreakAfter
}

// TakeBreak pauses execution for the configured break duration
func (s *Scheduler) TakeBreak() {
	s.logger.Info("taking break",
		zap.Int("actions_completed", s.actionCount),
		zap.Duration("duration", s.config.BreakDuration),
	)

	time.Sleep(s.config.BreakDuration)

	s.actionCount = 0
	s.lastBreakTime = time.Now()

	s.logger.Info("break completed")
}

// RecordAction records that an action was performed
func (s *Scheduler) RecordAction() {
	s.actionCount++
}

// GetActionCount returns the current action count
func (s *Scheduler) GetActionCount() int {
	return s.actionCount
}

// ResetActionCount resets the action counter
func (s *Scheduler) ResetActionCount() {
	s.actionCount = 0
}

// WaitUntilBusinessHours waits until business hours start
func (s *Scheduler) WaitUntilBusinessHours() {
	if !s.config.BusinessHoursOnly {
		return
	}

	now := time.Now()

	for {
		if s.isBusinessHours(now) && s.isValidWeekday(now) {
			return
		}

		nextStart := s.calculateNextStartTime(now)
		waitDuration := nextStart.Sub(now)

		s.logger.Info("waiting for business hours",
			zap.Time("next_start", nextStart),
			zap.Duration("wait_duration", waitDuration),
		)

		// Check every minute
		if waitDuration > time.Minute {
			time.Sleep(time.Minute)
			now = time.Now()
		} else {
			time.Sleep(waitDuration)
			return
		}
	}
}

// calculateNextStartTime calculates the next valid start time
func (s *Scheduler) calculateNextStartTime(from time.Time) time.Time {
	next := from

	// If we're past business hours today, move to next day
	if next.Hour() >= s.config.EndHour {
		next = next.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(),
			s.config.StartHour, 0, 0, 0, next.Location())
	} else if next.Hour() < s.config.StartHour {
		// Before business hours today
		next = time.Date(next.Year(), next.Month(), next.Day(),
			s.config.StartHour, 0, 0, 0, next.Location())
	}

	// Keep advancing until we hit a valid weekday
	for !s.isValidWeekday(next) {
		next = next.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(),
			s.config.StartHour, 0, 0, 0, next.Location())
	}

	return next
}

// GetTimeUntilNextRun returns how long until the next valid run time
func (s *Scheduler) GetTimeUntilNextRun() time.Duration {
	now := time.Now()

	if canRun, _ := s.ShouldRun(); canRun {
		return 0
	}

	nextStart := s.calculateNextStartTime(now)
	return nextStart.Sub(now)
}

// IsEndOfDay checks if we're near the end of business hours
func (s *Scheduler) IsEndOfDay() bool {
	if !s.config.BusinessHoursOnly {
		return false
	}

	now := time.Now()
	hour := now.Hour()

	// Within 1 hour of end time
	return hour >= s.config.EndHour-1
}

// GetRemainingTime returns remaining time in current business hours
func (s *Scheduler) GetRemainingTime() time.Duration {
	if !s.config.BusinessHoursOnly {
		return 24 * time.Hour // Unlimited
	}

	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(),
		s.config.EndHour, 0, 0, 0, now.Location())

	if now.After(endTime) {
		return 0
	}

	return endTime.Sub(now)
}
