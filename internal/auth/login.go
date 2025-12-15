package auth

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// Login handles LinkedIn authentication
type Login struct {
	ctx    *browser.Context
	email  string
	pass   string
	logger *zap.Logger
}

// NewLogin creates a new login handler
func NewLogin(ctx *browser.Context, email, password string) *Login {
	return &Login{
		ctx:    ctx,
		email:  email,
		pass:   password,
		logger: log.Module("auth"),
	}
}

// Execute performs the login process
func (l *Login) Execute() error {
	l.logger.Info("starting login process")

	page := l.ctx.Page()

	// Navigate to LinkedIn login page
	if err := l.ctx.Navigate("https://www.linkedin.com/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for page to stabilize
	time.Sleep(2 * time.Second)

	// Detect if already logged in
	if l.isLoggedIn(page) {
		l.logger.Info("already logged in")
		return nil
	}

	// Create stealth controllers
	typing := stealth.NewTyping(page, config.TypingConfig{
		MinKeystrokeDelay: 50 * time.Millisecond,
		MaxKeystrokeDelay: 200 * time.Millisecond,
		ErrorRate:         0.02,
		WordPauseMin:      100 * time.Millisecond,
		WordPauseMax:      300 * time.Millisecond,
	})

	mouse := stealth.NewMouse(page, config.MouseConfig{
		BezierSteps:     100,
		MinSteps:        80,
		MaxSteps:        150,
		Overshoot:       true,
		OvershootSpread: 10,
		JitterRadius:    3,
	})

	// Find email input
	l.logger.Debug("looking for email input")
	emailInput, err := page.Element("#username")
	if err != nil {
		return fmt.Errorf("failed to find email input: %w", err)
	}

	// Get element position for mouse movement
	shape, err := emailInput.Shape()
	if err != nil {
		return fmt.Errorf("failed to get email input position: %w", err)
	}

	if len(shape.Quads) > 0 && len(shape.Quads[0]) >= 4 {
		// Calculate center from quad points
		centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
		centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2
		if err := mouse.MoveTo(centerX, centerY); err != nil {
			return fmt.Errorf("failed to move mouse to email field: %w", err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Type email
	l.logger.Debug("entering email")
	if err := typing.TypeIntoElement(emailInput, l.email); err != nil {
		return fmt.Errorf("failed to enter email: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Find password input
	l.logger.Debug("looking for password input")
	passwordInput, err := page.Element("#password")
	if err != nil {
		return fmt.Errorf("failed to find password input: %w", err)
	}

	// Get password field position and click it
	shape, err = passwordInput.Shape()
	if err != nil {
		return fmt.Errorf("failed to get password input position: %w", err)
	}

	if len(shape.Quads) > 0 && len(shape.Quads[0]) >= 4 {
		// Move mouse to password field and click to focus
		centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
		centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2
		if err := mouse.MoveTo(centerX, centerY); err != nil {
			return fmt.Errorf("failed to move mouse to password field: %w", err)
		}
		time.Sleep(300 * time.Millisecond)

		// Click to focus the password field
		if err := passwordInput.Click("left", 1); err != nil {
			return fmt.Errorf("failed to click password field: %w", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Type password into the focused element
	l.logger.Debug("entering password")
	if err := typing.TypeIntoElement(passwordInput, l.pass); err != nil {
		return fmt.Errorf("failed to enter password: %w", err)
	}

	time.Sleep(800 * time.Millisecond)

	// Find and click submit button
	l.logger.Debug("looking for submit button")
	submitButton, err := page.Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("failed to find submit button: %w", err)
	}

	// Get submit button position
	shape, err = submitButton.Shape()
	if err != nil {
		return fmt.Errorf("failed to get submit button position: %w", err)
	}

	if len(shape.Quads) > 0 && len(shape.Quads[0]) >= 4 {
		// Move mouse to submit button and click
		centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
		centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2
		if err := mouse.ClickAt(centerX, centerY); err != nil {
			return fmt.Errorf("failed to click submit button: %w", err)
		}
	}

	l.logger.Info("login form submitted, waiting for response")

	// Wait for navigation or error
	time.Sleep(5 * time.Second)

	// Check for errors
	if err := l.detectErrors(page); err != nil {
		return err
	}

	// Verify login success
	if !l.isLoggedIn(page) {
		return fmt.Errorf("login appears to have failed")
	}

	l.logger.Info("login successful")
	return nil
}

// isLoggedIn checks if user is logged in
func (l *Login) isLoggedIn(page *rod.Page) bool {
	// Check for feed URL or profile icon
	currentURL := page.MustInfo().URL

	if len(currentURL) > 0 && (contains(currentURL, "/feed") ||
		contains(currentURL, "/mynetwork") ||
		contains(currentURL, "/jobs")) {
		return true
	}

	// Check for navigation elements that only appear when logged in
	_, err := page.Timeout(2 * time.Second).Element("nav.global-nav")
	return err == nil
}

// detectErrors checks for login errors
func (l *Login) detectErrors(page *rod.Page) error {
	// Check for error messages
	errorElement, err := page.Timeout(2 * time.Second).Element(".login-form__error")
	if err == nil {
		errorText, _ := errorElement.Text()
		return fmt.Errorf("login error: %s", errorText)
	}

	// Check for CAPTCHA
	_, err = page.Timeout(2 * time.Second).Element("#captcha-internal")
	if err == nil {
		return fmt.Errorf("CAPTCHA detected - manual intervention required")
	}

	// Check for 2FA prompt
	_, err = page.Timeout(2 * time.Second).Element("input[name='pin']")
	if err == nil {
		return fmt.Errorf("2FA required - manual intervention needed")
	}

	// Check for verification challenge
	_, err = page.Timeout(2 * time.Second).Element(".challenge-dialog")
	if err == nil {
		return fmt.Errorf("verification challenge detected")
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring is a helper for string contains
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
