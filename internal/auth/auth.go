package auth

import (
	"context"
	"fmt"
	"time"

	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/storage"

	"github.com/sirupsen/logrus"
)

type Service struct {
	browser *browser.Context
	store   *storage.Storage
	cfg     *config.Config
	log     *logrus.Logger
}

func New(browser *browser.Context, store *storage.Storage, cfg *config.Config) *Service {
	return &Service{
		browser: browser,
		store:   store,
		cfg:     cfg,
		log:     logger.Get(),
	}
}

// Login authenticates with LinkedIn
func (s *Service) Login(ctx context.Context) error {
	s.log.Info("Starting LinkedIn authentication...")

	// Try to load existing cookies first
	cookiePath := s.cfg.Storage.CookiePath
	if err := s.browser.LoadCookies(cookiePath); err == nil {
		s.log.Info("Loaded existing cookies, checking session...")

		// Navigate to LinkedIn to check if session is valid
		if err := s.browser.Navigate("https://www.linkedin.com/feed/"); err == nil {
			// Check if we're logged in
			if s.isLoggedIn() {
				s.log.Info("Session is valid, skipping login")
				return nil
			}
		}
	}

	s.log.Info("No valid session found, performing fresh login...")

	// Navigate to LinkedIn login page
	if err := s.browser.Navigate("https://www.linkedin.com/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	// Wait for login form
	emailInput, err := stealth.WaitForElement(page, "#username", 10*time.Second)
	if err != nil {
		return fmt.Errorf("email input not found: %w", err)
	}

	// Type email with human-like behavior
	s.log.Info("Entering email...")
	if err := stealth.HumanType(emailInput, s.cfg.LinkedIn.Email); err != nil {
		return fmt.Errorf("failed to enter email: %w", err)
	}

	stealth.RandomDelay("action")

	// Find password input
	passwordInput, err := stealth.WaitForElement(page, "#password", 5*time.Second)
	if err != nil {
		return fmt.Errorf("password input not found: %w", err)
	}

	// Type password with human-like behavior
	s.log.Info("Entering password...")
	if err := stealth.HumanType(passwordInput, s.cfg.LinkedIn.Password); err != nil {
		return fmt.Errorf("failed to enter password: %w", err)
	}

	stealth.RandomDelay("think")

	// Find and click login button
	loginButton, err := stealth.WaitForElement(page, "button[type='submit']", 5*time.Second)
	if err != nil {
		return fmt.Errorf("login button not found: %w", err)
	}

	s.log.Info("Clicking login button...")
	if err := stealth.HumanClick(loginButton); err != nil {
		return fmt.Errorf("failed to click login: %w", err)
	}

	// Wait for navigation
	time.Sleep(5 * time.Second)

	// Check for common login issues
	if err := s.checkLoginIssues(); err != nil {
		return err
	}

	// Verify login success
	if !s.isLoggedIn() {
		// Take screenshot for debugging
		s.browser.Screenshot("./logs/login_failed.png")
		return fmt.Errorf("login verification failed")
	}

	s.log.Info("Login successful!")

	// Save cookies for future use
	if err := s.browser.SaveCookies(cookiePath); err != nil {
		s.log.Warnf("Failed to save cookies: %v", err)
	}

	// Log activity
	s.store.LogActivity("login", "https://www.linkedin.com", "success", "")

	return nil
}

// isLoggedIn checks if the user is currently logged in
func (s *Service) isLoggedIn() bool {
	page := s.browser.GetPage()

	// Check for elements that only appear when logged in
	// LinkedIn's feed or navigation bar
	if s.browser.IsElementPresent("nav.global-nav") {
		return true
	}

	if s.browser.IsElementPresent(".feed-identity-module") {
		return true
	}

	// Check if we're on the feed page
	currentURL := page.MustInfo().URL
	if currentURL == "https://www.linkedin.com/feed/" {
		return true
	}

	return false
}

// checkLoginIssues checks for common login issues
func (s *Service) checkLoginIssues() error {
	page := s.browser.GetPage()

	// Check for CAPTCHA
	if s.browser.IsElementPresent("#captcha-internal") {
		s.browser.Screenshot("./logs/captcha_detected.png")
		s.store.LogActivity("login", "https://www.linkedin.com", "captcha", "CAPTCHA detected")
		return fmt.Errorf("CAPTCHA detected - manual intervention required")
	}

	// Check for 2FA/verification
	if s.browser.IsElementPresent("input[name='pin']") {
		s.browser.Screenshot("./logs/2fa_detected.png")
		s.store.LogActivity("login", "https://www.linkedin.com", "2fa", "2FA verification required")
		return fmt.Errorf("2FA verification required - manual intervention needed")
	}

	// Check for security challenge
	if s.browser.IsElementPresent(".challenge-dialog") {
		s.browser.Screenshot("./logs/security_challenge.png")
		s.store.LogActivity("login", "https://www.linkedin.com", "challenge", "Security challenge detected")
		return fmt.Errorf("security challenge detected - manual intervention required")
	}

	// Check for incorrect credentials
	currentURL := page.MustInfo().URL
	if currentURL == "https://www.linkedin.com/login" || currentURL == "https://www.linkedin.com/uas/login-submit" {
		// Still on login page, check for error messages
		if s.browser.IsElementPresent(".form__label--error") {
			s.browser.Screenshot("./logs/login_error.png")
			s.store.LogActivity("login", "https://www.linkedin.com", "failed", "Invalid credentials")
			return fmt.Errorf("login failed - invalid credentials")
		}
	}

	return nil
}

// Logout logs out from LinkedIn
func (s *Service) Logout() error {
	s.log.Info("Logging out from LinkedIn...")

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	// Navigate to LinkedIn home
	if err := s.browser.Navigate("https://www.linkedin.com/feed/"); err != nil {
		return fmt.Errorf("failed to navigate to feed: %w", err)
	}

	// Click on "Me" dropdown
	meButton, err := stealth.WaitForElement(page, "button.global-nav__primary-link--me", 10*time.Second)
	if err != nil {
		return fmt.Errorf("me button not found: %w", err)
	}

	if err := stealth.HumanClick(meButton); err != nil {
		return fmt.Errorf("failed to click me button: %w", err)
	}

	stealth.RandomDelay("action")

	// Click sign out
	signOutButton, err := stealth.WaitForElement(page, "a[href*='logout']", 5*time.Second)
	if err != nil {
		return fmt.Errorf("sign out button not found: %w", err)
	}

	if err := stealth.HumanClick(signOutButton); err != nil {
		return fmt.Errorf("failed to click sign out: %w", err)
	}

	s.log.Info("Logged out successfully")
	s.store.LogActivity("logout", "https://www.linkedin.com", "success", "")

	return nil
}

// VerifySession verifies that the current session is still valid
func (s *Service) VerifySession() error {
	if !s.isLoggedIn() {
		return fmt.Errorf("session is no longer valid")
	}
	return nil
}
