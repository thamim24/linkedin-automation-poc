package interaction

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// Connect handles sending connection requests
type Connect struct {
	ctx    *browser.Context
	mouse  *stealth.Mouse
	scroll *stealth.Scroll
	typing *stealth.Typing
	timing *stealth.Timing
	logger *zap.Logger
}

// NewConnect creates a new connection request handler
func NewConnect(ctx *browser.Context, mouse *stealth.Mouse, scroll *stealth.Scroll, typing *stealth.Typing, timing *stealth.Timing) *Connect {
	return &Connect{
		ctx:    ctx,
		mouse:  mouse,
		scroll: scroll,
		typing: typing,
		timing: timing,
		logger: log.Module("connect"),
	}
}

// SendRequest sends a connection request to a profile
func (c *Connect) SendRequest(profileURL string, note string) error {
	c.logger.Info("sending connection request", zap.String("profile", profileURL))

	page := c.ctx.Page()

	// Navigate to profile
	if err := c.ctx.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	// Wait for page to load
	time.Sleep(c.timing.PageLoadDelay())

	// Find Connect button
	connectButton, err := c.findConnectButton(page)
	if err != nil {
		return fmt.Errorf("failed to find connect button: %w", err)
	}

	// Scroll button into view
	if err := c.scroll.ScrollToElement(connectButton); err != nil {
		c.logger.Warn("failed to scroll to connect button", zap.Error(err))
	}

	time.Sleep(c.timing.ShortPause())

	// Hover over button before clicking
	shape, err := connectButton.Shape()
	if err == nil && len(shape.Quads) > 0 && len(shape.Quads[0]) >= 4 {
		centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
		centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2

		if err := c.mouse.Hover(centerX, centerY); err != nil {
			c.logger.Warn("failed to hover over button", zap.Error(err))
		}
	}

	time.Sleep(c.timing.ShortPause())

	// Click Connect button
	if err := connectButton.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click connect button: %w", err)
	}

	c.logger.Debug("connect button clicked")

	// Wait for modal to appear
	time.Sleep(2 * time.Second)

	// Handle connection modal
	if note != "" {
		if err := c.addNote(page, note); err != nil {
			c.logger.Warn("failed to add note, sending without note", zap.Error(err))
		}
	}

	// Click Send button in modal
	if err := c.clickSendButton(page); err != nil {
		return fmt.Errorf("failed to send connection request: %w", err)
	}

	c.logger.Info("connection request sent successfully")

	return nil
}

// SendRequestFromCard sends connection request directly from search result card
func (c *Connect) SendRequestFromCard(card *rod.Element, note string) error {
	c.logger.Debug("sending connection request from card")

	// Find Connect button within card
	connectButton, err := card.Timeout(2 * time.Second).Element("button[aria-label*='Connect']")
	if err != nil {
		return fmt.Errorf("failed to find connect button in card: %w", err)
	}

	// Scroll to button
	if err := c.scroll.ScrollToElement(connectButton); err != nil {
		c.logger.Warn("failed to scroll to button", zap.Error(err))
	}

	time.Sleep(c.timing.ShortPause())

	// Click button
	if err := connectButton.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click connect button: %w", err)
	}

	// Wait for modal
	time.Sleep(2 * time.Second)

	page := c.ctx.Page()

	// Add note if provided
	if note != "" {
		if err := c.addNote(page, note); err != nil {
			c.logger.Warn("failed to add note", zap.Error(err))
		}
	}

	// Send request
	if err := c.clickSendButton(page); err != nil {
		return fmt.Errorf("failed to send connection request: %w", err)
	}

	c.logger.Debug("connection request sent from card")

	return nil
}

// findConnectButton finds the Connect button on profile page
func (c *Connect) findConnectButton(page *rod.Page) (*rod.Element, error) {
	// Try multiple selectors
	selectors := []string{
		"button[aria-label*='Connect']",
		"button.pvs-profile-actions__action:has-text('Connect')",
		"button:has-text('Connect')",
	}

	for _, selector := range selectors {
		btn, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil {
			return btn, nil
		}
	}

	return nil, fmt.Errorf("connect button not found")
}

// addNote adds a personalized note to connection request
func (c *Connect) addNote(page *rod.Page, note string) error {
	c.logger.Debug("adding note to connection request")

	// Click "Add a note" button if present
	addNoteButton, err := page.Timeout(2 * time.Second).Element("button[aria-label='Add a note']")
	if err == nil {
		if err := addNoteButton.Click("left", 1); err != nil {
			return fmt.Errorf("failed to click add note button: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	// Find note textarea
	noteTextarea, err := page.Timeout(3 * time.Second).Element("textarea[name='message']")
	if err != nil {
		return fmt.Errorf("failed to find note textarea: %w", err)
	}

	// Type note with human-like behavior
	if err := c.typing.TypeIntoElement(noteTextarea, note); err != nil {
		return fmt.Errorf("failed to type note: %w", err)
	}

	c.logger.Debug("note added successfully")

	return nil
}

// clickSendButton clicks the Send button in connection modal
func (c *Connect) clickSendButton(page *rod.Page) error {
	sendButton, err := page.Timeout(3 * time.Second).Element("button[aria-label='Send now']")
	if err != nil {
		// Try alternative selector
		sendButton, err = page.Timeout(2 * time.Second).Element("button[aria-label='Send']")
		if err != nil {
			return fmt.Errorf("failed to find send button: %w", err)
		}
	}

	// Small delay before clicking
	time.Sleep(c.timing.ShortPause())

	if err := sendButton.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click send button: %w", err)
	}

	// Wait for confirmation
	time.Sleep(2 * time.Second)

	return nil
}

// IsAlreadyConnected checks if already connected to profile
func (c *Connect) IsAlreadyConnected(page *rod.Page) bool {
	// Look for Message button (indicates existing connection)
	_, err := page.Timeout(2 * time.Second).Element("button[aria-label*='Message']")
	return err == nil
}

// IsPendingConnection checks if connection request is pending
func (c *Connect) IsPendingConnection(page *rod.Page) bool {
	// Look for Pending button
	_, err := page.Timeout(2 * time.Second).Element("button[aria-label*='Pending']")
	return err == nil
}

// CanConnect checks if connection request can be sent
func (c *Connect) CanConnect(page *rod.Page) bool {
	if c.IsAlreadyConnected(page) {
		c.logger.Debug("already connected")
		return false
	}

	if c.IsPendingConnection(page) {
		c.logger.Debug("connection pending")
		return false
	}

	// Check if Connect button exists
	_, err := c.findConnectButton(page)
	return err == nil
}
