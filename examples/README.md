# GPIO Buttons Example

This example demonstrates how to use the gpio-buttons-go library to handle 4 buttons connected to Radxa GPIO pins.

## Setup

### Hardware Connections

Connect your buttons/limit switches as follows:

```
Button 1: GPIO1_A0 → Switch → GND
Button 2: GPIO1_A1 → Switch → GND
Button 3: GPIO1_A2 → Switch → GND
Button 4: GPIO1_A3 → Switch → GND
```

**Important:** Replace `GPIO1_A0`, `GPIO1_A1`, etc. in `main.go` with your actual GPIO pin names.

### Finding Your GPIO Pin Names

On your Radxa board, run:

```bash
# Install periph.io tools if not already installed
go install periph.io/x/cmd/periph-info@latest

# List all available GPIO pins
gpio-list
# or
periph-info header
```

Common Radxa pin names:
- **Rock 5B**: GPIO1_A0, GPIO1_A1, GPIO3_A0, GPIO4_B0, etc.
- **Rock 3A**: GPIO0_B5, GPIO0_C1, GPIO3_B1, etc.
- **Zero**: PIN7, PIN11, PIN13, PIN15, etc.

### Update Pin Names

Edit `main.go` and replace the pin names:

```go
// Before
PinName: "GPIO1_A0",

// After (example for Rock 5B)
PinName: "GPIO3_A0",  // Use your actual pin
```

## Building

### Option 1: Using Make

```bash
cd examples
make build
```

### Option 2: Using Go directly

```bash
cd examples
go build -o main .
```

## Running

The example needs GPIO access, which requires root privileges:

### Option 1: Using Make

```bash
cd examples
make run
```

### Option 2: Using sudo

```bash
cd examples
sudo ./main
```

## Expected Output

When you run the example and press buttons, you should see:

```
2025/10/19 10:30:15 Starting GPIO Button Example...
2025/10/19 10:30:15 Added button on pin GPIO1_A0
2025/10/19 10:30:15 Added button on pin GPIO1_A1
2025/10/19 10:30:15 Added button on pin GPIO1_A2
2025/10/19 10:30:15 Added button on pin GPIO1_A3
2025/10/19 10:30:15 Successfully configured 4 button(s)
2025/10/19 10:30:15 Started monitoring 4 buttons
2025/10/19 10:30:15 Button monitoring started. Press Ctrl+C to exit...
2025/10/19 10:30:15 Press any of the 4 buttons to trigger their callbacks.
2025/10/19 10:30:15 Monitoring button on pin GPIO1_A0
2025/10/19 10:30:15 Monitoring button on pin GPIO1_A1
2025/10/19 10:30:15 Monitoring button on pin GPIO1_A2
2025/10/19 10:30:15 Monitoring button on pin GPIO1_A3
2025/10/19 10:30:18 Button pressed on pin GPIO1_A0
2025/10/19 10:30:18 🔴 Button 1 pressed on GPIO1_A0 - Toggling LED
2025/10/19 10:30:20 Button pressed on pin GPIO1_A1
2025/10/19 10:30:20 🟢 Button 2 pressed on GPIO1_A1 - Counter: 1
2025/10/19 10:30:22 Button pressed on pin GPIO1_A2
2025/10/19 10:30:22 🔵 Button 3 pressed on GPIO1_A2 at 10:30:22.145
```

## Customizing Callbacks

Each button has a different callback function that you can customize:

### Button 1: LED Toggle

```go
Callback: func(pinName string) {
    log.Printf("🔴 Button 1 pressed on %s - Toggling LED", pinName)
    // Add your LED control code here
    // Example: toggleGPIOPin("GPIO1_B0")
},
```

### Button 2: Counter

```go
counter := 0
Callback: func(pinName string) {
    counter++
    log.Printf("🟢 Button 2 pressed on %s - Counter: %d", pinName, counter)
    // Use counter for your application logic
},
```

### Button 3: Timestamp

```go
Callback: func(pinName string) {
    timestamp := time.Now().Format("15:04:05.000")
    log.Printf("🔵 Button 3 pressed on %s at %s", pinName, timestamp)
    // Log events, record timing, etc.
},
```

### Button 4: Custom Action

```go
Callback: func(pinName string) {
    log.Printf("🟡 Button 4 pressed on %s - Performing custom action", pinName)
    fmt.Println("  ↳ Custom action executed!")
    // Implement your custom logic here
},
```

## Adjusting Debounce Time

If you experience multiple triggers from a single button press, increase the debounce time:

```go
manager.AddButton(gpiobuttons.ButtonConfig{
    PinName: "GPIO1_A0",
    Callback: myCallback,
    DebounceTime: 100 * time.Millisecond, // Increased from 50ms
})
```

## Troubleshooting

### "Failed to find pin" Error

The GPIO pin name doesn't exist. Check available pins with:
```bash
gpio-list
```

### No Response When Pressing Buttons

1. Verify physical connections
2. Check that buttons connect GPIO to GND when pressed
3. Try different GPIO pins
4. Test pins with `gpio` command:
   ```bash
   gpio -g mode 1 in
   gpio -g mode 1 up
   gpio read 1
   ```

### Permission Denied

Run with sudo:
```bash
sudo ./main
```

### Multiple Triggers Per Press

Increase debounce time in the ButtonConfig.

## Clean Up

Remove built binaries:

```bash
make clean
# or
rm main
```

## Integration with Your Project

To use this in your own project:

1. Import the library:
   ```go
   import gpiobuttons "github.com/piotrpaczula/gpio-buttons-go"
   ```

2. Create and configure manager:
   ```go
   manager, _ := gpiobuttons.NewButtonManager()
   defer manager.Stop()
   
   manager.AddButton(gpiobuttons.ButtonConfig{...})
   manager.Start()
   ```

3. Implement your application logic in the callbacks

## Next Steps

- Integrate with your application (robot control, automation, etc.)
- Add LED output control
- Implement state machines based on button sequences
- Add logging to file or database
- Create a web interface to monitor button presses
