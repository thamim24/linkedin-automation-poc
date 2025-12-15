package auth

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/log"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// CheckpointType represents different types of security challenges
type CheckpointType string

const (
	CheckpointNone         CheckpointType = "none"
	CheckpointCAPTCHA      CheckpointType = "captcha"
	Checkpoint2FA          CheckpointType = "2fa"
	CheckpointVerification CheckpointType = "verification"
	CheckpointRestricted   CheckpointType = "restricted"
	CheckpointUnknown      CheckpointType = "unknown"
)

// Checkpoint detects and handles LinkedIn security challenges
type Checkpoint struct {
	page   *rod.Page
	logger *zap.Logger
}

// NewCheckpoint creates a new checkpoint detector
func NewCheckpoint(page *rod.Page) *Checkpoint {
	return &Checkpoint{
		page:   page,
		logger: log.Module("checkpoint"),
	}
}

// Detect checks for any security challenges
func (c *Checkpoint) Detect() (CheckpointType, string) {
	// Check for CAPTCHA
	if c.hasCAPTCHA() {
		c.logger.Warn("CAPTCHA detected")
		return CheckpointCAPTCHA, "CAPTCHA challenge detected - manual intervention required"
	}

	// Check for 2FA
	if c.has2FA() {
		c.logger.Warn("2FA prompt detected")
		return Checkpoint2FA, "Two-factor authentication required"
	}

	// Check for verification challenge
	if c.hasVerificationChallenge() {
		c.logger.Warn("verification challenge detected")
		return CheckpointVerification, "Account verification required"
	}

	// Check for account restrictions
	if c.hasAccountRestriction() {
		c.logger.Error("account restriction detected")
		return CheckpointRestricted, "Account has been restricted"
	}

	// Check for unusual activity warning
	if c.hasUnusualActivity() {
		c.logger.Warn("unusual activity warning detected")
		return CheckpointVerification, "Unusual activity detected by LinkedIn"
	}

	return CheckpointNone, ""
}

// hasCAPTCHA checks for CAPTCHA elements that are actually visible
func (c *Checkpoint) hasCAPTCHA() bool {
	selectors := []string{
		"#captcha-internal",
		".g-recaptcha",
		"[data-test-id='captcha']",
		"iframe[src*='recaptcha']",
		"iframe[src*='captcha']",
	}

	for _, selector := range selectors {
		element, err := c.page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			// Check if element is actually visible
			visible, _ := element.Visible()
			if visible {
				return true
			}
		}
	}

	return false
}

// has2FA checks for two-factor authentication prompt that is visible
func (c *Checkpoint) has2FA() bool {
	selectors := []string{
		"input[name='pin']",
		"input[id='pin']",
		"[data-test-id='2fa-input']",
		".two-step-verification",
	}

	for _, selector := range selectors {
		element, err := c.page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			visible, _ := element.Visible()
			if visible {
				return true
			}
		}
	}

	return false
}

// hasVerificationChallenge checks for verification challenges that are visible
func (c *Checkpoint) hasVerificationChallenge() bool {
	selectors := []string{
		".challenge-dialog",
		"[data-test-id='verification-challenge']",
		".security-challenge",
		"#email-pin-challenge",
	}

	for _, selector := range selectors {
		element, err := c.page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			visible, _ := element.Visible()
			if visible {
				return true
			}
		}
	}

	// Check URL for verification paths
	url := c.page.MustInfo().URL
	verificationPaths := []string{
		"/checkpoint/challenge",
		"/checkpoint/lg/",
		"/challenge/",
	}

	for _, path := range verificationPaths {
		if contains(url, path) {
			return true
		}
	}

	return false
}

// hasAccountRestriction checks if account is restricted
func (c *Checkpoint) hasAccountRestriction() bool {
	selectors := []string{
		".account-restricted",
		"[data-test-id='account-restricted']",
		".restricted-account-notice",
	}

	for _, selector := range selectors {
		element, err := c.page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			text, _ := element.Text()
			c.logger.Error("restriction details", zap.String("message", text))
			return true
		}
	}

	// Check for restriction keywords in page text
	bodyText, err := c.page.Timeout(2 * time.Second).Element("body")
	if err == nil {
		text, _ := bodyText.Text()
		restrictionKeywords := []string{
			"account has been restricted",
			"temporarily restricted",
			"violated our terms",
			"suspended",
		}

		for _, keyword := range restrictionKeywords {
			if contains(text, keyword) {
				return true
			}
		}
	}

	return false
}

// hasUnusualActivity checks for unusual activity warnings
func (c *Checkpoint) hasUnusualActivity() bool {
	selectors := []string{
		".unusual-activity",
		"[data-test-id='unusual-activity']",
	}

	for _, selector := range selectors {
		_, err := c.page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			return true
		}
	}

	// Check page text for unusual activity messages
	bodyText, err := c.page.Timeout(2 * time.Second).Element("body")
	if err == nil {
		text, _ := bodyText.Text()
		if contains(text, "unusual activity") || contains(text, "verify your identity") {
			return true
		}
	}

	return false
}

// WaitAndDetect waits for potential challenges to appear
func (c *Checkpoint) WaitAndDetect(duration time.Duration) (CheckpointType, string) {
	time.Sleep(duration)
	return c.Detect()
}

// IsBlocked checks if we're blocked from proceeding
func (c *Checkpoint) IsBlocked() bool {
	checkpointType, _ := c.Detect()
	return checkpointType != CheckpointNone
}

// GetErrorMessage returns a user-friendly error message for the checkpoint
func (c *Checkpoint) GetErrorMessage(checkpointType CheckpointType) error {
	switch checkpointType {
	case CheckpointCAPTCHA:
		return fmt.Errorf("CAPTCHA challenge detected. Please solve it manually and restart")
	case Checkpoint2FA:
		return fmt.Errorf("two-factor authentication required. Please complete 2FA manually")
	case CheckpointVerification:
		return fmt.Errorf("LinkedIn requires account verification. Please verify manually")
	case CheckpointRestricted:
		return fmt.Errorf("account has been restricted or suspended by LinkedIn")
	case CheckpointUnknown:
		return fmt.Errorf("unknown security challenge detected")
	default:
		return nil
	}
}

// LogCheckpoint logs checkpoint information
func (c *Checkpoint) LogCheckpoint(checkpointType CheckpointType, message string) {
	c.logger.Warn("checkpoint detected",
		zap.String("type", string(checkpointType)),
		zap.String("message", message),
	)
}
