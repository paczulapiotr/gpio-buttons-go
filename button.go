package gpiobuttons

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	gpiocdev "github.com/warthog618/go-gpiocdev"
)

const (
	defaultDebounceTime = 50 * time.Millisecond
)

// ButtonCallback is invoked when a button press is detected.
type ButtonCallback func(pinName string)

// Pull represents internal bias configuration.
type Pull int

const (
	PullNoChange Pull = iota // Leave kernel default
	PullUp                   // Request internal pull-up
	PullDown                 // Request internal pull-down
	PullDisabled             // Disable internal bias
)

// ButtonConfig defines the configuration for a single GPIO button.
// PinName accepts formats like "gpiochip0:23" or just "23" (defaults to gpiochip0).
type ButtonConfig struct {
	PinName      string         // Logical pin identifier: "gpiochipX:line" or "line"
	Callback     ButtonCallback // Function called on button press
	DebounceTime time.Duration  // Minimum time between presses (default: 50ms)
	Pull         Pull           // Internal pull configuration (requires kernel support)
	ActiveLow    bool           // Treat low level as logical 1 (typical for buttons to GND)
}

// ButtonManager manages multiple GPIO button inputs with interrupt-driven detection.
type ButtonManager struct {
	buttons map[string]*button
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
}

// button represents the internal state of a single button.
type button struct {
	line      *gpiocdev.Line
	chip      string
	offset    int
	config    ButtonConfig
	lastPress time.Time
	mu        sync.Mutex
}

// NewButtonManager creates and initializes a new ButtonManager.
func NewButtonManager() (*ButtonManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &ButtonManager{
		buttons: make(map[string]*button),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// AddButton registers a new button with the manager.
// Must be called before Start().
func (bm *ButtonManager) AddButton(config ButtonConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Apply defaults
	if config.DebounceTime == 0 {
		config.DebounceTime = defaultDebounceTime
	}

	chip, offset, err := resolveChipLine(config.PinName)
	if err != nil {
		return err
	}

	// Build request options and try fallbacks. Use event handler for rising edge (press).
	base := []gpiocdev.LineReqOption{
		gpiocdev.AsInput,
		gpiocdev.WithConsumer("gpio-buttons-go"),
		gpiocdev.WithRisingEdge,
	}
	if config.ActiveLow {
		base = append(base, gpiocdev.AsActiveLow)
	}

	// Local debounce gate in case kernel debounce isn't available/enabled
	var last time.Time
	handler := func(evt gpiocdev.LineEvent) {
		// evt.Type is RisingEdge for press due to WithRisingEdge
		now := time.Now()
		if config.DebounceTime > 0 && !last.IsZero() && now.Sub(last) <= config.DebounceTime/2 {
			// Minimal guard against event bursts; kernel debounce should handle most cases
			return
		}
		if config.Callback != nil {
			pinName := config.PinName
			if pinName == "" {
				pinName = fmt.Sprintf("%s:%d", chip, offset)
			}
			config.Callback(pinName)
		}
		last = now
	}

	// Construct option sets: full (debounce + bias), no bias, no debounce, base only
	var combos [][]gpiocdev.LineReqOption
	{
		full := append([]gpiocdev.LineReqOption{}, base...)
		full = append(full, gpiocdev.WithEventHandler(handler))
		if config.DebounceTime > 0 {
			full = append(full, gpiocdev.WithDebounce(config.DebounceTime))
		}
		if pOpt := pullOption(config.Pull); pOpt != nil {
			full = append(full, pOpt)
		}
		combos = append(combos, full)

		if pOpt := pullOption(config.Pull); pOpt != nil {
			noBias := append([]gpiocdev.LineReqOption{}, base...)
			noBias = append(noBias, gpiocdev.WithEventHandler(handler))
			if config.DebounceTime > 0 {
				noBias = append(noBias, gpiocdev.WithDebounce(config.DebounceTime))
			}
			combos = append(combos, noBias)
		}
		if config.DebounceTime > 0 {
			noDeb := append([]gpiocdev.LineReqOption{}, base...)
			noDeb = append(noDeb, gpiocdev.WithEventHandler(handler))
			if pOpt := pullOption(config.Pull); pOpt != nil {
				noDeb = append(noDeb, pOpt)
			}
			combos = append(combos, noDeb)
		}
		baseOnly := append([]gpiocdev.LineReqOption{}, base...)
		baseOnly = append(baseOnly, gpiocdev.WithEventHandler(handler))
		combos = append(combos, baseOnly)
	}

	var line *gpiocdev.Line
	var reqErr error
	for _, opts := range combos {
		line, reqErr = gpiocdev.RequestLine(chip, offset, opts...)
		if reqErr == nil {
			break
		}
	}
	if reqErr != nil {
		return fmt.Errorf("failed to request line %s:%d: %w", chip, offset, reqErr)
	}

	btn := &button{
		line:   line,
		chip:   chip,
		offset: offset,
		config: config,
	}
	key := config.PinName
	if key == "" {
		key = fmt.Sprintf("%s:%d", chip, offset)
	}
	bm.buttons[key] = btn

	log.Printf("Added button on %s:%d (ActiveLow=%v, Debounce=%s, Pull=%v)", chip, offset, config.ActiveLow, config.DebounceTime, config.Pull)
	return nil
}

// Start is kept for API compatibility; with gpiocdev we attach event handlers at AddButton.
func (bm *ButtonManager) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if len(bm.buttons) == 0 {
		return fmt.Errorf("no buttons configured")
	}
	log.Printf("Monitoring %d button(s)", len(bm.buttons))
	return nil
}

// Removed monitor goroutine: events handled by WithEventHandler.

// Stop gracefully shuts down all button monitoring and releases GPIO resources.
func (bm *ButtonManager) Stop() {
	bm.mu.Lock()
	// Close all lines (this waits for any in-flight event handler to return)
	for _, btn := range bm.buttons {
		if btn.line != nil {
			if err := btn.line.Close(); err != nil {
				log.Printf("Error closing line %s:%d: %v", btn.chip, btn.offset, err)
			}
		}
	}
	bm.mu.Unlock()

	log.Println("Stopped all button monitoring")
}

// GetButtonCount returns the number of configured buttons.
func (bm *ButtonManager) GetButtonCount() int {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return len(bm.buttons)
}

// Helpers
// Map Pull to appropriate request option for v0.9.1.
func pullOption(p Pull) gpiocdev.LineReqOption {
	switch p {
	case PullUp:
		return gpiocdev.WithPullUp
	case PullDown:
		return gpiocdev.WithPullDown
	case PullDisabled:
		return gpiocdev.WithBiasDisabled
	default:
		return nil // WithBiasAsIs is default
	}
}

// resolveChipLine parses PinName into chip and offset.
// Accepts "gpiochipX:line" or just "line" (defaults to gpiochip0).
func resolveChipLine(pinName string) (string, int, error) {
	p := strings.TrimSpace(pinName)
	if p == "" {
		return "", 0, fmt.Errorf("PinName is required; format 'gpiochipX:line' or 'line'")
	}
	if strings.Contains(p, ":") {
		parts := strings.SplitN(p, ":", 2)
		off, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid line offset %q: %w", parts[1], err)
		}
		return parts[0], off, nil
	}
	off, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, fmt.Errorf("invalid PinName %q", p)
	}
	return "gpiochip0", off, nil
}
