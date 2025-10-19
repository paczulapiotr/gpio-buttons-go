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

// ButtonCallback is a function type that gets called when a button is pressed
type ButtonCallback func(pinName string)

// ButtonConfig represents the configuration for a single button
type ButtonConfig struct {
	// PinName is the name of the GPIO pin (e.g., "GPIO1_A0", "GPIO1_A1")
	PinName string
	
	// Callback is the function to call when the button is pressed
	Callback ButtonCallback
	
	// DebounceTime is the time to wait before considering another press (default: 50ms)
	DebounceTime time.Duration
	
	// Pull configures the pin's pull resistor (default: PullUp)
	Pull gpio.Pull
}

// ButtonManager manages multiple GPIO buttons
type ButtonManager struct {
	buttons      map[string]*button
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.Mutex
	initialized  bool
}

// button represents an internal button state
type button struct {
	pin          gpio.PinIO
	config       ButtonConfig
	lastPress    time.Time
	mu           sync.Mutex
}

// NewButtonManager creates a new ButtonManager instance
func NewButtonManager() (*ButtonManager, error) {
	// Initialize periph.io host
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph.io: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &ButtonManager{
		buttons:     make(map[string]*button),
		ctx:         ctx,
		cancel:      cancel,
		initialized: true,
	}, nil
}

// AddButton adds a button configuration to the manager
func (bm *ButtonManager) AddButton(config ButtonConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.initialized {
		return fmt.Errorf("button manager not initialized")
	}

	// Set defaults
	if config.DebounceTime == 0 {
		config.DebounceTime = 50 * time.Millisecond
	}
	if config.Pull == gpio.PullNoChange {
		config.Pull = gpio.PullUp
	}

	// Look up the GPIO pin
	pin := gpioreg.ByName(config.PinName)
	if pin == nil {
		return fmt.Errorf("failed to find pin: %s", config.PinName)
	}

	// Configure the pin as input with pull-up resistor
	if err := pin.In(config.Pull, gpio.BothEdges); err != nil {
		return fmt.Errorf("failed to configure pin %s: %w", config.PinName, err)
	}

	bm.buttons[config.PinName] = &button{
		pin:    pin,
		config: config,
	}

	log.Printf("Added button on pin %s", config.PinName)
	return nil
}

// Start begins monitoring all configured buttons
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

	log.Printf("Started monitoring %d buttons", len(bm.buttons))
	return nil
}

// monitorButton monitors a single button for press events
func (bm *ButtonManager) monitorButton(btn *button) {
	defer bm.wg.Done()

	log.Printf("Monitoring button on pin %s", btn.config.PinName)

	for {
		select {
		case <-bm.ctx.Done():
			return
		default:
			// Wait for the pin to change state
			if btn.pin.WaitForEdge(100 * time.Millisecond) {
				// Read the current state
				level := btn.pin.Read()
				
				// Check if button is pressed (LOW when using pull-up resistor)
				if level == gpio.Low {
					btn.mu.Lock()
					now := time.Now()
					
					// Check debounce
					if now.Sub(btn.lastPress) > btn.config.DebounceTime {
						btn.lastPress = now
						btn.mu.Unlock()
						
						// Call the callback in a goroutine to avoid blocking
						go func(pinName string, callback ButtonCallback) {
							if callback != nil {
								callback(pinName)
							}
						}(btn.config.PinName, btn.config.Callback)
						
						log.Printf("Button pressed on pin %s", btn.config.PinName)
					} else {
						btn.mu.Unlock()
					}
				}
			}
		}
	}
}

// Stop stops monitoring all buttons and cleans up resources
func (bm *ButtonManager) Stop() {
	bm.mu.Lock()
	if bm.cancel != nil {
		bm.cancel()
	}
	bm.mu.Unlock()

	// Wait for all monitoring goroutines to finish
	bm.wg.Wait()

	// Halt all pins
	bm.mu.Lock()
	for _, btn := range bm.buttons {
		if err := btn.pin.Halt(); err != nil {
			log.Printf("Error halting pin %s: %v", btn.config.PinName, err)
		}
	}
	bm.mu.Unlock()

	log.Println("Stopped all button monitoring")
}

// GetButtonCount returns the number of configured buttons
func (bm *ButtonManager) GetButtonCount() int {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return len(bm.buttons)
}
