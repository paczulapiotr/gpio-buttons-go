package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gpiobuttons "github.com/piotrpaczula/gpio-buttons-go"
)

func main() {
	log.Println("Starting GPIO Button Example...")
	log.Println("ℹ️ Using INTERNAL pull-up (gpio.PullUp). Your pin must support it.")
	log.Println()

	// Create a new button manager
	manager, err := gpiobuttons.NewButtonManager()
	if err != nil {
		log.Fatalf("Failed to create button manager: %v", err)
	}

	// Configure single button with counter and timestamp
	// IMPORTANT: Replace PinName with a valid chip:line or line number from your system (e.g., "gpiochip0:23" or "23")
	// Discover lines: cd tools/pin-discovery && go run main.go (or use `gpioinfo`)
	// Typical wiring: internal PullUp + ActiveLow + button to GND
	
	counter := 0
	err = manager.AddButton(gpiobuttons.ButtonConfig{
		PinName:  "23",                // example; change to your line (or "gpiochip0:23")
		Pull:     gpiobuttons.PullUp,  // Request internal pull-up (depends on kernel/pin support)
		ActiveLow: true,               // Press = GND (logical 1)
		Callback: func(pinName string) {
			counter++
			timestamp := time.Now().Format("15:04:05.000")
			log.Printf("🔵 Button pressed [#%d] on %s at %s", counter, pinName, timestamp)
		},
		DebounceTime: 50 * time.Millisecond,
	})
	if err != nil {
		log.Fatalf("Failed to add button: %v", err)
	}

	// Start monitoring buttons
	if manager.GetButtonCount() == 0 {
		log.Fatal("No buttons were successfully configured. Please check your GPIO pin names.")
	}

	log.Printf("Successfully configured %d button(s)", manager.GetButtonCount())
	
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start button manager: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Button monitoring started. Press Ctrl+C to exit...")
	log.Println("Press the button to see counter and timestamp.")

	// Wait for interrupt signal
	<-sigChan
	log.Println("\nShutting down...")

	// Stop the button manager
	manager.Stop()

	log.Println("Example program exited cleanly")
}
