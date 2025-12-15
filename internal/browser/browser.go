package browser

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
)

// Browser manages the browser lifecycle
type Browser struct {
	instance *rod.Browser
	config   config.BrowserConfig
	logger   *zap.Logger
}

// New creates a new browser instance
func New(cfg config.BrowserConfig) (*Browser, error) {
	logger := log.Module("browser")

	b := &Browser{
		config: cfg,
		logger: logger,
	}

	if err := b.launch(); err != nil {
		return nil, err
	}

	return b, nil
}

// launch starts the browser
func (b *Browser) launch() error {
	b.logger.Info("launching browser",
		zap.Bool("headless", b.config.Headless),
		zap.Int("width", b.config.WindowWidth),
		zap.Int("height", b.config.WindowHeight),
	)

	l := launcher.New()

	// Set chromium path if specified
	if b.config.ChromiumPath != "" {
		l = l.Bin(b.config.ChromiumPath)
	}

	// Set user data directory if specified
	if b.config.UserDataDir != "" {
		l = l.UserDataDir(b.config.UserDataDir)
	}

	// Configure headless mode
	if b.config.Headless {
		l = l.Headless(true)
	} else {
		l = l.Headless(false)
	}

	// Disable automation detection flags
	l = l.Set("disable-blink-features", "AutomationControlled")
	l = l.Set("excludeSwitches", "enable-automation")
	l = l.Set("useAutomationExtension", "false")

	// Additional flags
	if b.config.DisableGPU {
		l = l.Set("disable-gpu", "")
	}
	if b.config.NoSandbox {
		l = l.Set("no-sandbox", "")
	}

	// Set window size
	l = l.Set("window-size", fmt.Sprintf("%d,%d", b.config.WindowWidth, b.config.WindowHeight))

	// Extra flags are ignored for now as launcher.Set requires specific format
	// Users can add flags via chromium_path if needed

	// Launch browser
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	b.instance = browser
	b.logger.Info("browser launched successfully")

	return nil
}

// Instance returns the underlying Rod browser instance
func (b *Browser) Instance() *rod.Browser {
	return b.instance
}

// Close gracefully shuts down the browser
func (b *Browser) Close() error {
	if b.instance == nil {
		return nil
	}

	b.logger.Info("closing browser")

	// Give pages time to cleanup
	time.Sleep(500 * time.Millisecond)

	if err := b.instance.Close(); err != nil {
		b.logger.Error("failed to close browser", zap.Error(err))
		return err
	}

	b.logger.Info("browser closed successfully")
	return nil
}

// MustInstance returns the browser instance or panics
func (b *Browser) MustInstance() *rod.Browser {
	if b.instance == nil {
		panic("browser instance is nil")
	}
	return b.instance
}
