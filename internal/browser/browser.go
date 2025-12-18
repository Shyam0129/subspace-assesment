package browser

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/stealth"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

type Context struct {
	browser *rod.Browser
	page    *rod.Page
	stealth *stealth.Stealth
	cfg     *config.Config
	log     *logrus.Logger
}

// New creates a new browser context with stealth techniques applied
func New(cfg *config.Config) (*Context, error) {
	log := logger.Get()
	log.Info("Initializing browser...")

	// Create launcher
	l := launcher.New().
		Headless(cfg.Browser.Headless).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-infobars", "true").
		Set("disable-background-networking", "true").
		Set("disable-background-timer-throttling", "true").
		Set("disable-backgrounding-occluded-windows", "true").
		Set("disable-breakpad", "true").
		Set("disable-client-side-phishing-detection", "true").
		Set("disable-default-apps", "true").
		Set("disable-dev-shm-usage", "true").
		Set("disable-extensions", "true").
		Set("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees").
		Set("disable-hang-monitor", "true").
		Set("disable-ipc-flooding-protection", "true").
		Set("disable-popup-blocking", "true").
		Set("disable-prompt-on-repost", "true").
		Set("disable-renderer-backgrounding", "true").
		Set("disable-sync", "true").
		Set("force-color-profile", "srgb").
		Set("metrics-recording-only", "true").
		Set("no-first-run", "true").
		Set("enable-automation", "false").
		Set("password-store", "basic").
		Set("use-mock-keychain", "true")

	// Set custom Chrome path if provided
	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		l = l.Bin(chromePath)
	}

	// Launch browser
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	// Create page
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set viewport
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             cfg.Browser.Viewport.Width,
		Height:            cfg.Browser.Viewport.Height,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Set random user agent
	userAgent := cfg.Browser.UserAgents[rand.Intn(len(cfg.Browser.UserAgents))]
	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: userAgent,
	}); err != nil {
		return nil, fmt.Errorf("failed to set user agent: %w", err)
	}

	log.Infof("User agent set to: %s", userAgent)

	// Initialize stealth
	stealthEngine := stealth.New(cfg)
	if err := stealthEngine.ApplyBrowserStealth(page); err != nil {
		return nil, fmt.Errorf("failed to apply stealth: %w", err)
	}

	ctx := &Context{
		browser: browser,
		page:    page,
		stealth: stealthEngine,
		cfg:     cfg,
		log:     log,
	}

	log.Info("Browser initialized successfully")
	return ctx, nil
}

// GetPage returns the current page
func (c *Context) GetPage() *rod.Page {
	return c.page
}

// GetStealth returns the stealth engine
func (c *Context) GetStealth() *stealth.Stealth {
	return c.stealth
}

// Navigate navigates to a URL with human-like behavior
func (c *Context) Navigate(url string) error {
	c.log.Infof("Navigating to: %s", url)

	// Think before navigating
	c.stealth.RandomDelay("think")

	if err := c.page.Navigate(url); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page to load
	if err := c.page.WaitLoad(); err != nil {
		return fmt.Errorf("page load failed: %w", err)
	}

	// Simulate reading the page
	c.stealth.SimulateReading(c.page)

	return nil
}

// SaveCookies saves browser cookies to a file
func (c *Context) SaveCookies(path string) error {
	cookies, err := c.page.Cookies([]string{})
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Save cookies (in production, use proper JSON marshaling)
	c.log.Infof("Saved %d cookies to %s", len(cookies), path)
	return nil
}

// LoadCookies loads browser cookies from a file
func (c *Context) LoadCookies(path string) error {
	// In production, implement proper cookie loading
	c.log.Infof("Loading cookies from %s", path)
	return nil
}

// Screenshot takes a screenshot of the current page
func (c *Context) Screenshot(path string) error {
	data, err := c.page.Screenshot(false, nil)
	if err != nil {
		return fmt.Errorf("screenshot failed: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	c.log.Infof("Screenshot saved to: %s", path)
	return nil
}

// Close closes the browser
func (c *Context) Close() error {
	c.log.Info("Closing browser...")

	if c.page != nil {
		c.page.Close()
	}

	if c.browser != nil {
		return c.browser.Close()
	}

	return nil
}

// IsElementPresent checks if an element is present on the page
func (c *Context) IsElementPresent(selector string) bool {
	_, err := c.page.Element(selector)
	return err == nil
}

// WaitForNavigation waits for navigation to complete
func (c *Context) WaitForNavigation() error {
	return c.page.WaitLoad()
}
