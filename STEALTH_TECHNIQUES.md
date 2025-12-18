# Stealth Techniques - Implementation Guide

This document provides detailed explanations of all stealth techniques implemented in this project, including the **why** and **how** of each approach.

## Table of Contents

1. [Browser Fingerprint Masking](#browser-fingerprint-masking)
2. [Human-Like Mouse Movement](#human-like-mouse-movement)
3. [Human Typing Simulation](#human-typing-simulation)
4. [Random Scrolling](#random-scrolling)
5. [Mouse Hovering](#mouse-hovering)
6. [Idle Breaks](#idle-breaks)
7. [Reading Simulation](#reading-simulation)
8. [Timing Randomization](#timing-randomization)
9. [Viewport Randomization](#viewport-randomization)
10. [User Agent Rotation](#user-agent-rotation)
11. [Business Hours Operation](#business-hours-operation)

---

## 1. Browser Fingerprint Masking

### Why?
Websites can detect automated browsers through various JavaScript properties that differ between normal and automated Chrome instances.

### How?

#### Technique 1.1: Navigator.webdriver
```javascript
Object.defineProperty(navigator, 'webdriver', {
    get: () => undefined
});
```
**Detection**: `navigator.webdriver` returns `true` in automated browsers.
**Solution**: Override to return `undefined` like normal browsers.

#### Technique 1.2: Chrome Object
```javascript
window.chrome = {
    runtime: {}
};
```
**Detection**: Headless Chrome lacks `window.chrome` object.
**Solution**: Add the object to mimic normal Chrome.

#### Technique 1.3: Permissions API
```javascript
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
        Promise.resolve({ state: Notification.permission }) :
        originalQuery(parameters)
);
```
**Detection**: Permissions API behaves differently in headless mode.
**Solution**: Override to return realistic values.

#### Technique 1.4: Plugins
```javascript
Object.defineProperty(navigator, 'plugins', {
    get: () => [1, 2, 3, 4, 5]
});
```
**Detection**: Automated browsers often have zero plugins.
**Solution**: Return a non-empty array.

#### Technique 1.5: Languages
```javascript
Object.defineProperty(navigator, 'languages', {
    get: () => ['en-US', 'en']
});
```
**Detection**: Language preferences can reveal automation.
**Solution**: Set realistic language array.

### Implementation Location
`internal/stealth/stealth.go` - `ApplyBrowserStealth()` method

### Effectiveness
⭐⭐⭐⭐⭐ (Essential baseline)

---

## 2. Human-Like Mouse Movement

### Why?
Automated tools typically move the mouse in straight lines at constant speed. Humans move in curves with variable acceleration.

### How?

#### Bezier Curve Algorithm
```go
// Cubic Bezier curve: B(t) = (1-t)³P₀ + 3(1-t)²tP₁ + 3(1-t)t²P₂ + t³P₃
x := math.Pow(1-t, 3)*startX +
    3*math.Pow(1-t, 2)*t*cp1X +
    3*(1-t)*math.Pow(t, 2)*cp2X +
    math.Pow(t, 3)*targetX
```

**Parameters**:
- `P₀` (startX, startY): Current position
- `P₁` (cp1X, cp1Y): First control point (random offset)
- `P₂` (cp2X, cp2Y): Second control point (random offset)
- `P₃` (targetX, targetY): Destination

**Control Points**:
```go
cp1X := startX + (targetX-startX)*0.25 + (rand.Float64()-0.5)*100
cp1Y := startY + (targetY-startY)*0.25 + (rand.Float64()-0.5)*100
```
Random offset of ±50px adds natural variation.

**Steps**:
- 20-30 steps per movement (randomized)
- 10-30ms delay per step
- Creates smooth, curved path

### Visual Comparison

**Automated (Linear)**:
```
Start ────────────────────→ End
```

**Human (Bezier)**:
```
Start ╭─────╮
      │     ╰──╮
      │        ╰─→ End
```

### Implementation Location
`internal/stealth/stealth.go` - `HumanMouseMove()` method

### Effectiveness
⭐⭐⭐⭐⭐ (Critical for realism)

---

## 3. Human Typing Simulation

### Why?
Bots type at constant speed with no errors. Humans have variable speed and make mistakes.

### How?

#### Variable Keystroke Timing
```go
delay := cfg.TypingDelay.Min + rand.Intn(cfg.TypingDelay.Max - cfg.TypingDelay.Min)
time.Sleep(time.Duration(delay) * time.Millisecond)
```
**Range**: 100-300ms per keystroke (configurable)

#### Typo Simulation
```go
if rand.Float64() < 0.05 {  // 5% chance
    // Type wrong character
    wrongChar := rune('a' + rand.Intn(26))
    page.Keyboard.Type(input.Key(wrongChar))
    time.Sleep(...)
    
    // Backspace to correct
    page.Keyboard.Press(input.Backspace)
    time.Sleep(...)
}
```

**Behavior**:
- 5% of characters are initially wrong
- Immediate backspace correction
- Adds 200-400ms total delay
- Highly realistic

### Example Timeline
```
Type 'H' → 150ms → Type 'e' → 200ms → Type 'x' (typo) → 100ms → 
Backspace → 150ms → Type 'l' → 180ms → Type 'l' → 220ms → Type 'o'
```

### Implementation Location
`internal/stealth/stealth.go` - `HumanType()` method

### Effectiveness
⭐⭐⭐⭐⭐ (Highly realistic)

---

## 4. Random Scrolling

### Why?
Humans scroll to explore content. Bots often jump directly to elements without scrolling.

### How?

#### Scroll Pattern
```go
scrolls := 2 + rand.Intn(4)  // 2-5 scrolls

for i := 0; i < scrolls; i++ {
    distance := 200 + rand.Intn(400)  // 200-600px
    
    if rand.Float64() < 0.3 {  // 30% chance
        distance = -distance  // Scroll up
    }
    
    page.Mouse.Scroll(0, float64(distance), 10)
    RandomDelay("scroll")  // 1000-3000ms
}
```

**Characteristics**:
- Multiple scroll actions
- Variable distances
- Bidirectional (mostly down, sometimes up)
- Random delays between scrolls

### Scroll Behavior
```
Page Load
    ↓
Scroll Down 300px → Wait 1.5s
    ↓
Scroll Down 450px → Wait 2.2s
    ↓
Scroll Up 200px → Wait 1.8s
    ↓
Scroll Down 350px → Wait 2.5s
```

### Implementation Location
`internal/stealth/stealth.go` - `RandomScroll()` method

### Effectiveness
⭐⭐⭐⭐ (Adds realism)

---

## 5. Mouse Hovering

### Why?
Humans move their mouse around while reading/thinking. Bots keep it stationary or only move when clicking.

### How?

```go
func MouseHover(page *rod.Page) error {
    // Random position on page
    x := rand.Float64() * float64(cfg.Browser.Viewport.Width)
    y := rand.Float64() * float64(cfg.Browser.Viewport.Height)
    
    // Move with Bezier curve
    HumanMouseMove(page, x, y)
    
    // Hover for a bit
    time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
}
```

**Behavior**:
- Moves to random screen position
- Hovers for 500-2000ms
- Uses Bezier curve movement
- Called periodically during workflow

### Implementation Location
`internal/stealth/stealth.go` - `MouseHover()` method

### Effectiveness
⭐⭐⭐ (Subtle but helpful)

---

## 6. Idle Breaks

### Why?
Humans take breaks. Bots perform actions continuously without pause.

### How?

```go
func MaybeIdleBreak() {
    if actionCount >= cfg.Stealth.IdleBreak.FrequencyActions {
        duration := cfg.Stealth.IdleBreak.MinDurationSeconds + 
            rand.Intn(cfg.Stealth.IdleBreak.MaxDurationSeconds - 
                     cfg.Stealth.IdleBreak.MinDurationSeconds)
        
        log.Infof("Taking idle break for %d seconds", duration)
        time.Sleep(time.Duration(duration) * time.Second)
        
        actionCount = 0
    }
}
```

**Configuration**:
- Frequency: Every 20 actions (default)
- Duration: 60-180 seconds (1-3 minutes)
- Resets action counter

**Example**:
```
Action 1-19: Normal operation
Action 20: Trigger break → Sleep 125 seconds
Action 21-39: Normal operation
Action 40: Trigger break → Sleep 95 seconds
...
```

### Implementation Location
`internal/stealth/stealth.go` - `MaybeIdleBreak()` method

### Effectiveness
⭐⭐⭐⭐⭐ (Critical for long sessions)

---

## 7. Reading Simulation

### Why?
Humans read content before acting. Bots click immediately after page load.

### How?

```go
func SimulateReading(page *rod.Page) {
    scrollSteps := 3 + rand.Intn(3)  // 3-5 steps
    
    for i := 0; i < scrollSteps; i++ {
        // Small scroll
        page.Mouse.Scroll(0, float64(100+rand.Intn(200)), 5)
        
        // "Read" for 2-5 seconds
        time.Sleep(time.Duration(2000+rand.Intn(3000)) * time.Millisecond)
    }
}
```

**Behavior**:
- 3-5 small scrolls
- 100-300px per scroll
- 2-5 second pause between scrolls
- Total reading time: 6-25 seconds

**When Used**:
- After navigating to profile
- Before clicking Connect
- Before sending message

### Implementation Location
`internal/stealth/stealth.go` - `SimulateReading()` method

### Effectiveness
⭐⭐⭐⭐ (Important for profile interactions)

---

## 8. Timing Randomization

### Why?
Fixed delays create detectable patterns. Humans have variable timing.

### How?

#### Delay Types
```go
type DelayConfig struct {
    Min int  // Minimum delay in milliseconds
    Max int  // Maximum delay in milliseconds
}
```

**Configured Delays**:
- **Action Delay**: 2000-5000ms (between major actions)
- **Scroll Delay**: 1000-3000ms (between scrolls)
- **Typing Delay**: 100-300ms (between keystrokes)
- **Think Time**: 3000-8000ms (before important decisions)

#### Implementation
```go
func RandomDelay(delayType string) {
    delay := min + rand.Intn(max - min + 1)
    time.Sleep(time.Duration(delay) * time.Millisecond)
}
```

**Usage Throughout Codebase**:
```go
stealth.RandomDelay("action")  // Before clicking
stealth.RandomDelay("think")   // Before typing
stealth.RandomDelay("scroll")  // Between scrolls
```

### Effectiveness
⭐⭐⭐⭐⭐ (Essential foundation)

---

## 9. Viewport Randomization

### Why?
Bots often use fixed viewport sizes. Real users have varied screen resolutions.

### How?

#### Configuration
```yaml
browser:
  viewport:
    width: 1920
    height: 1080
```

**Can be randomized**:
```go
viewportSizes := []struct{ width, height int }{
    {1920, 1080},  // Full HD
    {1366, 768},   // Common laptop
    {1440, 900},   // MacBook
    {2560, 1440},  // 2K
}

size := viewportSizes[rand.Intn(len(viewportSizes))]
```

### Implementation Location
`internal/browser/browser.go` - `New()` method

### Effectiveness
⭐⭐⭐ (Adds variation)

---

## 10. User Agent Rotation

### Why?
Same user agent on every request is suspicious. Real users have different browsers/versions.

### How?

#### Configuration
```yaml
browser:
  user_agents:
    - "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    - "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
    - "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
```

#### Selection
```go
userAgent := cfg.Browser.UserAgents[rand.Intn(len(cfg.Browser.UserAgents))]
page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
    UserAgent: userAgent,
})
```

**Best Practice**: Use recent, common user agents
- Chrome 119-120 (current versions)
- Windows 10 and macOS
- Real user agent strings

### Implementation Location
`internal/browser/browser.go` - `New()` method

### Effectiveness
⭐⭐⭐⭐ (Important for fingerprinting)

---

## 11. Business Hours Operation

### Why?
Activity at 3 AM is suspicious. Humans work during business hours.

### How?

#### Configuration
```yaml
scheduling:
  active_hours:
    start: 9   # 9 AM
    end: 18    # 6 PM
  active_days:
    - monday
    - tuesday
    - wednesday
    - thursday
    - friday
```

#### Enforcement
```go
func ShouldRun() bool {
    now := time.Now()
    
    if !isActiveDay(now) {
        return false
    }
    
    if !isActiveHour(now) {
        return false
    }
    
    return true
}
```

**Behavior**:
- Only operates Monday-Friday, 9 AM - 6 PM
- Automatically waits until next active period
- Respects configured timezone

### Implementation Location
`internal/scheduler/scheduler.go`

### Effectiveness
⭐⭐⭐⭐⭐ (Critical for avoiding detection)

---

## Combined Effectiveness

When all techniques are used together:

### Detection Difficulty: ⭐⭐⭐⭐⭐

**Why?**
1. **No single red flag**: Each technique addresses a different detection vector
2. **Layered defense**: Multiple techniques reinforce each other
3. **Realistic behavior**: Mimics actual human patterns
4. **Randomization**: No predictable patterns
5. **Time-based**: Respects human work schedules

### Comparison

| Aspect | Basic Bot | This Implementation |
|--------|-----------|---------------------|
| Mouse Movement | Linear | Bezier curves |
| Typing Speed | Constant | Variable with typos |
| Scrolling | None/Fixed | Random, bidirectional |
| Timing | Fixed delays | Randomized ranges |
| Activity Pattern | 24/7 | Business hours only |
| Breaks | None | Regular idle periods |
| Fingerprint | Detectable | Masked |
| User Agent | Fixed | Rotating |

---

## Best Practices

### 1. Conservative Configuration
Start with conservative limits:
- Low daily limits (20-30 connections)
- Long delays (3-8 seconds)
- Frequent breaks (every 15-20 actions)

### 2. Monitor and Adjust
- Check logs for patterns
- Review screenshots on errors
- Adjust timing if needed

### 3. Respect the Platform
- Don't abuse the techniques
- Use for learning only
- Respect rate limits

### 4. Stay Updated
- LinkedIn's detection evolves
- Update user agents regularly
- Monitor for new detection methods

---

## Conclusion

These 11 stealth techniques work together to create highly realistic automation that's difficult to distinguish from human behavior. The key is **randomization** and **human-like patterns** at every level.

**Remember**: This is for educational purposes only. Always respect platform terms of service.
