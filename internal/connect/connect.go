package connect

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/stealth"
	"linkedin-automation/internal/storage"

	"github.com/go-rod/rod"
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

// SendConnectionRequests sends connection requests to profiles
func (s *Service) SendConnectionRequests(ctx context.Context, profiles []*storage.Profile) (int, error) {
	s.log.Info("Starting to send connection requests...")

	sent := 0

	for _, profile := range profiles {
		select {
		case <-ctx.Done():
			s.log.Info("Context cancelled, stopping connection requests")
			return sent, ctx.Err()
		default:
		}

		// Check rate limits
		if !s.canSendConnection() {
			s.log.Warn("Rate limit reached for connections")
			break
		}

		// Check if already sent
		alreadySent, err := s.store.IsConnectionSent(profile.ProfileURL)
		if err != nil {
			s.log.Errorf("Failed to check connection status: %v", err)
			continue
		}

		if alreadySent {
			s.log.Debugf("Connection already sent to %s, skipping", profile.ProfileURL)
			continue
		}

		// Send connection request
		if err := s.sendConnectionRequest(profile); err != nil {
			s.log.Errorf("Failed to send connection to %s: %v", profile.ProfileURL, err)
			s.store.LogActivity("connection_request", profile.ProfileURL, "failed", err.Error())
			continue
		}

		sent++
		s.log.Infof("Connection request sent to %s (%d/%d)", profile.Name, sent, len(profiles))

		// Random delay between requests
		s.browser.GetStealth().RandomDelay("action")

		// Longer delay every few requests
		if sent%5 == 0 {
			s.browser.GetStealth().RandomDelay("think")
		}
	}

	s.log.Infof("Sent %d connection requests", sent)
	return sent, nil
}

// sendConnectionRequest sends a connection request to a single profile
func (s *Service) sendConnectionRequest(profile *storage.Profile) error {
	s.log.Infof("Sending connection request to: %s", profile.ProfileURL)

	// Navigate to profile
	if err := s.browser.Navigate(profile.ProfileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	// Simulate reading the profile
	stealth.SimulateReading(page)
	stealth.RandomDelay("think")

	// Find the Connect button
	connectButton, err := s.findConnectButton(page)
	if err != nil {
		return fmt.Errorf("connect button not found: %w", err)
	}

	// Click Connect button
	if err := stealth.HumanClick(connectButton); err != nil {
		return fmt.Errorf("failed to click connect: %w", err)
	}

	stealth.RandomDelay("action")

	// Check if we need to add a note
	if s.cfg.Connection.SendNote {
		if err := s.addConnectionNote(page, stealth, profile); err != nil {
			s.log.Warnf("Failed to add note, sending without note: %v", err)
			// Try to send without note
			if err := s.clickSendButton(page, stealth, false); err != nil {
				return fmt.Errorf("failed to send connection: %w", err)
			}
		}
	} else {
		// Send without note
		if err := s.clickSendButton(page, stealth, false); err != nil {
			return fmt.Errorf("failed to send connection: %w", err)
		}
	}

	// Save to database
	connectionReq := &storage.ConnectionRequest{
		ProfileID:  profile.ID,
		ProfileURL: profile.ProfileURL,
		SentAt:     time.Now(),
		Status:     "pending",
	}

	if err := s.store.SaveConnectionRequest(connectionReq); err != nil {
		return fmt.Errorf("failed to save connection request: %w", err)
	}

	s.store.LogActivity("connection_request", profile.ProfileURL, "success", "")

	return nil
}

// findConnectButton finds the Connect button on a profile page
func (s *Service) findConnectButton(page *rod.Page) (*rod.Element, error) {
	// LinkedIn has different button structures, try multiple selectors
	selectors := []string{
		"button[aria-label*='Connect']",
		"button.pvs-profile-actions__action:has-text('Connect')",
		"button:has-text('Connect')",
		".pvs-profile-actions button:has-text('Connect')",
	}

	for _, selector := range selectors {
		element, err := page.Element(selector)
		if err == nil {
			return element, nil
		}
	}

	return nil, fmt.Errorf("connect button not found with any selector")
}

// addConnectionNote adds a personalized note to the connection request
func (s *Service) addConnectionNote(page *rod.Page, st *stealth.Stealth, profile *storage.Profile) error {
	// Look for "Add a note" button
	addNoteButton, err := page.Element("button[aria-label*='Add a note']")
	if err != nil {
		// Try alternative selector
		addNoteButton, err = page.Element("button:has-text('Add a note')")
		if err != nil {
			return fmt.Errorf("add note button not found: %w", err)
		}
	}

	// Click "Add a note"
	if err := st.HumanClick(addNoteButton); err != nil {
		return fmt.Errorf("failed to click add note: %w", err)
	}

	st.RandomDelay("action")

	// Find note textarea
	noteTextarea, err := st.WaitForElement(page, "textarea[name='message']", 5*time.Second)
	if err != nil {
		return fmt.Errorf("note textarea not found: %w", err)
	}

	// Generate personalized note
	note := s.generateNote(profile)

	// Type note with human-like behavior
	if err := st.HumanType(noteTextarea, note); err != nil {
		return fmt.Errorf("failed to type note: %w", err)
	}

	st.RandomDelay("think")

	// Click Send button
	return s.clickSendButton(page, st, true)
}

// clickSendButton clicks the Send button
func (s *Service) clickSendButton(page *rod.Page, st *stealth.Stealth, withNote bool) error {
	var sendButton *rod.Element
	var err error

	if withNote {
		sendButton, err = page.Element("button[aria-label*='Send now']")
	} else {
		sendButton, err = page.Element("button[aria-label*='Send without a note']")
		if err != nil {
			// Try generic send button
			sendButton, err = page.Element("button:has-text('Send')")
		}
	}

	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	if err := st.HumanClick(sendButton); err != nil {
		return fmt.Errorf("failed to click send: %w", err)
	}

	// Wait for confirmation
	time.Sleep(2 * time.Second)

	return nil
}

// generateNote generates a personalized connection note
func (s *Service) generateNote(profile *storage.Profile) string {
	if len(s.cfg.Connection.NoteTemplates) == 0 {
		return "Hi, I'd love to connect!"
	}

	// Select random template
	template := s.cfg.Connection.NoteTemplates[rand.Intn(len(s.cfg.Connection.NoteTemplates))]

	// Extract first name
	firstName := extractFirstName(profile.Name)

	// Replace placeholders
	note := strings.ReplaceAll(template, "{{FirstName}}", firstName)
	note = strings.ReplaceAll(note, "{{Company}}", profile.Company)
	note = strings.ReplaceAll(note, "{{Field}}", profile.Keywords)
	note = strings.ReplaceAll(note, "{{Topic}}", profile.JobTitle)

	// Ensure note doesn't exceed max length
	if len(note) > s.cfg.Connection.NoteMaxLength {
		note = note[:s.cfg.Connection.NoteMaxLength-3] + "..."
	}

	return note
}

// extractFirstName extracts the first name from a full name
func extractFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return "there"
}

// canSendConnection checks if we can send more connections based on rate limits
func (s *Service) canSendConnection() bool {
	// Check daily limit
	dailyStats := s.store.GetTodayStats()
	if dailyStats.ConnectionsSent >= s.cfg.RateLimits.Connections.PerDay {
		s.log.Warn("Daily connection limit reached")
		return false
	}

	// Check hourly limit
	hourlyStats := s.store.GetHourlyStats()
	if hourlyStats.ConnectionsSent >= s.cfg.RateLimits.Connections.PerHour {
		s.log.Warn("Hourly connection limit reached")
		return false
	}

	return true
}

// WithdrawPendingRequests withdraws pending connection requests (optional feature)
func (s *Service) WithdrawPendingRequests() error {
	s.log.Info("Withdrawing old pending requests...")

	// Navigate to "My Network" -> "Manage invitations"
	if err := s.browser.Navigate("https://www.linkedin.com/mynetwork/invitation-manager/sent/"); err != nil {
		return fmt.Errorf("failed to navigate to invitations: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	time.Sleep(3 * time.Second)

	// Find withdraw buttons
	withdrawButtons, err := page.Elements("button[aria-label*='Withdraw']")
	if err != nil {
		return fmt.Errorf("no withdraw buttons found: %w", err)
	}

	withdrawn := 0
	for i, button := range withdrawButtons {
		if i >= 10 { // Limit to 10 withdrawals per run
			break
		}

		if err := stealth.HumanClick(button); err != nil {
			s.log.Errorf("Failed to click withdraw: %v", err)
			continue
		}

		stealth.RandomDelay("action")

		// Confirm withdrawal
		confirmButton, err := page.Element("button[data-control-name='withdraw_single']")
		if err == nil {
			stealth.HumanClick(confirmButton)
			withdrawn++
		}

		stealth.RandomDelay("action")
	}

	s.log.Infof("Withdrew %d pending requests", withdrawn)
	return nil
}


