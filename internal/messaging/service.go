package messaging

import (
	"strings"
	"time"

	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"go.uber.org/zap"
)

// Service handles message operations
type Service struct {
	typing *stealth.Typing
	logger *zap.Logger
}

// NewService creates a new messaging service
func NewService(typing *stealth.Typing) *Service {
	return &Service{
		typing: typing,
		logger: log.Module("messaging"),
	}
}

// RenderTemplate replaces template variables with actual values
func RenderTemplate(t MessageTemplate, vars map[string]string) string {
	msg := t.Body
	for k, v := range vars {
		// Use fallback for empty values
		value := v
		if strings.TrimSpace(value) == "" && k == "first_name" {
			value = "there"
		}
		msg = strings.ReplaceAll(msg, "{{"+k+"}}", value)
	}
	return strings.TrimSpace(msg)
}

// DryRunSendMessage demonstrates message sending without actual execution
func (s *Service) DryRunSendMessage(profileURL, firstName string) MessageRecord {
	s.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	s.logger.Info("💬 DRY-RUN MESSAGING START")
	s.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Log connection state assumption for dry-run
	s.logger.Info("DRY_RUN connection_status",
		zap.String("status", "unknown (dry-run assumption)"),
		zap.String("note", "Real flow requires connection acceptance detection"))

	// Select template
	template := FollowUpTemplate
	s.logger.Info("DRY_RUN message_template", zap.String("name", template.Name))

	// Render template with variables
	rendered := RenderTemplate(template, map[string]string{
		"first_name": firstName,
	})

	s.logger.Info("DRY_RUN resolved_message",
		zap.String("content", rendered))

	// Simulate typing characteristics
	charCount := len(rendered)
	avgDelay := 94 // ms per character (typical human typing speed)
	typos := s.calculateTypos(rendered)
	corrections := typos

	s.logger.Info("DRY_RUN typing_simulated",
		zap.Int("chars", charCount),
		zap.Int("avg_delay_ms", avgDelay),
		zap.Int("typos", typos),
		zap.Int("corrections", corrections))

	// Brief delay to simulate thinking
	time.Sleep(500 * time.Millisecond)

	// Log the SKIPPED action
	s.logger.Warn("DRY_RUN would_send_message",
		zap.String("action", "SKIPPED - dry run mode"),
		zap.String("profile", profileURL))

	// Create message record
	record := NewMessageRecord(profileURL, template.Name, MessageSkipped)

	s.logger.Info("DRY_RUN message_state_saved",
		zap.String("status", string(record.Status)),
		zap.String("profile", profileURL))

	s.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	s.logger.Info("✅ DRY-RUN MESSAGING COMPLETE")
	s.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return record
}

// calculateTypos estimates realistic typo count for a message
func (s *Service) calculateTypos(message string) int {
	// Typical error rate: ~1-2% of characters
	length := len(message)
	if length < 30 {
		return 0
	}
	if length < 60 {
		return 1
	}
	return 2
}
