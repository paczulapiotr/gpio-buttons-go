package gpiobuttons

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	defaultDebounceTime = 50 * time.Millisecond
	defaultPull         = gpio.PullNoChange // No internal pull - use external resistors
)

// ButtonCallback is invoked when a button press is detected.
type ButtonCallback func(pinName string)

// ButtonConfig defines the configuration for a single GPIO button.
type ButtonConfig struct {
	PinName      string         // GPIO pin name (e.g., "GPIO1_A0")
	Callback     ButtonCallback // Function called on button press
	DebounceTime time.Duration  // Minimum time between presses (default: 50ms)
	Pull         gpio.Pull      // Pull resistor configuration (default: PullNoChange - external resistor required)
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
	pin       gpio.PinIO
	config    ButtonConfig
	lastPress time.Time
	mu        sync.Mutex
}

// NewButtonManager creates and initializes a new ButtonManager.
// It initializes the periph.io host for GPIO access.
func NewButtonManager() (*ButtonManager, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph.io: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &ButtonManager{
		buttons: make(map[string]*button),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// AddButton registers a new button with the manager.
// Must be called before Start().
// Note: External pull-up resistor (10kΩ recommended) is required for reliable operation.
func (bm *ButtonManager) AddButton(config ButtonConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Apply defaults
	if config.DebounceTime == 0 {
		config.DebounceTime = defaultDebounceTime
	}
	// Default to PullNoChange (external pull-up required)
	if config.Pull == 0 {
		config.Pull = defaultPull
	}

	// Validate and configure GPIO pin
	pin := gpioreg.ByName(config.PinName)
	if pin == nil {
		return fmt.Errorf("failed to find pin: %s", config.PinName)
	}

	// Configure pin as input with edge detection
	if err := pin.In(config.Pull, gpio.BothEdges); err != nil {
		return fmt.Errorf("failed to configure pin %s: %w", config.PinName, err)
	}

	bm.buttons[config.PinName] = &button{
		pin:    pin,
		config: config,
	}

	log.Printf("Added button on pin %s (external pull-up resistor required)", config.PinName)
	return nil
}

// Start begins monitoring all configured buttons using hardware interrupts.
// Returns an error if no buttons are configured.
func (bm *ButtonManager) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if len(bm.buttons) == 0 {
		return fmt.Errorf("no buttons configured")
	}

	for _, btn := range bm.buttons {
		bm.wg.Add(1)
		go bm.monitorButton(btn)
	}

	log.Printf("Started monitoring %d button(s)", len(bm.buttons))
	return nil
}

// monitorButton handles button press detection using GPIO hardware interrupts.
// This is event-driven, not polling-based, for maximum efficiency.
func (bm *ButtonManager) monitorButton(btn *button) {
	defer bm.wg.Done()

	log.Printf("Monitoring button on pin %s", btn.config.PinName)

	for {
		// Block until GPIO interrupt fires (edge detected)
		if !btn.pin.WaitForEdge(-1) {
			// Edge detection failed, check if we're shutting down
			if bm.isShuttingDown() {
				return
			}
			continue
		}

		// Check for shutdown signal
		if bm.isShuttingDown() {
			return
		}

		// Handle button press if debounce period has passed
		if btn.isPressed() && btn.isDebouncePassed() {
			btn.updateLastPress()
			bm.invokeCallback(btn)
		}
	}
}

// isShuttingDown checks if the context has been cancelled.
func (bm *ButtonManager) isShuttingDown() bool {
	select {
	case <-bm.ctx.Done():
		return true
	default:
		return false
	}
}

// isPressed returns true if the button is currently pressed.
// With pull-up resistors, LOW means pressed.
func (btn *button) isPressed() bool {
	return btn.pin.Read() == gpio.Low
}

// isDebouncePassed checks if enough time has passed since the last press.
func (btn *button) isDebouncePassed() bool {
	btn.mu.Lock()
	defer btn.mu.Unlock()
	return time.Since(btn.lastPress) > btn.config.DebounceTime
}

// updateLastPress records the current time as the last press time.
func (btn *button) updateLastPress() {
	btn.mu.Lock()
	btn.lastPress = time.Now()
	btn.mu.Unlock()
}

// invokeCallback executes the button's callback in a separate goroutine.
func (bm *ButtonManager) invokeCallback(btn *button) {
	log.Printf("Button pressed on pin %s", btn.config.PinName)
	
	if btn.config.Callback != nil {
		go btn.config.Callback(btn.config.PinName)
	}
}

// Stop gracefully shuts down all button monitoring and releases GPIO resources.
// It's safe to call multiple times.
func (bm *ButtonManager) Stop() {
	// Signal all monitoring goroutines to stop
	bm.mu.Lock()
	if bm.cancel != nil {
		bm.cancel()
	}
	bm.mu.Unlock()

	// Wait for all goroutines to finish
	bm.wg.Wait()

	// Release GPIO pins
	bm.haltAllPins()

	log.Println("Stopped all button monitoring")
}

// haltAllPins releases all GPIO pin resources.
func (bm *ButtonManager) haltAllPins() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, btn := range bm.buttons {
		if err := btn.pin.Halt(); err != nil {
			log.Printf("Error halting pin %s: %v", btn.config.PinName, err)
		}
	}
}

// GetButtonCount returns the number of configured buttons.
func (bm *ButtonManager) GetButtonCount() int {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return len(bm.buttons)
}
