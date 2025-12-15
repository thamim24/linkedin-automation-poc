package interaction

import (
	"fmt"
	"strings"
	"time"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// Message handles sending messages to connections
type Message struct {
	ctx    *browser.Context
	mouse  *stealth.Mouse
	scroll *stealth.Scroll
	typing *stealth.Typing
	timing *stealth.Timing
	logger *zap.Logger
}

// NewMessage creates a new message handler
func NewMessage(ctx *browser.Context, mouse *stealth.Mouse, scroll *stealth.Scroll, typing *stealth.Typing, timing *stealth.Timing) *Message {
	return &Message{
		ctx:    ctx,
		mouse:  mouse,
		scroll: scroll,
		typing: typing,
		timing: timing,
		logger: log.Module("message"),
	}
}

// Send sends a message to a connection
func (m *Message) Send(profileURL string, messageText string) error {
	m.logger.Info("sending message", zap.String("profile", profileURL))

	page := m.ctx.Page()

	// Navigate to profile
	if err := m.ctx.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	// Wait for page to load
	time.Sleep(m.timing.PageLoadDelay())

	// Find Message button
	messageButton, err := m.findMessageButton(page)
	if err != nil {
		return fmt.Errorf("failed to find message button (not connected?): %w", err)
	}

	// Scroll button into view
	if err := m.scroll.ScrollToElement(messageButton); err != nil {
		m.logger.Warn("failed to scroll to message button", zap.Error(err))
	}

	time.Sleep(m.timing.ShortPause())

	// Click Message button
	if err := messageButton.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click message button: %w", err)
	}

	m.logger.Debug("message button clicked")

	// Wait for messaging window to open
	time.Sleep(2 * time.Second)

	// Find message input box
	messageBox, err := m.findMessageBox(page)
	if err != nil {
		return fmt.Errorf("failed to find message input: %w", err)
	}

	// Type message
	if err := m.typing.TypeIntoElement(messageBox, messageText); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	m.logger.Debug("message typed")

	// Wait before sending (reading time)
	time.Sleep(m.timing.ThinkingDelay())

	// Send message
	if err := m.clickSendButton(page); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	m.logger.Info("message sent successfully")

	return nil
}

// SendFollowUp sends a follow-up message in an existing conversation
func (m *Message) SendFollowUp(conversationURL string, messageText string) error {
	m.logger.Info("sending follow-up message")

	page := m.ctx.Page()

	// Navigate to conversation
	if err := m.ctx.Navigate(conversationURL); err != nil {
		return fmt.Errorf("failed to navigate to conversation: %w", err)
	}

	time.Sleep(m.timing.PageLoadDelay())

	// Find message input
	messageBox, err := m.findMessageBox(page)
	if err != nil {
		return fmt.Errorf("failed to find message input: %w", err)
	}

	// Type message
	if err := m.typing.TypeIntoElement(messageBox, messageText); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	time.Sleep(m.timing.ShortPause())

	// Send message
	if err := m.clickSendButton(page); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	m.logger.Info("follow-up message sent")

	return nil
}

// findMessageButton finds the Message button on profile page
func (m *Message) findMessageButton(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"button[aria-label*='Message']",
		"a[aria-label*='Message']",
		"button.pvs-profile-actions__action:has-text('Message')",
	}

	for _, selector := range selectors {
		btn, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil {
			return btn, nil
		}
	}

	return nil, fmt.Errorf("message button not found")
}

// findMessageBox finds the message input textarea
func (m *Message) findMessageBox(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"div[role='textbox'][contenteditable='true']",
		"textarea.msg-form__contenteditable",
		".msg-form__msg-content-container .msg-form__contenteditable",
	}

	for _, selector := range selectors {
		box, err := page.Timeout(5 * time.Second).Element(selector)
		if err == nil {
			return box, nil
		}
	}

	return nil, fmt.Errorf("message input not found")
}

// clickSendButton clicks the Send button to send message
func (m *Message) clickSendButton(page *rod.Page) error {
	selectors := []string{
		"button[type='submit'][aria-label*='Send']",
		"button.msg-form__send-button",
		"button[aria-label='Send']",
	}

	for _, selector := range selectors {
		sendBtn, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil {
			time.Sleep(m.timing.ShortPause())

			if err := sendBtn.Click("left", 1); err != nil {
				continue
			}

			// Wait for message to send
			time.Sleep(1 * time.Second)
			return nil
		}
	}

	return fmt.Errorf("send button not found")
}

// IsMessagingAvailable checks if can send messages to this profile
func (m *Message) IsMessagingAvailable(page *rod.Page) bool {
	_, err := m.findMessageButton(page)
	return err == nil
}

// RenderMessageTemplate replaces template variables in message
func (m *Message) RenderMessageTemplate(template string, variables map[string]string) string {
	message := template

	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		message = strings.ReplaceAll(message, placeholder, value)
	}

	return message
}
