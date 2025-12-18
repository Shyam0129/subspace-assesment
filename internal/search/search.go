package search

import (
	"context"
	"fmt"
	"net/url"
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

// SearchProfiles searches for profiles based on configured targets
func (s *Service) SearchProfiles(ctx context.Context) ([]*storage.Profile, error) {
	s.log.Info("Starting profile search...")

	var allProfiles []*storage.Profile
	seenURLs := make(map[string]bool)

	for _, target := range s.cfg.Search.Targets {
		s.log.Infof("Searching for: %s in %s", target.JobTitle, target.Location)

		profiles, err := s.searchTarget(ctx, target)
		if err != nil {
			s.log.Errorf("Search failed for target %s: %v", target.JobTitle, err)
			continue
		}

		// Deduplicate profiles
		for _, profile := range profiles {
			if !seenURLs[profile.ProfileURL] {
				seenURLs[profile.ProfileURL] = true
				allProfiles = append(allProfiles, profile)
			}
		}

		// Delay between searches
		s.browser.GetStealth().RandomDelay("think")
	}

	s.log.Infof("Found %d unique profiles", len(allProfiles))
	s.store.LogActivity("search", "", "success", fmt.Sprintf("Found %d profiles", len(allProfiles)))

	return allProfiles, nil
}

// searchTarget performs a search for a specific target
func (s *Service) searchTarget(ctx context.Context, target config.SearchTarget) ([]*storage.Profile, error) {
	// Build search URL
	searchURL := s.buildSearchURL(target)

	s.log.Debugf("Search URL: %s", searchURL)

	// Navigate to search page
	if err := s.browser.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search: %w", err)
	}

	page := s.browser.GetPage()
	stealth := s.browser.GetStealth()

	// Wait for search results to load
	time.Sleep(3 * time.Second)

	var profiles []*storage.Profile

	// Iterate through pagination
	for i := 0; i < s.cfg.Search.PaginationLimit; i++ {
		s.log.Infof("Processing search results page %d", i+1)

		// Scroll to load all results
		stealth.RandomScroll(page)
		stealth.RandomDelay("scroll")

		// Extract profile URLs from current page
		pageProfiles, err := s.extractProfilesFromPage(page, target)
		if err != nil {
			s.log.Errorf("Failed to extract profiles from page %d: %v", i+1, err)
			break
		}

		profiles = append(profiles, pageProfiles...)

		s.log.Infof("Extracted %d profiles from page %d", len(pageProfiles), i+1)

		// Check if we've reached the limit
		if len(profiles) >= s.cfg.Search.MaxResultsPerSearch {
			profiles = profiles[:s.cfg.Search.MaxResultsPerSearch]
			break
		}

		// Try to go to next page
		if !s.goToNextPage(page, stealth) {
			s.log.Info("No more pages available")
			break
		}

		stealth.RandomDelay("action")
	}

	return profiles, nil
}

// buildSearchURL constructs the LinkedIn search URL
func (s *Service) buildSearchURL(target config.SearchTarget) string {
	baseURL := "https://www.linkedin.com/search/results/people/"

	params := url.Values{}

	// Build keywords query
	keywords := []string{}
	if target.JobTitle != "" {
		keywords = append(keywords, target.JobTitle)
	}
	if target.Keywords != "" {
		keywords = append(keywords, target.Keywords)
	}

	if len(keywords) > 0 {
		params.Add("keywords", strings.Join(keywords, " "))
	}

	// Add location if specified
	if target.Location != "" {
		params.Add("geoUrn", target.Location)
	}

	// Add filters for 2nd and 3rd degree connections
	params.Add("network", "[\"S\",\"O\"]")

	return baseURL + "?" + params.Encode()
}

// extractProfilesFromPage extracts profile information from the current page
func (s *Service) extractProfilesFromPage(page *rod.Page, target config.SearchTarget) ([]*storage.Profile, error) {
	// Wait for search results container
	time.Sleep(2 * time.Second)

	// Find all profile cards
	// LinkedIn search results are in list items with specific classes
	elements, err := page.Elements(".reusable-search__result-container")
	if err != nil {
		return nil, fmt.Errorf("failed to find search results: %w", err)
	}

	var profiles []*storage.Profile

	for _, element := range elements {
		profile, err := s.extractProfileFromElement(element, target)
		if err != nil {
			s.log.Debugf("Failed to extract profile: %v", err)
			continue
		}

		if profile != nil {
			// Save to database
			profileID, err := s.store.SaveProfile(profile)
			if err != nil {
				s.log.Errorf("Failed to save profile: %v", err)
				continue
			}
			profile.ID = profileID
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// extractProfileFromElement extracts profile data from a search result element
func (s *Service) extractProfileFromElement(element *rod.Element, target config.SearchTarget) (*storage.Profile, error) {
	// Extract profile URL
	linkElement, err := element.Element("a.app-aware-link")
	if err != nil {
		return nil, fmt.Errorf("profile link not found: %w", err)
	}

	profileURL, err := linkElement.Attribute("href")
	if err != nil || profileURL == nil {
		return nil, fmt.Errorf("failed to get profile URL: %w", err)
	}

	// Clean URL (remove query parameters)
	cleanURL := strings.Split(*profileURL, "?")[0]

	// Extract name
	nameElement, err := element.Element(".entity-result__title-text a span[aria-hidden='true']")
	var name string
	if err == nil {
		nameText, _ := nameElement.Text()
		name = strings.TrimSpace(nameText)
	}

	// Extract job title
	jobTitleElement, err := element.Element(".entity-result__primary-subtitle")
	var jobTitle string
	if err == nil {
		jobTitleText, _ := jobTitleElement.Text()
		jobTitle = strings.TrimSpace(jobTitleText)
	}

	// Extract company/location
	secondaryElement, err := element.Element(".entity-result__secondary-subtitle")
	var company string
	if err == nil {
		companyText, _ := secondaryElement.Text()
		company = strings.TrimSpace(companyText)
	}

	profile := &storage.Profile{
		ProfileURL:   cleanURL,
		Name:         name,
		JobTitle:     jobTitle,
		Company:      company,
		Location:     target.Location,
		Keywords:     target.Keywords,
		DiscoveredAt: time.Now(),
	}

	s.log.Debugf("Extracted profile: %s - %s at %s", name, jobTitle, company)

	return profile, nil
}

// goToNextPage attempts to navigate to the next page of search results
func (s *Service) goToNextPage(page *rod.Page, st *stealth.Stealth) bool {
	// Look for "Next" button
	nextButton, err := page.Element("button[aria-label='Next']")
	if err != nil {
		return false
	}

	// Check if button is disabled
	disabled, _ := nextButton.Attribute("disabled")
	if disabled != nil {
		return false
	}

	// Click next button with human-like behavior
	if err := st.HumanClick(nextButton); err != nil {
		s.log.Errorf("Failed to click next button: %v", err)
		return false
	}

	// Wait for new results to load
	time.Sleep(3 * time.Second)

	return true
}

// SearchByURL searches for a specific profile by URL
func (s *Service) SearchByURL(profileURL string) (*storage.Profile, error) {
	s.log.Infof("Searching for profile: %s", profileURL)

	// Check if profile already exists in database
	existingProfile, err := s.store.GetProfileByURL(profileURL)
	if err == nil && existingProfile != nil {
		return existingProfile, nil
	}

	// Navigate to profile
	if err := s.browser.Navigate(profileURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to profile: %w", err)
	}

	// Extract profile information
	page := s.browser.GetPage()
	time.Sleep(2 * time.Second)

	profile := &storage.Profile{
		ProfileURL:   profileURL,
		DiscoveredAt: time.Now(),
	}

	// Extract name
	nameElement, err := page.Element("h1.text-heading-xlarge")
	if err == nil {
		nameText, _ := nameElement.Text()
		profile.Name = strings.TrimSpace(nameText)
	}

	// Extract job title
	titleElement, err := page.Element(".text-body-medium.break-words")
	if err == nil {
		titleText, _ := titleElement.Text()
		profile.JobTitle = strings.TrimSpace(titleText)
	}

	// Save to database
	profileID, err := s.store.SaveProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}
	profile.ID = profileID

	return profile, nil
}

