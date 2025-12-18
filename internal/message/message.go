package message

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"
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

// SendMessages sends messages to accepted connections
func (s *Service) SendMessages(ctx context.Context) (int, error) {
	if !s.cfg.Messaging.Enabled {
		s.log.Info("Messaging is disabled in config")
		return 0, nil
	}

	s.log.Info("Starting to send messages to accepted connections...")

	// Get accepted connections that haven't been messaged
	connections, err := s.store.GetAcceptedConnections()
	if err != nil {
		return 0, fmt.Errorf("failed to get accepted connections: %w", err)
	}

	if len(connections) == 0 {
		s.log.Info("No accepted connections to message")
		return 0, nil
	}

	s.log.Infof("Found %d accepted connections to message", len(connections))

	sent := 0

	for _, conn := range connections {
		select {
		case <-ctx.Done():
			s.log.Info("Context cancelled, stopping messaging")
			return sent, ctx.Err()
		default:
		}

		// Check rate limits
		if !s.canSendMessage() {
			s.log.Warn("Rate limit reached for messages")
			break
		}

		// Check if connection was accepted recently (respect delay)
		if conn.AcceptedAt != nil {
			hoursSinceAccepted := time.Since(*conn.AcceptedAt).Hours()
			if hoursSinceAccepted < float64(s.cfg.Messaging.DelayAfterConnectionHours) {
				s.log.Debugf("Connection accepted too recently, skipping: %s", conn.ProfileURL)
				continue
			}
		}

		// Send message
		if err := s.sendMessage(&conn); err != nil {
			s.log.Errorf("Failed to send message to %s: %v", conn.ProfileURL, err)
			s.store.LogActivity("message", conn.ProfileURL, "failed", err.Error())
			continue
		}

		sent++
		s.log.Infof("Message sent (%d/%d)", sent, len(connections))

		// Random delay between messages
		s.browser.GetStealth().RandomDelay("action")

		// Longer delay every few messages
		if sent%3 == 0 {
			s.browser.GetStealth().RandomDelay("think")
		}
	}

	s.log.Infof("Sent %d messages", sent)
	return sent, nil
}

// sendMessage sends a message to a specific connection
func (s *Service) sendMessage(conn *storage.ConnectionRequest) error {
	s.log.Infof("Sending message to: %s", conn.ProfileURL)

	// Navigate to messaging page with the profile
	messagingURL := s.getMessagingURL(conn.ProfileURL)

	if err := s.browser.Navigate(messagingURL); err != nil {
		return fmt.Errorf("failed to navigate to messaging: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	// Wait for messaging interface to load
	time.Sleep(3 * time.Second)

	// Find message input box
	messageBox, err := stealth.WaitForElement(page, ".msg-form__contenteditable", 10*time.Second)
	if err != nil {
		// Try alternative selector
		messageBox, err = stealth.WaitForElement(page, "div[role='textbox']", 5*time.Second)
		if err != nil {
			return fmt.Errorf("message box not found: %w", err)
		}
	}

	// Generate message content
	messageContent := s.generateMessage(conn)

	// Click on message box
	if err := stealth.HumanClick(messageBox); err != nil {
		return fmt.Errorf("failed to click message box: %w", err)
	}

	stealth.RandomDelay("action")

	// Type message with human-like behavior
	if err := stealth.HumanType(messageBox, messageContent); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	stealth.RandomDelay("think")

	// Find and click send button
	sendButton, err := s.findSendButton(page)
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	if err := stealth.HumanClick(sendButton); err != nil {
		return fmt.Errorf("failed to click send: %w", err)
	}

	// Wait for message to be sent
	time.Sleep(2 * time.Second)

	// Save to database
	msg := &storage.Message{
		ProfileID:  conn.ProfileID,
		ProfileURL: conn.ProfileURL,
		Content:    messageContent,
		SentAt:     time.Now(),
		Status:     "sent",
	}

	if err := s.store.SaveMessage(msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.store.LogActivity("message", conn.ProfileURL, "success", "")

	return nil
}

// getMessagingURL constructs the messaging URL for a profile
func (s *Service) getMessagingURL(profileURL string) string {
	// Extract profile ID from URL
	// LinkedIn profile URLs are like: https://www.linkedin.com/in/username/
	parts := strings.Split(strings.TrimSuffix(profileURL, "/"), "/")
	if len(parts) > 0 {
		username := parts[len(parts)-1]
		return fmt.Sprintf("https://www.linkedin.com/messaging/thread/new/?recipient=%s", username)
	}

	// Fallback: navigate to profile and use message button
	return profileURL
}

// findSendButton finds the send button in the messaging interface
func (s *Service) findSendButton(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"button.msg-form__send-button",
		"button[type='submit']",
		"button:has-text('Send')",
		".msg-form__send-button",
	}

	for _, selector := range selectors {
		element, err := page.Element(selector)
		if err == nil {
			// Check if button is enabled
			disabled, _ := element.Attribute("disabled")
			if disabled == nil {
				return element, nil
			}
		}
	}

	return nil, fmt.Errorf("send button not found or disabled")
}

// generateMessage generates a personalized message
func (s *Service) generateMessage(conn *storage.ConnectionRequest) string {
	if len(s.cfg.Messaging.Templates) == 0 {
		return "Thanks for connecting! Looking forward to staying in touch."
	}

	// Select random template
	template := s.cfg.Messaging.Templates[rand.Intn(len(s.cfg.Messaging.Templates))]

	// Get profile information
	profile, err := s.store.GetProfileByURL(conn.ProfileURL)
	if err != nil || profile == nil {
		return template
	}

	// Extract first name
	firstName := extractFirstName(profile.Name)

	// Replace placeholders
	message := strings.ReplaceAll(template, "{{FirstName}}", firstName)
	message = strings.ReplaceAll(message, "{{Company}}", profile.Company)
	message = strings.ReplaceAll(message, "{{Topic}}", profile.Keywords)
	message = strings.ReplaceAll(message, "{{Field}}", profile.JobTitle)

	return message
}

// extractFirstName extracts the first name from a full name
func extractFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return "there"
}

// canSendMessage checks if we can send more messages based on rate limits
func (s *Service) canSendMessage() bool {
	// Check daily limit
	dailyStats := s.store.GetTodayStats()
	if dailyStats.MessagesSent >= s.cfg.RateLimits.Messages.PerDay {
		s.log.Warn("Daily message limit reached")
		return false
	}

	// Check hourly limit
	hourlyStats := s.store.GetHourlyStats()
	if hourlyStats.MessagesSent >= s.cfg.RateLimits.Messages.PerHour {
		s.log.Warn("Hourly message limit reached")
		return false
	}

	return true
}

// SendMessageToProfile sends a message to a specific profile URL
func (s *Service) SendMessageToProfile(profileURL, message string) error {
	s.log.Infof("Sending custom message to: %s", profileURL)

	// Get or create profile
	profile, err := s.store.GetProfileByURL(profileURL)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if profile == nil {
		// Create minimal profile entry
		profile = &storage.Profile{
			ProfileURL:   profileURL,
			DiscoveredAt: time.Now(),
		}
		profileID, err := s.store.SaveProfile(profile)
		if err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}
		profile.ID = profileID
	}

	// Navigate to messaging
	messagingURL := s.getMessagingURL(profileURL)
	if err := s.browser.Navigate(messagingURL); err != nil {
		return fmt.Errorf("failed to navigate to messaging: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	time.Sleep(3 * time.Second)

	// Find message box
	messageBox, err := stealth.WaitForElement(page, ".msg-form__contenteditable", 10*time.Second)
	if err != nil {
		messageBox, err = stealth.WaitForElement(page, "div[role='textbox']", 5*time.Second)
		if err != nil {
			return fmt.Errorf("message box not found: %w", err)
		}
	}

	// Click and type message
	if err := stealth.HumanClick(messageBox); err != nil {
		return fmt.Errorf("failed to click message box: %w", err)
	}

	stealth.RandomDelay("action")

	if err := stealth.HumanType(messageBox, message); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	stealth.RandomDelay("think")

	// Send message
	sendButton, err := s.findSendButton(page)
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	if err := stealth.HumanClick(sendButton); err != nil {
		return fmt.Errorf("failed to click send: %w", err)
	}

	time.Sleep(2 * time.Second)

	// Save to database
	msg := &storage.Message{
		ProfileID:  profile.ID,
		ProfileURL: profileURL,
		Content:    message,
		SentAt:     time.Now(),
		Status:     "sent",
	}

	if err := s.store.SaveMessage(msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.store.LogActivity("message", profileURL, "success", "")

	return nil
}
