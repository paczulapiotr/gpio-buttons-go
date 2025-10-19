# Quick Reference Guide

## Installation

```bash
go get github.com/piotrpaczula/gpio-buttons-go
```

## Basic Usage

```go
package main

import (
    "log"
    "time"
    gpiobuttons "github.com/piotrpaczula/gpio-buttons-go"
)

func main() {
    // 1. Create manager
    manager, err := gpiobuttons.NewButtonManager()
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Stop()

    // 2. Add buttons
    manager.AddButton(gpiobuttons.ButtonConfig{
        PinName: "GPIO1_A0",
        Callback: func(pin string) {
            log.Printf("Button pressed on %s", pin)
        },
        DebounceTime: 50 * time.Millisecond,
    })

    // 3. Start monitoring
    manager.Start()

    // 4. Keep running
    select {}
}
```

## ButtonConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `PinName` | `string` | **required** | GPIO pin name (e.g., "GPIO1_A0") |
| `Callback` | `ButtonCallback` | **required** | Function called on button press |
| `DebounceTime` | `time.Duration` | 50ms | Time to wait before next press |
| `Pull` | `gpio.Pull` | `PullUp` | Pull resistor (PullUp/PullDown/PullNoChange) |

## ButtonManager Methods

### `NewButtonManager() (*ButtonManager, error)`
Creates a new manager and initializes periph.io.

### `AddButton(config ButtonConfig) error`
Adds a button. Call before `Start()`.

### `Start() error`
Starts monitoring all configured buttons.

### `Stop()`
Stops monitoring and releases resources.

### `GetButtonCount() int`
Returns number of configured buttons.

## Common Patterns

### Multiple Buttons

```go
buttons := []struct {
    pin      string
    action   string
}{
    {"GPIO1_A0", "action1"},
    {"GPIO1_A1", "action2"},
    {"GPIO1_A2", "action3"},
    {"GPIO1_A3", "action4"},
}

for _, btn := range buttons {
    action := btn.action // Capture for closure
    manager.AddButton(gpiobuttons.ButtonConfig{
        PinName: btn.pin,
        Callback: func(pin string) {
            handleAction(action)
        },
    })
}
```

### With State

```go
type AppState struct {
    counter int
    mu      sync.Mutex
}

state := &AppState{}

manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A0",
    Callback: func(pin string) {
        state.mu.Lock()
        state.counter++
        state.mu.Unlock()
    },
})
```

### Graceful Shutdown

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan
log.Println("Shutting down...")
manager.Stop()
```

## GPIO Pin Names by Board

### Radxa Rock 5B
- `GPIO1_A0`, `GPIO1_A1`, `GPIO1_A2`, `GPIO1_A3`
- `GPIO3_A0`, `GPIO3_A1`, `GPIO3_B0`, `GPIO3_B1`
- `GPIO4_B0`, `GPIO4_B1`, `GPIO4_C0`, `GPIO4_C1`

### Radxa Rock 3A
- `GPIO0_B5`, `GPIO0_C1`, `GPIO0_C2`
- `GPIO3_B1`, `GPIO3_B2`, `GPIO3_B3`

### Radxa Zero
- `PIN7`, `PIN11`, `PIN13`, `PIN15`
- `PIN29`, `PIN31`, `PIN33`, `PIN35`

**Always verify with:**
```bash
gpio-list
```

## Debounce Time Recommendations

| Button Type | Recommended Time |
|-------------|------------------|
| Mechanical switch | 50-100ms |
| Limit switch | 20-50ms |
| Touch button | 10-30ms |
| Rotary encoder | 5-10ms |

## Pull Resistor Configuration

```go
import "periph.io/x/conn/v3/gpio"

// Internal pull-up (button connects to GND)
Pull: gpio.PullUp  // DEFAULT

// Internal pull-down (button connects to 3.3V)
Pull: gpio.PullDown

// No internal pull (external resistor required)
Pull: gpio.PullNoChange
```

## Error Handling

```go
if err := manager.AddButton(config); err != nil {
    if strings.Contains(err.Error(), "failed to find pin") {
        log.Println("Invalid GPIO pin name")
    } else if strings.Contains(err.Error(), "failed to configure pin") {
        log.Println("Pin configuration failed")
    }
}
```

## Running Examples

```bash
# Build
cd examples && go build

# Run (requires sudo for GPIO access)
sudo ./main

# Or use make
make run
```

## Testing

```bash
# Run tests
go test -v

# With coverage
go test -cover

# Verbose
go test -v ./...
```

## Dependencies

```bash
# Update dependencies
go get -u periph.io/x/conn/v3
go get -u periph.io/x/host/v3
go mod tidy

# Or use make
make tidy
```

## Troubleshooting Quick Fixes

| Problem | Solution |
|---------|----------|
| "Failed to find pin" | Check pin name with `gpio-list` |
| No response | Verify wiring, check with `gpio read <pin>` |
| Multiple triggers | Increase `DebounceTime` |
| Permission denied | Run with `sudo` |
| Module errors | Run `go mod tidy` |

## Performance Tips

1. **Callbacks are async** - They run in goroutines, don't block
2. **Thread-safe by default** - All methods use mutexes
3. **Efficient polling** - Uses 100ms timeout on WaitForEdge
4. **Low CPU usage** - Event-driven, not busy-wait

## License

MIT License - See LICENSE file for details
