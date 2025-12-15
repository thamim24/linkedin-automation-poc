package stealth

import (
	"math"
	"time"

	"linkedin-automation-poc/internal/config"
)

// Timing manages delays and timing patterns
type Timing struct {
	config        config.TimingConfig
	actionCount   int
	fatigueActive bool
}

// NewTiming creates a new timing controller
func NewTiming(cfg config.TimingConfig) *Timing {
	return &Timing{
		config:        cfg,
		actionCount:   0,
		fatigueActive: false,
	}
}

// ActionDelay returns a human-like delay between actions
func (t *Timing) ActionDelay() time.Duration {
	// Log-normal distribution for more realistic delays
	delay := t.logNormalDelay(t.config.MinActionDelay, t.config.MaxActionDelay)
	
	// Apply fatigue multiplier if threshold reached
	if t.actionCount >= t.config.FatigueAfter {
		if !t.fatigueActive {
			t.fatigueActive = true
		}
		delay = time.Duration(float64(delay) * t.config.FatigueMultiplier)
	}
	
	t.actionCount++
	return delay
}

// PageLoadDelay returns delay for page loading
func (t *Timing) PageLoadDelay() time.Duration {
	// Longer base delay for page loads
	min := 1 * time.Second
	max := 3 * time.Second
	return t.logNormalDelay(min, max)
}

// ShortPause returns a short pause duration
func (t *Timing) ShortPause() time.Duration {
	min := 200 * time.Millisecond
	max := 800 * time.Millisecond
	return t.logNormalDelay(min, max)
}

// MediumPause returns a medium pause duration
func (t *Timing) MediumPause() time.Duration {
	min := 1 * time.Second
	max := 3 * time.Second
	return t.logNormalDelay(min, max)
}

// LongPause returns a long pause duration
func (t *Timing) LongPause() time.Duration {
	min := 3 * time.Second
	max := 8 * time.Second
	return t.logNormalDelay(min, max)
}

// ReadingTime simulates time spent reading content
func (t *Timing) ReadingTime(wordCount int) time.Duration {
	// Average reading speed: 200-250 words per minute
	wordsPerSecond := 3.5
	baseTime := float64(wordCount) / wordsPerSecond
	
	// Add variance
	variance := randomFloat(0.8, 1.2)
	seconds := baseTime * variance
	
	// Minimum 2 seconds, maximum 30 seconds
	if seconds < 2 {
		seconds = 2
	}
	if seconds > 30 {
		seconds = 30
	}
	
	return time.Duration(seconds * float64(time.Second))
}

// ThinkingDelay simulates user thinking time
func (t *Timing) ThinkingDelay() time.Duration {
	// Random thinking time between 1-5 seconds
	min := 1 * time.Second
	max := 5 * time.Second
	return t.logNormalDelay(min, max)
}

// ResetFatigue resets the fatigue counter
func (t *Timing) ResetFatigue() {
	t.actionCount = 0
	t.fatigueActive = false
}

// IncrementAction increments the action counter
func (t *Timing) IncrementAction() {
	t.actionCount++
}

// GetActionCount returns current action count
func (t *Timing) GetActionCount() int {
	return t.actionCount
}

// IsFatigued returns whether fatigue is active
func (t *Timing) IsFatigued() bool {
	return t.fatigueActive
}

// logNormalDelay generates a delay using log-normal distribution
func (t *Timing) logNormalDelay(min, max time.Duration) time.Duration {
	// Log-normal distribution is more realistic than uniform
	// Most delays are near the minimum, with occasional longer delays
	
	minMs := float64(min.Milliseconds())
	maxMs := float64(max.Milliseconds())
	
	// Calculate mu and sigma for log-normal distribution
	mean := (minMs + maxMs) / 2
	stdDev := (maxMs - minMs) / 4
	
	// Generate log-normal value
	mu := math.Log(mean*mean / math.Sqrt(mean*mean+stdDev*stdDev))
	sigma := math.Sqrt(math.Log(1 + (stdDev*stdDev)/(mean*mean)))
	
	// Box-Muller transform for normal distribution
	u1 := randomFloat(0, 1)
	u2 := randomFloat(0, 1)
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	
	// Convert to log-normal
	value := math.Exp(mu + sigma*z)
	
	// Clamp to bounds
	if value < minMs {
		value = minMs
	}
	if value > maxMs {
		value = maxMs
	}
	
	return time.Duration(value) * time.Millisecond
}

// Wait sleeps for the specified duration
func (t *Timing) Wait(duration time.Duration) {
	time.Sleep(duration)
}

// RandomDelay generates a random delay between min and max
func RandomDelay(min, max time.Duration) time.Duration {
	minMs := int(min.Milliseconds())
	maxMs := int(max.Milliseconds())
	
	delayMs, _ := randomInt(minMs, maxMs)
	return time.Duration(delayMs) * time.Millisecond
}