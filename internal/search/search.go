package search

import (
	"fmt"
	"net/url"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/log"

	"go.uber.org/zap"
)

// Search handles LinkedIn search functionality
type Search struct {
	ctx    *browser.Context
	logger *zap.Logger
}

// SearchParams defines search criteria
type SearchParams struct {
	Keywords   string
	Location   string
	JobTitle   string
	Company    string
	School     string
	FirstName  string
	LastName   string
	Industries []string

	// Connection filters
	ConnectionLevel string // "F" for 1st, "S" for 2nd, "O" for 3rd+
}

// NewSearch creates a new search handler
func NewSearch(ctx *browser.Context) *Search {
	return &Search{
		ctx:    ctx,
		logger: log.Module("search"),
	}
}

// BuildSearchURL constructs a LinkedIn search URL from parameters
func (s *Search) BuildSearchURL(params SearchParams) (string, error) {
	baseURL := "https://www.linkedin.com/search/results/people/"

	values := url.Values{}

	// Keywords (general search)
	if params.Keywords != "" {
		values.Add("keywords", params.Keywords)
	}

	// Location
	if params.Location != "" {
		values.Add("geoUrn", params.Location)
	}

	// Job title
	if params.JobTitle != "" {
		values.Add("title", params.JobTitle)
	}

	// Company
	if params.Company != "" {
		values.Add("company", params.Company)
	}

	// School
	if params.School != "" {
		values.Add("school", params.School)
	}

	// First name
	if params.FirstName != "" {
		values.Add("firstName", params.FirstName)
	}

	// Last name
	if params.LastName != "" {
		values.Add("lastName", params.LastName)
	}

	// Connection level
	if params.ConnectionLevel != "" {
		values.Add("network", params.ConnectionLevel)
	}

	// Industries
	for _, industry := range params.Industries {
		values.Add("industry", industry)
	}

	searchURL := baseURL
	if len(values) > 0 {
		searchURL += "?" + values.Encode()
	}

	s.logger.Debug("built search URL",
		zap.String("url", searchURL),
		zap.Any("params", params),
	)

	return searchURL, nil
}

// ExecuteSearch navigates to search URL and returns results
func (s *Search) ExecuteSearch(params SearchParams) error {
	searchURL, err := s.BuildSearchURL(params)
	if err != nil {
		return fmt.Errorf("failed to build search URL: %w", err)
	}

	s.logger.Info("executing search",
		zap.String("keywords", params.Keywords),
		zap.String("location", params.Location),
	)

	if err := s.ctx.Navigate(searchURL); err != nil {
		return fmt.Errorf("failed to navigate to search: %w", err)
	}

	s.logger.Info("search executed successfully")
	return nil
}

// QuickSearch performs a simple keyword search
func (s *Search) QuickSearch(keywords string) error {
	params := SearchParams{
		Keywords: keywords,
	}
	return s.ExecuteSearch(params)
}

// SearchByJobTitle searches people by job title
func (s *Search) SearchByJobTitle(jobTitle, location string) error {
	params := SearchParams{
		JobTitle: jobTitle,
		Location: location,
	}
	return s.ExecuteSearch(params)
}

// SearchByCompany searches people working at a company
func (s *Search) SearchByCompany(company string) error {
	params := SearchParams{
		Company: company,
	}
	return s.ExecuteSearch(params)
}

// SearchConnections searches within specific connection levels
func (s *Search) SearchConnections(level string, keywords string) error {
	params := SearchParams{
		Keywords:        keywords,
		ConnectionLevel: level,
	}
	return s.ExecuteSearch(params)
}

// GetCurrentSearchURL returns the current search URL
func (s *Search) GetCurrentSearchURL() string {
	return s.ctx.Page().MustInfo().URL
}

// IsSearchResultsPage checks if we're on a search results page
func (s *Search) IsSearchResultsPage() bool {
	url := s.GetCurrentSearchURL()
	return stringContains(url, "/search/results/people")
}

// stringContains checks if string contains substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
