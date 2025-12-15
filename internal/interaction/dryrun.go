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

// DryRun handles dry-run profile visits (no actual actions)
type DryRun struct {
	ctx    *browser.Context
	mouse  *stealth.Mouse
	scroll *stealth.Scroll
	timing *stealth.Timing
	logger *zap.Logger
}

// NewDryRun creates a new dry-run handler
func NewDryRun(ctx *browser.Context, mouse *stealth.Mouse, scroll *stealth.Scroll, timing *stealth.Timing) *DryRun {
	return &DryRun{
		ctx:    ctx,
		mouse:  mouse,
		scroll: scroll,
		timing: timing,
		logger: log.Module("dryrun"),
	}
}

// VisitProfile performs a human-like profile visit without taking action
func (d *DryRun) VisitProfile(profileURL string) error {
	d.logger.Info("DRY_RUN visiting_profile", zap.String("url", profileURL))

	page := d.ctx.Page()

	// Navigate to profile
	if err := d.ctx.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for profile main section to load
	d.logger.Debug("DRY_RUN waiting for profile to load")
	_, err := page.Timeout(10 * time.Second).Element("main")
	if err != nil {
		d.logger.Warn("DRY_RUN profile main section not found", zap.Error(err))
		// Continue anyway - page might have loaded differently
	}

	time.Sleep(d.timing.PageLoadDelay())

	// Human-like scrolling behavior
	d.logger.Info("DRY_RUN scrolling_profile")

	// Small scroll down
	if err := d.scroll.ScrollDown(); err != nil {
		d.logger.Warn("DRY_RUN scroll failed", zap.Error(err))
	}
	time.Sleep(time.Duration(1000+d.timing.ShortPause().Milliseconds()) * time.Millisecond)

	// Slight scroll back up (reading behavior)
	if err := d.scroll.ScrollUp(); err != nil {
		d.logger.Warn("DRY_RUN scroll up failed", zap.Error(err))
	}
	time.Sleep(d.timing.ShortPause())

	// Move mouse naturally across page
	d.logger.Debug("DRY_RUN moving mouse naturally")
	if err := d.moveMouseNaturally(page); err != nil {
		d.logger.Warn("DRY_RUN mouse movement failed", zap.Error(err))
	}

	// Find and hover over Connect button (WITHOUT CLICKING)
	connectButton := d.findConnectButton(page)
	if connectButton != nil {
		d.logger.Info("DRY_RUN hovering_button", zap.String("button", "Connect"))

		// Get button position
		shape, err := connectButton.Shape()
		if err == nil && len(shape.Quads) > 0 && len(shape.Quads[0]) >= 4 {
			centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
			centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2

			// Move mouse to Connect button
			if err := d.mouse.MoveTo(centerX, centerY); err == nil {
				// Hold hover for 1-2 seconds
				time.Sleep(time.Duration(1000+d.timing.ShortPause().Milliseconds()) * time.Millisecond)

				d.logger.Warn("DRY_RUN would_click",
					zap.String("button", "Connect"),
					zap.String("action", "SKIPPED - dry run mode"),
					zap.String("profile", profileURL))
			}
		}
	} else {
		d.logger.Debug("DRY_RUN connect button not found (may already be connected)")
	}

	d.logger.Info("DRY_RUN profile visit complete", zap.String("url", profileURL))
	return nil
}

// moveMouseNaturally simulates reading behavior with mouse movement
func (d *DryRun) moveMouseNaturally(page *rod.Page) error {
	// Get page dimensions
	viewport := page.MustGetWindow()

	// Convert pointers to values and calculate positions
	width := float64(*viewport.Width)
	height := float64(*viewport.Height)

	// Move to a few random points on page (simulating reading)
	points := []struct{ x, y float64 }{
		{width * 0.3, height * 0.3},
		{width * 0.5, height * 0.4},
		{width * 0.4, height * 0.6},
	}

	for _, point := range points {
		if err := d.mouse.MoveTo(point.x, point.y); err != nil {
			return err
		}
		time.Sleep(d.timing.ShortPause())
	}

	return nil
}

// findConnectButton locates the Connect button (if exists)
func (d *DryRun) findConnectButton(page *rod.Page) *rod.Element {
	selectors := []string{
		"button[aria-label*='Connect']",
		"button[aria-label*='connect']",
		"button:has-text('Connect')",
		".pvs-profile-actions button",
	}

	for _, selector := range selectors {
		btn, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil {
			return btn
		}
	}

	return nil
}
