package browser

import (
	"fmt"
	"os"
	"path/filepath"

	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

// Context represents an isolated browser context (like incognito mode)
type Context struct {
	page        *rod.Page
	fingerprint *stealth.Fingerprint
	logger      *zap.Logger
	sessionID   string
}

// NewContext creates a new isolated browser context with stealth features
func NewContext(browser *rod.Browser, sessionID string, fp *stealth.Fingerprint) (*Context, error) {
	logger := log.Session(sessionID)

	ctx := &Context{
		fingerprint: fp,
		logger:      logger,
		sessionID:   sessionID,
	}

	// Create incognito page
	page, err := browser.Incognito()
	if err != nil {
		return nil, fmt.Errorf("failed to create incognito context: %w", err)
	}

	ctx.page = page.MustPage()

	// Apply fingerprint and stealth
	if err := ctx.applyFingerprint(); err != nil {
		return nil, fmt.Errorf("failed to apply fingerprint: %w", err)
	}

	if err := ctx.injectStealthScript(); err != nil {
		return nil, fmt.Errorf("failed to inject stealth script: %w", err)
	}

	logger.Info("browser context created",
		zap.String("user_agent", fp.UserAgent),
		zap.Int("viewport_width", fp.ViewportWidth),
		zap.Int("viewport_height", fp.ViewportHeight),
	)

	return ctx, nil
}

// applyFingerprint applies the fingerprint to the browser context
func (c *Context) applyFingerprint() error {
	// Set user agent
	if err := c.page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: c.fingerprint.UserAgent,
	}); err != nil {
		return fmt.Errorf("failed to set user agent: %w", err)
	}

	// Set viewport
	if err := c.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  c.fingerprint.ViewportWidth,
		Height: c.fingerprint.ViewportHeight,
	}); err != nil {
		return fmt.Errorf("failed to set viewport: %w", err)
	}

	c.logger.Debug("fingerprint applied",
		zap.String("timezone", c.fingerprint.Timezone),
		zap.String("language", c.fingerprint.Language),
	)

	return nil
}

// injectStealthScript injects JavaScript to patch browser APIs
func (c *Context) injectStealthScript() error {
	// Read stealth.js file
	scriptPath := filepath.Join("assets", "js", "stealth.js")
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read stealth script: %w", err)
	}

	// Inject script before every page load
	script := string(scriptContent)

	// Add fingerprint-specific overrides
	script += fmt.Sprintf(`
		Object.defineProperty(navigator, 'language', { get: () => '%s' });
		Object.defineProperty(navigator, 'languages', { get: () => ['%s'] });
		Object.defineProperty(navigator, 'platform', { get: () => '%s' });
		Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => %d });
		
		// Override timezone
		Intl.DateTimeFormat = function() {
			return {
				resolvedOptions: () => ({ timeZone: '%s' })
			};
		};
	`,
		c.fingerprint.Language,
		c.fingerprint.Language,
		c.fingerprint.Platform,
		c.fingerprint.HardwareConcurrency,
		c.fingerprint.Timezone,
	)

	// Evaluate script in isolated world
	_, err = c.page.EvalOnNewDocument(script)
	if err != nil {
		return fmt.Errorf("failed to inject stealth script: %w", err)
	}

	c.logger.Debug("stealth script injected")
	return nil
}

// Page returns the underlying Rod page
func (c *Context) Page() *rod.Page {
	return c.page
}

// LoadCookies loads saved cookies into the context
func (c *Context) LoadCookies(cookies []*proto.NetworkCookie) error {
	if len(cookies) == 0 {
		c.logger.Debug("no cookies to load")
		return nil
	}

	// Convert NetworkCookie to NetworkCookieParam
	cookieParams := make([]*proto.NetworkCookieParam, len(cookies))
	for i, cookie := range cookies {
		cookieParams[i] = &proto.NetworkCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
			SameSite: cookie.SameSite,
			Expires:  cookie.Expires,
		}
	}

	if err := c.page.SetCookies(cookieParams); err != nil {
		return fmt.Errorf("failed to load cookies: %w", err)
	}

	c.logger.Info("cookies loaded", zap.Int("count", len(cookies)))
	return nil
}

// GetCookies extracts current cookies from the context
func (c *Context) GetCookies() ([]*proto.NetworkCookie, error) {
	cookies, err := c.page.Cookies(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}

	c.logger.Debug("cookies extracted", zap.Int("count", len(cookies)))
	return cookies, nil
}

// Navigate navigates to a URL
func (c *Context) Navigate(url string) error {
	c.logger.Info("navigating", zap.String("url", url))

	if err := c.page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	if err := c.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	c.logger.Debug("page loaded successfully")
	return nil
}

// Close closes the context
func (c *Context) Close() error {
	if c.page == nil {
		return nil
	}

	c.logger.Info("closing browser context")

	if err := c.page.Close(); err != nil {
		return fmt.Errorf("failed to close context: %w", err)
	}

	return nil
}

// Screenshot takes a screenshot and saves it
func (c *Context) Screenshot(filename string) error {
	data, err := c.page.Screenshot(true, nil)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	c.logger.Info("screenshot saved", zap.String("file", filename))
	return nil
}
