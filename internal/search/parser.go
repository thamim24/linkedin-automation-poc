package search

import (
	"fmt"
	"strings"
	"time"

	"linkedin-automation-poc/internal/log"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// ProfileResult represents a parsed profile from search results
type ProfileResult struct {
	ProfileURL       string
	Name             string
	Headline         string
	Location         string
	ImageURL         string
	IsConnected      bool
	ConnectionDegree string
}

// Parser extracts profile data from search results
type Parser struct {
	page   *rod.Page
	logger *zap.Logger
}

// NewParser creates a new search results parser
func NewParser(page *rod.Page) *Parser {
	return &Parser{
		page:   page,
		logger: log.Module("parser"),
	}
}

// ParseResults extracts all profiles from current search results page
func (p *Parser) ParseResults() ([]ProfileResult, error) {
	p.logger.Info("starting profile parsing with multiple strategies")

	// Strategy 1: Try container-based parsing
	profiles, err := p.parseWithContainers()
	if err == nil && len(profiles) > 0 {
		p.logger.Info("container parsing successful", zap.Int("count", len(profiles)))
		return profiles, nil
	}

	// Strategy 2: Fallback to direct link parsing
	p.logger.Warn("container parsing failed, using direct link strategy")
	profiles, err = p.parseDirectLinks()
	if err != nil {
		return nil, err
	}

	p.logger.Info("direct link parsing complete", zap.Int("count", len(profiles)))
	return profiles, nil
}

// parseWithContainers tries to find profile cards using container selectors
func (p *Parser) parseWithContainers() ([]ProfileResult, error) {
	containerSelectors := []string{
		"li.reusable-search__result-container",
		"div.entity-result",
		"li.reusable-search__result",
		".search-results-container li",
		"ul.reusable-search__entity-result-list li",
	}

	var results []*rod.Element
	var usedSelector string

	for _, selector := range containerSelectors {
		p.logger.Debug("trying container selector", zap.String("selector", selector))
		elements, err := p.page.Elements(selector)
		if err == nil && len(elements) > 0 {
			results = elements
			usedSelector = selector
			p.logger.Info("found containers",
				zap.String("selector", selector),
				zap.Int("count", len(elements)))
			break
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no containers found")
	}

	var profiles []ProfileResult
	for i, result := range results {
		profile, err := p.parseProfileCard(result)
		if err != nil {
			p.logger.Debug("skipping card",
				zap.Int("index", i),
				zap.String("selector", usedSelector),
				zap.Error(err))
			continue
		}

		if profile != nil && profile.ProfileURL != "" {
			profiles = append(profiles, *profile)
			p.logger.Debug("parsed profile",
				zap.Int("index", i),
				zap.String("url", profile.ProfileURL),
				zap.String("name", profile.Name))
		}
	}

	return profiles, nil
}

// parseDirectLinks finds all profile links directly (fallback strategy)
func (p *Parser) parseDirectLinks() ([]ProfileResult, error) {
	p.logger.Info("using direct link parsing strategy")

	// Find all app-aware-link elements with /in/ in href
	links, err := p.page.Elements("a.app-aware-link[href*='/in/']")
	if err != nil {
		// Try without class restriction
		p.logger.Warn("app-aware-link not found, trying generic links")
		links, err = p.page.Elements("a[href*='/in/']")
		if err != nil {
			return nil, fmt.Errorf("failed to find any profile links: %w", err)
		}
	}

	p.logger.Info("found profile links", zap.Int("total_links", len(links)))

	var profiles []ProfileResult
	seen := make(map[string]bool)

	for i, link := range links {
		href, err := link.Attribute("href")
		if err != nil || href == nil {
			continue
		}

		url := cleanProfileURL(*href)

		// Validate it's a real profile URL
		if !strings.Contains(url, "/in/") || seen[url] {
			continue
		}

		// Skip if it looks like a company or other non-profile link
		if strings.Contains(url, "/company/") || strings.Contains(url, "/school/") {
			continue
		}

		seen[url] = true

		// Try to extract name from the link's child span
		name := ""
		nameElement, err := link.Element("span.entity-result__title-text")
		if err == nil {
			name, _ = nameElement.Text()
		}

		// If no name found, try link text directly
		if name == "" {
			name, _ = link.Text()
		}

		name = cleanName(name)

		if name != "" {
			profile := ProfileResult{
				ProfileURL: url,
				Name:       name,
			}

			profiles = append(profiles, profile)

			p.logger.Debug("extracted profile via direct link",
				zap.Int("index", i),
				zap.String("url", url),
				zap.String("name", name))
		}
	}

	p.logger.Info("direct parsing results",
		zap.Int("total_links", len(links)),
		zap.Int("valid_profiles", len(profiles)))

	return profiles, nil
}

// parseProfileCard extracts data from a single profile card
func (p *Parser) parseProfileCard(card *rod.Element) (*ProfileResult, error) {
	profile := &ProfileResult{}

	// Extract profile URL
	linkElement, err := card.Element("a.app-aware-link[href*='/in/']")
	if err != nil {
		// Try without class
		linkElement, err = card.Element("a[href*='/in/']")
		if err != nil {
			return nil, fmt.Errorf("no profile link found")
		}
	}

	href, err := linkElement.Attribute("href")
	if err != nil || href == nil {
		return nil, fmt.Errorf("failed to get href")
	}
	profile.ProfileURL = cleanProfileURL(*href)

	// Extract name
	nameElement, err := card.Element("span.entity-result__title-text")
	if err == nil {
		name, _ := nameElement.Text()
		profile.Name = cleanName(name)
	}

	// If name is empty, try alternative selectors
	if profile.Name == "" {
		altSelectors := []string{
			".entity-result__title-text span[aria-hidden='true']",
			".entity-result__title-text span",
			"a[href*='/in/'] span",
		}
		for _, selector := range altSelectors {
			el, err := card.Element(selector)
			if err == nil {
				text, _ := el.Text()
				cleanedText := cleanName(text)
				if cleanedText != "" {
					profile.Name = cleanedText
					break
				}
			}
		}
	}

	// Extract headline
	headlineElement, err := card.Element(".entity-result__primary-subtitle")
	if err == nil {
		headline, _ := headlineElement.Text()
		profile.Headline = strings.TrimSpace(headline)
	}

	// Extract location
	locationElement, err := card.Element(".entity-result__secondary-subtitle")
	if err == nil {
		location, _ := locationElement.Text()
		profile.Location = strings.TrimSpace(location)
	}

	// Check connection status
	profile.IsConnected = p.checkConnectionStatus(card)
	profile.ConnectionDegree = p.getConnectionDegree(card)

	return profile, nil
}

// checkConnectionStatus checks if already connected to this profile
func (p *Parser) checkConnectionStatus(card *rod.Element) bool {
	connectedBadge, err := card.Timeout(500 * time.Millisecond).Element(".artdeco-button--tertiary")
	if err == nil {
		text, _ := connectedBadge.Text()
		if strings.Contains(strings.ToLower(text), "message") {
			return true
		}
	}
	return false
}

// getConnectionDegree extracts connection degree (1st, 2nd, 3rd+)
func (p *Parser) getConnectionDegree(card *rod.Element) string {
	degreeElement, err := card.Timeout(500 * time.Millisecond).Element(".dist-value")
	if err == nil {
		degree, _ := degreeElement.Text()
		return strings.TrimSpace(degree)
	}

	badgeElement, err := card.Timeout(500 * time.Millisecond).Element(".entity-result__badge-text")
	if err == nil {
		badge, _ := badgeElement.Text()
		return strings.TrimSpace(badge)
	}

	return "Unknown"
}

// GetResultCount returns the total number of search results
func (p *Parser) GetResultCount() (int, error) {
	countElement, err := p.page.Timeout(2 * time.Second).Element(".search-results-container h2")
	if err != nil {
		return 0, fmt.Errorf("failed to find result count: %w", err)
	}

	text, err := countElement.Text()
	if err != nil {
		return 0, err
	}

	p.logger.Debug("result count text", zap.String("text", text))
	return parseResultCount(text), nil
}

// FilterConnectable filters profiles that can receive connection requests
func (p *Parser) FilterConnectable(profiles []ProfileResult) []ProfileResult {
	var connectable []ProfileResult

	for _, profile := range profiles {
		if !profile.IsConnected {
			connectable = append(connectable, profile)
		}
	}

	p.logger.Debug("filtered connectable profiles",
		zap.Int("total", len(profiles)),
		zap.Int("connectable", len(connectable)))

	return connectable
}

// DeduplicateProfiles removes duplicate profiles based on URL
func (p *Parser) DeduplicateProfiles(profiles []ProfileResult) []ProfileResult {
	seen := make(map[string]bool)
	var unique []ProfileResult

	for _, profile := range profiles {
		if !seen[profile.ProfileURL] {
			seen[profile.ProfileURL] = true
			unique = append(unique, profile)
		}
	}

	if len(unique) < len(profiles) {
		p.logger.Debug("removed duplicates",
			zap.Int("original", len(profiles)),
			zap.Int("unique", len(unique)))
	}

	return unique
}

// cleanProfileURL removes query parameters from profile URL
func cleanProfileURL(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
}

// parseResultCount extracts number from result count text
func parseResultCount(text string) int {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ",", "")

	var numStr string
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else if len(numStr) > 0 {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	var num int
	fmt.Sscanf(numStr, "%d", &num)
	return num
}

// cleanName removes junk text from profile names
func cleanName(name string) string {
	name = strings.TrimSpace(name)

	// Split by newlines and take only the first non-empty line
	// This handles cases like "Rakesh V\nRakesh V's profile"
	lines := strings.Split(name, "\n")
	if len(lines) > 0 {
		name = strings.TrimSpace(lines[0])
	}

	// Remove common junk strings
	junkPhrases := []string{
		"View ",
		"'s profile",
		"Status is offline",
		"Status is online",
		"Status is reachable",
		"Open to work",
		"LinkedIn Member",
	}

	for _, junk := range junkPhrases {
		name = strings.ReplaceAll(name, junk, "")
	}

	// Clean up any double spaces
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}

	return strings.TrimSpace(name)
}
