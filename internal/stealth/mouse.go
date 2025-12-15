package stealth

import (
	"math"
	"time"

	"linkedin-automation-poc/internal/config"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Mouse handles human-like mouse movements
type Mouse struct {
	config config.MouseConfig
	page   *rod.Page
}

// NewMouse creates a new mouse controller
func NewMouse(page *rod.Page, cfg config.MouseConfig) *Mouse {
	return &Mouse{
		config: cfg,
		page:   page,
	}
}

// Point represents a 2D coordinate
type Point struct {
	X float64
	Y float64
}

// MoveTo moves the mouse to target coordinates using Bézier curves
func (m *Mouse) MoveTo(targetX, targetY float64) error {
	// Get current mouse position (default to center if unknown)
	currentX, currentY := 960.0, 540.0

	// Generate Bézier curve path
	points := m.generateBezierPath(
		Point{X: currentX, Y: currentY},
		Point{X: targetX, Y: targetY},
	)

	// Move along the path with variable speed
	for i, point := range points {
		if err := m.page.Mouse.MoveLinear(proto.Point{X: point.X, Y: point.Y}, 1); err != nil {
			return err
		}

		// Variable delay between steps (faster in middle, slower at ends)
		delay := m.calculateStepDelay(i, len(points))
		time.Sleep(delay)
	}

	// Optional overshoot and correction
	if m.config.Overshoot {
		if err := m.performOvershoot(targetX, targetY); err != nil {
			return err
		}
	}

	// Small final jitter
	if m.config.JitterRadius > 0 {
		if err := m.addJitter(targetX, targetY); err != nil {
			return err
		}
	}

	return nil
}

// Click performs a human-like click at current position
func (m *Mouse) Click() error {
	// Small random delay before click
	delay, _ := randomInt(50, 150)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if err := m.page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}

	// Small delay after click
	delay, _ = randomInt(100, 250)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	return nil
}

// ClickAt moves to coordinates and clicks
func (m *Mouse) ClickAt(x, y float64) error {
	if err := m.MoveTo(x, y); err != nil {
		return err
	}
	return m.Click()
}

// Hover hovers over an element for a natural duration
func (m *Mouse) Hover(x, y float64) error {
	if err := m.MoveTo(x, y); err != nil {
		return err
	}

	// Random hover duration
	duration, _ := randomInt(500, 1500)
	time.Sleep(time.Duration(duration) * time.Millisecond)

	return nil
}

// generateBezierPath creates a curved path between two points
func (m *Mouse) generateBezierPath(start, end Point) []Point {
	// Determine number of steps
	steps, _ := randomInt(m.config.MinSteps, m.config.MaxSteps)

	// Generate control points for cubic Bézier curve
	distance := math.Sqrt(math.Pow(end.X-start.X, 2) + math.Pow(end.Y-start.Y, 2))

	// Control points create the curve
	controlPoint1 := Point{
		X: start.X + (end.X-start.X)*0.33 + randomFloat(-distance*0.2, distance*0.2),
		Y: start.Y + (end.Y-start.Y)*0.33 + randomFloat(-distance*0.2, distance*0.2),
	}

	controlPoint2 := Point{
		X: start.X + (end.X-start.X)*0.66 + randomFloat(-distance*0.2, distance*0.2),
		Y: start.Y + (end.Y-start.Y)*0.66 + randomFloat(-distance*0.2, distance*0.2),
	}

	// Calculate points along the curve
	points := make([]Point, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		points[i] = cubicBezier(t, start, controlPoint1, controlPoint2, end)
	}

	return points
}

// cubicBezier calculates a point on a cubic Bézier curve
func cubicBezier(t float64, p0, p1, p2, p3 Point) Point {
	u := 1 - t
	tt := t * t
	uu := u * u
	uuu := uu * u
	ttt := tt * t

	return Point{
		X: uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X,
		Y: uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y,
	}
}

// calculateStepDelay returns delay based on position in movement
func (m *Mouse) calculateStepDelay(step, totalSteps int) time.Duration {
	// Faster in the middle, slower at start and end (velocity curve)
	progress := float64(step) / float64(totalSteps)

	// Sine-based velocity: slow-fast-slow
	velocity := math.Sin(progress * math.Pi)

	// Map velocity to delay (inverse relationship)
	minDelay := 1.0
	maxDelay := 5.0
	delay := maxDelay - (velocity * (maxDelay - minDelay))

	return time.Duration(delay) * time.Millisecond
}

// performOvershoot simulates overshooting and correcting
func (m *Mouse) performOvershoot(targetX, targetY float64) error {
	// Overshoot by small amount
	overshootX := targetX + randomFloat(-float64(m.config.OvershootSpread), float64(m.config.OvershootSpread))
	overshootY := targetY + randomFloat(-float64(m.config.OvershootSpread), float64(m.config.OvershootSpread))

	if err := m.page.Mouse.MoveLinear(proto.Point{X: overshootX, Y: overshootY}, 1); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	// Correct back to target
	if err := m.page.Mouse.MoveLinear(proto.Point{X: targetX, Y: targetY}, 1); err != nil {
		return err
	}

	return nil
}

// addJitter adds micro-movements around target
func (m *Mouse) addJitter(x, y float64) error {
	jitterX := x + randomFloat(-float64(m.config.JitterRadius), float64(m.config.JitterRadius))
	jitterY := y + randomFloat(-float64(m.config.JitterRadius), float64(m.config.JitterRadius))

	if err := m.page.Mouse.MoveLinear(proto.Point{X: jitterX, Y: jitterY}, 1); err != nil {
		return err
	}

	return nil
}

// randomFloat generates a random float between min and max
func randomFloat(min, max float64) float64 {
	n, _ := randomInt(0, 10000)
	normalized := float64(n) / 10000.0
	return min + normalized*(max-min)
}
