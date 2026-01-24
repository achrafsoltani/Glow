# Extension: Gamepad/Joystick Support

## Overview
Add support for game controllers (Xbox, PlayStation, etc.).

## Linux Joystick Interface

On Linux, joysticks appear as `/dev/input/js*` devices.

### Event Structure

```go
// Linux joystick event (8 bytes)
type JoystickEvent struct {
    Time   uint32  // Timestamp (ms)
    Value  int16   // Axis value or button state
    Type   uint8   // Event type
    Number uint8   // Axis or button number
}

const (
    JS_EVENT_BUTTON = 0x01  // Button pressed/released
    JS_EVENT_AXIS   = 0x02  // Axis moved
    JS_EVENT_INIT   = 0x80  // Initial state
)
```

### Gamepad Structure

```go
type Gamepad struct {
    file    *os.File
    name    string
    buttons []bool
    axes    []float64

    // Mapped inputs
    ButtonA     bool
    ButtonB     bool
    ButtonX     bool
    ButtonY     bool
    ButtonStart bool
    LeftStickX  float64
    LeftStickY  float64
    RightStickX float64
    RightStickY float64
    LeftTrigger float64
    RightTrigger float64
    DPadX       int  // -1, 0, 1
    DPadY       int
}

func OpenGamepad(deviceNum int) (*Gamepad, error) {
    path := fmt.Sprintf("/dev/input/js%d", deviceNum)
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    // Get name
    name := make([]byte, 128)
    // ioctl JSIOCGNAME

    // Get number of axes and buttons
    // ioctl JSIOCGAXES, JSIOCGBUTTONS

    return &Gamepad{
        file:    file,
        buttons: make([]bool, 16),
        axes:    make([]float64, 8),
    }, nil
}
```

### Reading Events

```go
func (g *Gamepad) Poll() error {
    // Set non-blocking
    g.file.SetReadDeadline(time.Now().Add(time.Millisecond))

    buf := make([]byte, 8)
    for {
        _, err := g.file.Read(buf)
        if err != nil {
            break  // No more events
        }

        event := JoystickEvent{
            Time:   binary.LittleEndian.Uint32(buf[0:4]),
            Value:  int16(binary.LittleEndian.Uint16(buf[4:6])),
            Type:   buf[6],
            Number: buf[7],
        }

        g.processEvent(event)
    }

    return nil
}

func (g *Gamepad) processEvent(e JoystickEvent) {
    eventType := e.Type & ^uint8(JS_EVENT_INIT)

    switch eventType {
    case JS_EVENT_BUTTON:
        if int(e.Number) < len(g.buttons) {
            g.buttons[e.Number] = e.Value != 0
        }

    case JS_EVENT_AXIS:
        if int(e.Number) < len(g.axes) {
            // Normalize to -1.0 to 1.0
            g.axes[e.Number] = float64(e.Value) / 32767.0
        }
    }

    g.mapInputs()
}
```

### Controller Mapping

Different controllers have different button layouts:

```go
// Xbox-style mapping
func (g *Gamepad) mapInputsXbox() {
    g.ButtonA = g.buttons[0]
    g.ButtonB = g.buttons[1]
    g.ButtonX = g.buttons[2]
    g.ButtonY = g.buttons[3]
    g.ButtonStart = g.buttons[7]

    g.LeftStickX = g.axes[0]
    g.LeftStickY = g.axes[1]
    g.RightStickX = g.axes[3]
    g.RightStickY = g.axes[4]
    g.LeftTrigger = g.axes[2]
    g.RightTrigger = g.axes[5]

    // D-Pad (often on axes 6, 7)
    g.DPadX = int(g.axes[6])
    g.DPadY = int(g.axes[7])
}

// Deadzone handling
func applyDeadzone(value float64, deadzone float64) float64 {
    if math.Abs(value) < deadzone {
        return 0
    }
    return value
}
```

### Finding Gamepads

```go
func FindGamepads() []int {
    var gamepads []int

    for i := 0; i < 8; i++ {
        path := fmt.Sprintf("/dev/input/js%d", i)
        if _, err := os.Stat(path); err == nil {
            gamepads = append(gamepads, i)
        }
    }

    return gamepads
}
```

### Usage Example

```go
func main() {
    win, _ := glow.NewWindow("Gamepad Test", 800, 600)
    defer win.Close()

    gamepad, err := glow.OpenGamepad(0)
    if err != nil {
        fmt.Println("No gamepad found, using keyboard")
    }

    for running {
        // Poll gamepad
        if gamepad != nil {
            gamepad.Poll()

            // Use gamepad input
            player.X += gamepad.LeftStickX * speed
            player.Y += gamepad.LeftStickY * speed

            if gamepad.ButtonA {
                player.Jump()
            }
        }

        // ... render
    }
}
```

### Hot-Plugging

Watch for device changes:

```go
func WatchGamepads(callback func(added bool, deviceNum int)) {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add("/dev/input")

    for event := range watcher.Events {
        if strings.HasPrefix(filepath.Base(event.Name), "js") {
            num, _ := strconv.Atoi(event.Name[len(event.Name)-1:])
            callback(event.Op == fsnotify.Create, num)
        }
    }
}
```

## Files to Create

1. `gamepad.go` - Gamepad type, event reading
2. `gamepad_linux.go` - Linux-specific implementation

## Testing

```bash
# List devices
ls /dev/input/js*

# Test with jstest
sudo apt install joystick
jstest /dev/input/js0
```

## Resources

- Linux Joystick API: https://www.kernel.org/doc/Documentation/input/joystick-api.txt
- SDL GameController DB: https://github.com/gabomdq/SDL_GameControllerDB
