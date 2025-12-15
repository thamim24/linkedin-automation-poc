package stealth

import (
	"math"
	"time"

	"linkedin-automation-poc/internal/config"

	"github.com/go-rod/rod"
)

// Scroll handles human-like scrolling behavior
type Scroll struct {
	config config.ScrollConfig
	page   *rod.Page
}

// NewScroll creates a new scroll controller
func NewScroll(page *rod.Page, cfg config.ScrollConfig) *Scroll {
	return &Scroll{
		config: cfg,
		page:   page,
	}
}

// ScrollDown scrolls down the page naturally
func (s *Scroll) ScrollDown() error {
	// Random scroll amount within configured range
	amount, _ := randomInt(s.config.MinScrollAmount, s.config.MaxScrollAmount)

	return s.smoothScroll(amount)
}

// ScrollUp scrolls up the page naturally
func (s *Scroll) ScrollUp() error {
	// Random scroll amount (negative for up)
	amount, _ := randomInt(s.config.MinScrollAmount, s.config.MaxScrollAmount)

	return s.smoothScroll(-amount)
}

// ScrollToBottom scrolls to the bottom of the page with breaks
func (s *Scroll) ScrollToBottom() error {
	for {
		// Get current scroll position
		beforeHeight, err := s.getScrollHeight()
		if err != nil {
			return err
		}

		// Scroll down
		if err := s.ScrollDown(); err != nil {
			return err
		}

		// Wait for content to load
		pause := RandomDelay(s.config.PauseBetweenMin, s.config.PauseBetweenMax)
		time.Sleep(pause)

		// Check if we've reached the bottom
		afterHeight, err := s.getScrollHeight()
		if err != nil {
			return err
		}

		if beforeHeight == afterHeight {
			// No more content to load
			break
		}

		// Random chance to scroll back up slightly (reading behavior)
		if shouldScrollBack(s.config.ScrollBackChance) {
			if err := s.ScrollUp(); err != nil {
				return err
			}
			time.Sleep(RandomDelay(500*time.Millisecond, 1500*time.Millisecond))
		}
	}

	return nil
}

// ScrollToElement scrolls to bring an element into view
func (s *Scroll) ScrollToElement(element *rod.Element) error {
	// Simply scroll element into view using Rod's built-in method
	if err := element.ScrollIntoView(); err != nil {
		return err
	}

	// Add small delay for natural behavior
	time.Sleep(200 * time.Millisecond)
	return nil
}

// ScrollToElementManual scrolls to bring an element into view with custom behavior
func (s *Scroll) ScrollToElementManual(element *rod.Element) error {
	// Get element position using Eval
	result, err := element.Eval(`() => {
		const rect = this.getBoundingClientRect();
		return { y: rect.top + window.pageYOffset };
	}`)
	if err != nil {
		return err
	}

	targetY := result.Value.Get("y").Num()

	// Calculate target scroll position (center element in viewport)
	viewportHeight, _ := s.getViewportHeight()
	adjustedTargetY := targetY - float64(viewportHeight)/2

	// Get current scroll position
	currentY, err := s.getCurrentScrollY()
	if err != nil {
		return err
	}

	// Calculate scroll distance
	distance := int(adjustedTargetY - currentY)

	// Smooth scroll to position
	return s.smoothScroll(distance)
}

// smoothScroll performs a smooth scroll with momentum
func (s *Scroll) smoothScroll(totalDistance int) error {
	if totalDistance == 0 {
		return nil
	}

	steps := s.config.ScrollSteps
	if steps <= 0 {
		steps = 10
	}

	// Calculate step sizes with easing (ease-out)
	for i := 0; i < steps; i++ {
		progress := float64(i+1) / float64(steps)

		// Ease-out cubic function
		eased := 1 - math.Pow(1-progress, 3)

		targetPosition := int(float64(totalDistance) * eased)
		currentPosition := 0
		if i > 0 {
			prevProgress := float64(i) / float64(steps)
			prevEased := 1 - math.Pow(1-prevProgress, 3)
			currentPosition = int(float64(totalDistance) * prevEased)
		}

		stepDistance := targetPosition - currentPosition

		// Execute scroll
		if err := s.page.Mouse.Scroll(0, float64(stepDistance), steps); err != nil {
			return err
		}

		// Variable delay between steps
		delay := s.calculateScrollDelay(i, steps)
		time.Sleep(delay)
	}

	return nil
}

// calculateScrollDelay returns delay based on scroll progress
func (s *Scroll) calculateScrollDelay(step, totalSteps int) time.Duration {
	// Faster at start, slower at end (deceleration)
	progress := float64(step) / float64(totalSteps)

	// Inverse progress for deceleration
	speed := 1 - progress

	minDelay := 20.0
	maxDelay := 80.0
	delay := minDelay + (1-speed)*(maxDelay-minDelay)

	return time.Duration(delay) * time.Millisecond
}

// getScrollHeight returns the total scrollable height
func (s *Scroll) getScrollHeight() (int, error) {
	height, err := s.page.Eval(`() => document.documentElement.scrollHeight`)
	if err != nil {
		return 0, err
	}

	return height.Value.Int(), nil
}

// getCurrentScrollY returns current vertical scroll position
func (s *Scroll) getCurrentScrollY() (float64, error) {
	pos, err := s.page.Eval(`() => window.pageYOffset || document.documentElement.scrollTop`)
	if err != nil {
		return 0, err
	}

	return pos.Value.Num(), nil
}

// getViewportHeight returns the viewport height
func (s *Scroll) getViewportHeight() (int, error) {
	height, err := s.page.Eval(`() => window.innerHeight`)
	if err != nil {
		return 0, err
	}

	return height.Value.Int(), nil
}

// shouldScrollBack determines if scrolling back should occur
func shouldScrollBack(chance float64) bool {
	if chance <= 0 {
		return false
	}

	rand := randomFloat(0, 1)
	return rand < chance
}

// ScrollIntoView scrolls element into view with natural behavior
func (s *Scroll) ScrollIntoView(element *rod.Element) error {
	// First check if element is already visible
	visible, err := element.Visible()
	if err != nil {
		return err
	}

	if visible {
		return nil
	}

	// Scroll to element
	return s.ScrollToElement(element)
}

// RandomScroll performs a random scroll action (up or down)
func (s *Scroll) RandomScroll() error {
	direction, _ := randomInt(0, 1)

	if direction == 0 {
		return s.ScrollDown()
	}
	return s.ScrollUp()
}
