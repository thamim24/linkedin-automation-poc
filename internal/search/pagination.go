package search

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/stealth"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

// Pagination handles navigating through search results pages
type Pagination struct {
	page        *rod.Page
	currentPage int
	totalPages  int
	scroll      *stealth.Scroll
	logger      *zap.Logger
}

// NewPagination creates a new pagination handler
func NewPagination(page *rod.Page, scrollCfg config.ScrollConfig) *Pagination {
	return &Pagination{
		page:        page,
		currentPage: 1,
		totalPages:  0,
		scroll:      stealth.NewScroll(page, scrollCfg),
		logger:      log.Module("pagination"),
	}
}

// HasNextPage checks if there's a next page
func (p *Pagination) HasNextPage() bool {
	// Look for next button
	nextButton, err := p.page.Timeout(2 * time.Second).Element("button[aria-label='Next']")
	if err != nil {
		return false
	}

	// Check if button is disabled
	disabled, _ := nextButton.Attribute("disabled")
	return disabled == nil
}

// GoToNextPage navigates to the next search results page
func (p *Pagination) GoToNextPage() error {
	if !p.HasNextPage() {
		return fmt.Errorf("no next page available")
	}

	p.logger.Info("navigating to next page", zap.Int("current", p.currentPage))

	// Find and click next button
	nextButton, err := p.page.Element("button[aria-label='Next']")
	if err != nil {
		return fmt.Errorf("failed to find next button: %w", err)
	}

	// Scroll button into view first
	if err := p.scroll.ScrollToElement(nextButton); err != nil {
		p.logger.Warn("failed to scroll to next button", zap.Error(err))
	}

	time.Sleep(500 * time.Millisecond)

	// Click the button
	if err := nextButton.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click next button: %w", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	p.currentPage++
	p.logger.Info("moved to next page", zap.Int("page", p.currentPage))

	return nil
}

// GoToPage navigates to a specific page number
func (p *Pagination) GoToPage(pageNum int) error {
	if pageNum < 1 {
		return fmt.Errorf("invalid page number: %d", pageNum)
	}

	p.logger.Info("navigating to specific page", zap.Int("page", pageNum))

	// LinkedIn uses page parameter in URL
	currentURL := p.page.MustInfo().URL

	// Add or update page parameter
	newURL := updatePageParam(currentURL, pageNum)

	if err := p.page.Navigate(newURL); err != nil {
		return fmt.Errorf("failed to navigate to page %d: %w", pageNum, err)
	}

	if err := p.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	p.currentPage = pageNum
	return nil
}

// GetCurrentPage returns the current page number
func (p *Pagination) GetCurrentPage() int {
	return p.currentPage
}

// EstimateTotalPages attempts to estimate total pages
func (p *Pagination) EstimateTotalPages() int {
	// Try to find pagination info
	paginationText, err := p.page.Timeout(2 * time.Second).Element(".artdeco-pagination__pages")
	if err == nil {
		text, _ := paginationText.Text()
		// Parse "Page X of Y" or similar
		p.logger.Debug("pagination text", zap.String("text", text))
	}

	return p.totalPages
}

// ScrollThroughResults scrolls through current page results
func (p *Pagination) ScrollThroughResults() error {
	p.logger.Debug("scrolling through search results")

	// Scroll to load all results on current page
	if err := p.scroll.ScrollToBottom(); err != nil {
		return fmt.Errorf("failed to scroll through results: %w", err)
	}

	// Scroll back up slightly (natural behavior)
	time.Sleep(1 * time.Second)
	if err := p.scroll.ScrollUp(); err != nil {
		p.logger.Debug("failed to scroll back up", zap.Error(err))
	}

	return nil
}

// WaitForResults waits for search results to load
func (p *Pagination) WaitForResults() error {
	p.logger.Debug("waiting for search results to load")

	// Wait for results container
	_, err := p.page.Timeout(10 * time.Second).Element(".search-results-container")
	if err != nil {
		return fmt.Errorf("search results did not load: %w", err)
	}

	// Additional wait for dynamic content
	time.Sleep(2 * time.Second)

	p.logger.Debug("search results loaded")
	return nil
}

// IsLastPage checks if we're on the last page
func (p *Pagination) IsLastPage() bool {
	return !p.HasNextPage()
}

// Reset resets pagination state
func (p *Pagination) Reset() {
	p.currentPage = 1
	p.totalPages = 0
}

// updatePageParam updates or adds page parameter to URL
func updatePageParam(urlStr string, page int) string {
	if findSubstringIndex(urlStr, "&page=") != -1 {
		// Replace existing page param
		start := findSubstringIndex(urlStr, "&page=")
		end := start + 6 // len("&page=")

		// Find end of page number
		for end < len(urlStr) && urlStr[end] >= '0' && urlStr[end] <= '9' {
			end++
		}

		return urlStr[:start] + fmt.Sprintf("&page=%d", page) + urlStr[end:]
	}

	// Add page param
	if findSubstringIndex(urlStr, "?") != -1 {
		return fmt.Sprintf("%s&page=%d", urlStr, page)
	}
	return fmt.Sprintf("%s?page=%d", urlStr, page)
}

// findSubstringIndex finds the start index of a substring
func findSubstringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
