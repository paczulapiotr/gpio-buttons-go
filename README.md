# GPIO Buttons Go Library

A Go library for handling GPIO button inputs on Radxa boards (and other Linux SBCs) using the periph.io library. This library provides an easy way to configure multiple buttons with custom callbacks that trigger on button press events.

## Features

- 🎯 Simple API for configuring multiple GPIO buttons
- 🔔 Callback-based event handling
- ⏱️ Built-in debouncing support
- 🔧 Configurable pull resistors (Pull-Up/Pull-Down/None)
- 🧵 Thread-safe operations
- 🛑 Graceful shutdown support

## Installation

```bash
go get github.com/piotrpaczula/gpio-buttons-go
```

## Hardware Requirements

- Radxa board (or any Linux SBC supported by periph.io)
- 4 limit switches or buttons connected to GPIO pins
- Recommended: Pull-up resistors (or use internal pull-ups)

### Wiring

Connect your buttons/limit switches to GPIO pins:
- Button 1 → GPIO1_A0 (or your chosen pin)
- Button 2 → GPIO1_A1
- Button 3 → GPIO1_A2
- Button 4 → GPIO1_A3

Each button should connect the GPIO pin to GND when pressed.

## Quick Start

```go
package main

import (
    "log"
    "time"
    
    gpiobuttons "github.com/piotrpaczula/gpio-buttons-go"
)

func main() {
    // Create a new button manager
    manager, err := gpiobuttons.NewButtonManager()
    if err != nil {
        log.Fatalf("Failed to create button manager: %v", err)
    }
    defer manager.Stop()

    // Add a button with a callback
    err = manager.AddButton(gpiobuttons.ButtonConfig{
        PinName: "GPIO1_A0",
        Callback: func(pinName string) {
            log.Printf("Button pressed on %s!", pinName)
        },
        DebounceTime: 50 * time.Millisecond,
    })
    if err != nil {
        log.Fatalf("Failed to add button: %v", err)
    }

    // Start monitoring
    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }

    // Keep the program running
    select {}
}
```

## API Reference

### ButtonManager

#### `NewButtonManager() (*ButtonManager, error)`

Creates a new ButtonManager instance and initializes the periph.io host.

#### `AddButton(config ButtonConfig) error`

Adds a button configuration to the manager. Must be called before `Start()`.

**ButtonConfig fields:**
- `PinName` (string): GPIO pin name (e.g., "GPIO1_A0")
- `Callback` (ButtonCallback): Function to call on button press
- `DebounceTime` (time.Duration): Debounce period (default: 50ms)
- `Pull` (gpio.Pull): Pull resistor configuration (default: PullUp)

#### `Start() error`

Begins monitoring all configured buttons.

#### `Stop()`

Stops monitoring all buttons and releases resources. Safe to call multiple times.

#### `GetButtonCount() int`

Returns the number of configured buttons.

### ButtonCallback

```go
type ButtonCallback func(pinName string)
```

A function that receives the pin name when a button is pressed.

## Example Usage

See the complete example in `examples/main.go`:

```bash
cd examples
go run main.go
```

### Example: Multiple Buttons with Different Actions

```go
manager, _ := gpiobuttons.NewButtonManager()
defer manager.Stop()

// Button 1: Toggle LED
manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A0",
    Callback: func(pin string) {
        toggleLED()
    },
})

// Button 2: Increment counter
counter := 0
manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A1",
    Callback: func(pin string) {
        counter++
        log.Printf("Counter: %d", counter)
    },
})

// Button 3: Custom action
manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A2",
    Callback: func(pin string) {
        performCustomAction()
    },
    DebounceTime: 100 * time.Millisecond, // Custom debounce
})

// Button 4: Emergency stop
manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A3",
    Callback: func(pin string) {
        log.Println("Emergency stop!")
        manager.Stop()
    },
})

manager.Start()
```

## Finding GPIO Pin Names

To find available GPIO pins on your Radxa board:

```bash
# List all GPIO pins
gpio-list

# Or use the periph.io tool
periph-info
```

Common Radxa GPIO pin names:
- `GPIO1_A0`, `GPIO1_A1`, `GPIO1_A2`, `GPIO1_A3`
- `GPIO1_B0`, `GPIO1_B1`, `GPIO1_B2`, `GPIO1_B3`
- `GPIO2_A0`, `GPIO2_A1`, etc.

## Configuration Tips

### Debounce Time

Adjust based on your button type:
- Mechanical switches: 50-100ms
- Limit switches: 20-50ms
- Touch buttons: 10-30ms

### Pull Resistors

```go
import "periph.io/x/conn/v3/gpio"

// Use internal pull-up (default)
Pull: gpio.PullUp

// Use internal pull-down
Pull: gpio.PullDown

// No pull resistor (external required)
Pull: gpio.PullNoChange
```

## Error Handling

The library provides detailed error messages:

```go
if err := manager.AddButton(config); err != nil {
    if strings.Contains(err.Error(), "failed to find pin") {
        log.Println("GPIO pin not found. Check your pin name.")
    }
}
```

## Thread Safety

All ButtonManager methods are thread-safe. Callbacks are executed in separate goroutines, so they won't block button monitoring.

## Dependencies

- [periph.io/x/conn/v3](https://pkg.go.dev/periph.io/x/conn/v3) - GPIO interfaces
- [periph.io/x/host/v3](https://pkg.go.dev/periph.io/x/host/v3) - Hardware drivers

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [periph.io documentation](https://periph.io/)

## Acknowledgments

Built with [periph.io](https://periph.io/) - a comprehensive Go library for peripheral I/O.
