package main

import (
	"fmt"
	"log"
	"sort"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func main() {
	fmt.Println("Discovering available GPIO pins...")
	fmt.Println("=====================================")
	
	// Initialize periph.io
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to initialize periph.io: %v", err)
	}

	// Get all available pins
	allPins := gpioreg.All()
	
	if len(allPins) == 0 {
		fmt.Println("❌ No GPIO pins found!")
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Make sure you're running with sudo")
		fmt.Println("2. Check if GPIO is enabled in your system")
		fmt.Println("3. Try: ls /sys/class/gpio/")
		return
	}

	fmt.Printf("Found %d GPIO pins\n\n", len(allPins))

	// Sort pins by name
	var pinNames []string
	pinMap := make(map[string]gpio.PinIO)
	for _, pin := range allPins {
		name := pin.Name()
		pinNames = append(pinNames, name)
		pinMap[name] = pin
	}
	sort.Strings(pinNames)

	// Test each pin for pull resistor support
	fmt.Println("Pin Name                  | Pull-Up Support | Pin Number")
	fmt.Println("--------------------------|-----------------|------------")
	
	supportedPins := []string{}
	
	for _, name := range pinNames {
		pin := pinMap[name]
		
		// Try to configure with pull-up
		err := pin.In(gpio.PullUp, gpio.NoEdge)
		pullSupport := "✅ YES"
		if err != nil {
			pullSupport = "❌ NO"
		} else {
			supportedPins = append(supportedPins, name)
		}
		
		// Get pin number if available
		pinNum := ""
		if pin.Number() != -1 {
			pinNum = fmt.Sprintf("%d", pin.Number())
		}
		
		fmt.Printf("%-25s | %-15s | %s\n", name, pullSupport, pinNum)
		
		// Clean up
		pin.Halt()
	}

	fmt.Println("\n=====================================")
	fmt.Printf("✅ Pins with pull-up support: %d\n\n", len(supportedPins))
	
	if len(supportedPins) > 0 {
		fmt.Println("🎯 RECOMMENDED PINS FOR BUTTONS:")
		fmt.Println("Use these pin names in your ButtonConfig:")
		fmt.Println()
		count := 0
		for _, name := range supportedPins {
			if count >= 10 {
				break
			}
			pin := pinMap[name]
			pinNum := ""
			if pin.Number() != -1 {
				pinNum = fmt.Sprintf(" (Pin #%d)", pin.Number())
			}
			fmt.Printf("  • %s%s\n", name, pinNum)
			count++
		}
		
		if len(supportedPins) > 10 {
			fmt.Printf("\n  ... and %d more pins\n", len(supportedPins)-10)
		}
		
		fmt.Println("\nExample usage:")
		if len(supportedPins) >= 4 {
			fmt.Println("```go")
			fmt.Printf("manager.AddButton(gpiobuttons.ButtonConfig{\n")
			fmt.Printf("    PinName: \"%s\",\n", supportedPins[0])
			fmt.Printf("    Callback: button1Handler,\n")
			fmt.Printf("})\n")
			fmt.Println("```")
		}
	} else {
		fmt.Println("⚠️  No pins found with pull-up support!")
		fmt.Println("You'll need to use external pull-up resistors.")
		fmt.Println("\nExample with external resistor:")
		if len(pinNames) > 0 {
			fmt.Println("```go")
			fmt.Printf("manager.AddButton(gpiobuttons.ButtonConfig{\n")
			fmt.Printf("    PinName: \"%s\",\n", pinNames[0])
			fmt.Printf("    Callback: handler,\n")
			fmt.Printf("    Pull: gpio.PullNoChange, // External resistor required\n")
			fmt.Printf("})\n")
			fmt.Println("```")
		}
	}
}
