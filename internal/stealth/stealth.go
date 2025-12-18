package stealth

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

type Stealth struct {
	cfg         *config.Config
	log         *logrus.Logger
	actionCount int
}

func New(cfg *config.Config) *Stealth {
	return &Stealth{
		cfg: cfg,
		log: logger.Get(),
	}
}

// ApplyBrowserStealth applies stealth techniques to the browser
func (s *Stealth) ApplyBrowserStealth(page *rod.Page) error {
	// Technique 1: Disable navigator.webdriver
	_, err := page.Eval(`() => {
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});
	}`)
	if err != nil {
		return fmt.Errorf("failed to disable webdriver: %w", err)
	}

	// Technique 2: Override Chrome detection
	_, err = page.Eval(`() => {
		window.chrome = {
			runtime: {}
		};
	}`)
	if err != nil {
		return fmt.Errorf("failed to set chrome object: %w", err)
	}

	// Technique 3: Override permissions
	_, err = page.Eval(`() => {
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	}`)
	if err != nil {
		return fmt.Errorf("failed to override permissions: %w", err)
	}

	// Technique 4: Override plugins
	_, err = page.Eval(`() => {
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5]
		});
	}`)
	if err != nil {
		return fmt.Errorf("failed to override plugins: %w", err)
	}

	// Technique 5: Override languages
	_, err = page.Eval(`() => {
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});
	}`)
	if err != nil {
		return fmt.Errorf("failed to override languages: %w", err)
	}

	s.log.Info("Browser stealth techniques applied")
	return nil
}

// RandomDelay introduces a random delay based on configuration
func (s *Stealth) RandomDelay(delayType string) {
	var min, max int

	switch delayType {
	case "action":
		min = s.cfg.Stealth.ActionDelay.Min
		max = s.cfg.Stealth.ActionDelay.Max
	case "scroll":
		min = s.cfg.Stealth.ScrollDelay.Min
		max = s.cfg.Stealth.ScrollDelay.Max
	case "typing":
		min = s.cfg.Stealth.TypingDelay.Min
		max = s.cfg.Stealth.TypingDelay.Max
	case "think":
		min = s.cfg.Stealth.ThinkTime.Min
		max = s.cfg.Stealth.ThinkTime.Max
	default:
		min = 1000
		max = 3000
	}

	delay := min + rand.Intn(max-min+1)
	s.log.Debugf("Random %s delay: %dms", delayType, delay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// HumanMouseMove moves the mouse in a human-like way using Bezier curves
// Technique 6: Bezier curve mouse movement
func (s *Stealth) HumanMouseMove(page *rod.Page, targetX, targetY float64) error {
	if !s.cfg.Stealth.EnableMouseMovement {
		return nil
	}

	// Get current mouse position (start from random position if first move)
	startX := rand.Float64() * float64(s.cfg.Browser.Viewport.Width)
	startY := rand.Float64() * float64(s.cfg.Browser.Viewport.Height)

	// Generate control points for Bezier curve
	cp1X := startX + (targetX-startX)*0.25 + (rand.Float64()-0.5)*100
	cp1Y := startY + (targetY-startY)*0.25 + (rand.Float64()-0.5)*100
	cp2X := startX + (targetX-startX)*0.75 + (rand.Float64()-0.5)*100
	cp2Y := startY + (targetY-startY)*0.75 + (rand.Float64()-0.5)*100

	// Number of steps for smooth movement
	steps := 20 + rand.Intn(10)

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)

		// Cubic Bezier curve formula
		x := math.Pow(1-t, 3)*startX +
			3*math.Pow(1-t, 2)*t*cp1X +
			3*(1-t)*math.Pow(t, 2)*cp2X +
			math.Pow(t, 3)*targetX

		y := math.Pow(1-t, 3)*startY +
			3*math.Pow(1-t, 2)*t*cp1Y +
			3*(1-t)*math.Pow(t, 2)*cp2Y +
			math.Pow(t, 3)*targetY

		page.Mouse.MoveLinear(proto.Point{X: x, Y: y}, 1)
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
	}

	s.log.Debugf("Human mouse move to (%.0f, %.0f)", targetX, targetY)
	return nil
}

// HumanClick performs a human-like click with movement and delay
func (s *Stealth) HumanClick(element *rod.Element) error {
	// Move mouse to element with Bezier curve
	box, err := element.Shape()
	if err != nil {
		return fmt.Errorf("failed to get element shape: %w", err)
	}

	if len(box.Quads) == 0 {
		return fmt.Errorf("element has no quads")
	}

	// Get center of element with slight randomization
	centerX := (box.Quads[0][0] + box.Quads[0][2]) / 2
	centerY := (box.Quads[0][1] + box.Quads[0][5]) / 2

	// Add small random offset
	centerX += (rand.Float64() - 0.5) * 10
	centerY += (rand.Float64() - 0.5) * 10

	page := element.Page()
	s.HumanMouseMove(page, centerX, centerY)

	// Small delay before click
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	// Click
	if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click: %w", err)
	}

	s.log.Debug("Human click performed")
	s.actionCount++

	// Check if we should take an idle break
	s.MaybeIdleBreak()

	return nil
}

// HumanType types text in a human-like way with random delays and occasional mistakes
// Technique 7: Human typing simulation with mistakes
func (s *Stealth) HumanType(element *rod.Element, text string) error {
	if !s.cfg.Stealth.EnableHumanTyping {
		return element.Input(text)
	}

	// Click on element first
	if err := s.HumanClick(element); err != nil {
		return err
	}

	// Type each character with random delay
	for i, char := range text {
		// 5% chance of making a typo
		if rand.Float64() < 0.05 && i < len(text)-1 {
			// Type wrong character
			wrongChar := rune('a' + rand.Intn(26))
			element.Page().Keyboard.Type(input.Key(wrongChar))
			time.Sleep(time.Duration(s.cfg.Stealth.TypingDelay.Min+rand.Intn(s.cfg.Stealth.TypingDelay.Max-s.cfg.Stealth.TypingDelay.Min)) * time.Millisecond)

			// Backspace
			element.Page().Keyboard.Press(input.Backspace)
			time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
		}

		// Type correct character
		element.Page().Keyboard.Type(input.Key(char))

		// Random delay between keystrokes
		delay := s.cfg.Stealth.TypingDelay.Min + rand.Intn(s.cfg.Stealth.TypingDelay.Max-s.cfg.Stealth.TypingDelay.Min)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	s.log.Debugf("Human typed: %s", text)
	return nil
}

// RandomScroll performs random scrolling on the page
// Technique 8: Random scrolling behavior
func (s *Stealth) RandomScroll(page *rod.Page) error {
	if !s.cfg.Stealth.EnableRandomScrolling {
		return nil
	}

	// Random number of scroll actions
	scrolls := 2 + rand.Intn(4)

	for i := 0; i < scrolls; i++ {
		// Random scroll distance (can be up or down)
		distance := 200 + rand.Intn(400)
		if rand.Float64() < 0.3 {
			distance = -distance // Scroll up sometimes
		}

		// Smooth scroll
		page.Mouse.Scroll(0, float64(distance), 10)

		s.RandomDelay("scroll")
	}

	s.log.Debug("Random scrolling performed")
	return nil
}

// MouseHover performs random mouse hovering
// Technique 9: Mouse hovering and wandering
func (s *Stealth) MouseHover(page *rod.Page) error {
	if !s.cfg.Stealth.EnableMouseHovering {
		return nil
	}

	// Random position on page
	x := rand.Float64() * float64(s.cfg.Browser.Viewport.Width)
	y := rand.Float64() * float64(s.cfg.Browser.Viewport.Height)

	s.HumanMouseMove(page, x, y)

	// Hover for a bit
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

	s.log.Debug("Mouse hover performed")
	return nil
}

// MaybeIdleBreak takes an idle break if the action count threshold is reached
// Technique 10: Idle breaks and cool-down periods
func (s *Stealth) MaybeIdleBreak() {
	if !s.cfg.Stealth.EnableIdleBreaks {
		return
	}

	if s.actionCount >= s.cfg.Stealth.IdleBreak.FrequencyActions {
		duration := s.cfg.Stealth.IdleBreak.MinDurationSeconds +
			rand.Intn(s.cfg.Stealth.IdleBreak.MaxDurationSeconds-s.cfg.Stealth.IdleBreak.MinDurationSeconds)

		s.log.Infof("Taking idle break for %d seconds", duration)
		time.Sleep(time.Duration(duration) * time.Second)

		s.actionCount = 0
	}
}

// WaitForElement waits for an element with human-like behavior
func (s *Stealth) WaitForElement(page *rod.Page, selector string, timeout time.Duration) (*rod.Element, error) {
	// Add some think time before searching
	s.RandomDelay("think")

	element, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return nil, fmt.Errorf("element not found: %w", err)
	}

	// Small delay after finding element
	time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

	return element, nil
}

// SimulateReading simulates reading content on the page
// Technique 11: Reading simulation
func (s *Stealth) SimulateReading(page *rod.Page) {
	// Scroll slowly down the page as if reading
	scrollSteps := 3 + rand.Intn(3)

	for i := 0; i < scrollSteps; i++ {
		page.Mouse.Scroll(0, float64(100+rand.Intn(200)), 5)
		time.Sleep(time.Duration(2000+rand.Intn(3000)) * time.Millisecond)
	}

	s.log.Debug("Reading simulation performed")
}
