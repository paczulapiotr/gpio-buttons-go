package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gpiobuttons "github.com/piotrpaczula/gpio-buttons-go"
)

func main() {
	log.Println("Starting GPIO Button Example...")

	// Create a new button manager
	manager, err := gpiobuttons.NewButtonManager()
	if err != nil {
		log.Fatalf("Failed to create button manager: %v", err)
	}

	// Configure 4 buttons with different callbacks
	// Note: Replace these pin names with actual Radxa GPIO pin names for your board
	// Common Radxa GPIO pins: GPIO1_A0, GPIO1_A1, GPIO1_A2, GPIO1_A3, etc.
	
	// Button 1 - LED Toggle simulation
	err = manager.AddButton(gpiobuttons.ButtonConfig{
		PinName: "GPIO1_A0", // Replace with your actual pin name
		Callback: func(pinName string) {
			log.Printf("🔴 Button 1 pressed on %s - Toggling LED", pinName)
			// Add your LED toggle logic here
		},
		DebounceTime: 50 * time.Millisecond,
	})
	if err != nil {
		log.Printf("Warning: Failed to add button 1: %v", err)
	}

	// Button 2 - Counter increment
	counter := 0
	err = manager.AddButton(gpiobuttons.ButtonConfig{
		PinName: "GPIO1_A1", // Replace with your actual pin name
		Callback: func(pinName string) {
			counter++
			log.Printf("🟢 Button 2 pressed on %s - Counter: %d", pinName, counter)
		},
		DebounceTime: 50 * time.Millisecond,
	})
	if err != nil {
		log.Printf("Warning: Failed to add button 2: %v", err)
	}

	// Button 3 - Print timestamp
	err = manager.AddButton(gpiobuttons.ButtonConfig{
		PinName: "GPIO1_A2", // Replace with your actual pin name
		Callback: func(pinName string) {
			timestamp := time.Now().Format("15:04:05.000")
			log.Printf("🔵 Button 3 pressed on %s at %s", pinName, timestamp)
		},
		DebounceTime: 50 * time.Millisecond,
	})
	if err != nil {
		log.Printf("Warning: Failed to add button 3: %v", err)
	}

	// Button 4 - Custom action
	err = manager.AddButton(gpiobuttons.ButtonConfig{
		PinName: "GPIO1_A3", // Replace with your actual pin name
		Callback: func(pinName string) {
			log.Printf("🟡 Button 4 pressed on %s - Performing custom action", pinName)
			// Add your custom action here
			fmt.Println("  ↳ Custom action executed!")
		},
		DebounceTime: 50 * time.Millisecond,
	})
	if err != nil {
		log.Printf("Warning: Failed to add button 4: %v", err)
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
	log.Println("Press any of the 4 buttons to trigger their callbacks.")

	// Wait for interrupt signal
	<-sigChan
	log.Println("\nShutting down...")

	// Stop the button manager
	manager.Stop()

	log.Println("Example program exited cleanly")
}
