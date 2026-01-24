# Step 8: Event Handling

## Goal
Read and parse X11 events (keyboard, mouse, window events).

## What You'll Learn
- X11 event structure
- Parsing different event types
- Building an event loop

## 8.1 Event Basics

All X11 events are exactly 32 bytes. The first byte is the event type.

```
┌─────────┬───────────────────────────────────────────────────────────┐
│  Type   │                      Event Data                          │
│ 1 byte  │                       31 bytes                           │
└─────────┴───────────────────────────────────────────────────────────┘
```

## 8.2 Event Types

Add to `protocol.go`:

```go
// Event types
const (
    EventKeyPress        = 2
    EventKeyRelease      = 3
    EventButtonPress     = 4
    EventButtonRelease   = 5
    EventMotionNotify    = 6
    EventEnterNotify     = 7
    EventLeaveNotify     = 8
    EventFocusIn         = 9
    EventFocusOut        = 10
    EventExpose          = 12
    EventDestroyNotify   = 17
    EventUnmapNotify     = 18
    EventMapNotify       = 19
    EventConfigureNotify = 22
    EventClientMessage   = 33
)
```

## 8.3 Event Structures

Create `internal/x11/event.go`:

```go
package x11

import (
    "encoding/binary"
    "io"
)

// Event interface
type Event interface {
    Type() int
}

// KeyEvent for keyboard
type KeyEvent struct {
    EventType int
    Keycode   uint8
    State     uint16  // Modifiers (shift, ctrl, etc.)
    X, Y      int16   // Position in window
    RootX     int16   // Position on screen
    RootY     int16
}

func (e KeyEvent) Type() int { return e.EventType }

// ButtonEvent for mouse buttons
type ButtonEvent struct {
    EventType int
    Button    uint8   // 1=left, 2=middle, 3=right, 4/5=wheel
    State     uint16
    X, Y      int16
    RootX     int16
    RootY     int16
}

func (e ButtonEvent) Type() int { return e.EventType }

// MotionEvent for mouse movement
type MotionEvent struct {
    X, Y      int16
    RootX     int16
    RootY     int16
    State     uint16  // Which buttons held
}

func (e MotionEvent) Type() int { return EventMotionNotify }

// ExposeEvent - window needs redraw
type ExposeEvent struct {
    Window uint32
    X, Y   uint16
    Width  uint16
    Height uint16
    Count  uint16  // More expose events following
}

func (e ExposeEvent) Type() int { return EventExpose }

// ConfigureEvent - window moved/resized
type ConfigureEvent struct {
    Window uint32
    X, Y   int16
    Width  uint16
    Height uint16
}

func (e ConfigureEvent) Type() int { return EventConfigureNotify }

// ClientMessageEvent - from window manager
type ClientMessageEvent struct {
    Window      uint32
    Format      uint8
    MessageType uint32
    Data        [20]byte
}

func (e ClientMessageEvent) Type() int { return EventClientMessage }
```

## 8.4 Reading Events

```go
// NextEvent reads and parses the next event (blocks)
func (c *Connection) NextEvent() (Event, error) {
    buf := make([]byte, 32)
    _, err := io.ReadFull(c.conn, buf)
    if err != nil {
        return nil, err
    }

    // Event type (high bit is "sent by SendEvent")
    eventType := int(buf[0] & 0x7F)

    switch eventType {
    case EventKeyPress, EventKeyRelease:
        return KeyEvent{
            EventType: eventType,
            Keycode:   buf[1],
            State:     binary.LittleEndian.Uint16(buf[28:30]),
            X:         int16(binary.LittleEndian.Uint16(buf[24:26])),
            Y:         int16(binary.LittleEndian.Uint16(buf[26:28])),
            RootX:     int16(binary.LittleEndian.Uint16(buf[20:22])),
            RootY:     int16(binary.LittleEndian.Uint16(buf[22:24])),
        }, nil

    case EventButtonPress, EventButtonRelease:
        return ButtonEvent{
            EventType: eventType,
            Button:    buf[1],
            State:     binary.LittleEndian.Uint16(buf[28:30]),
            X:         int16(binary.LittleEndian.Uint16(buf[24:26])),
            Y:         int16(binary.LittleEndian.Uint16(buf[26:28])),
            RootX:     int16(binary.LittleEndian.Uint16(buf[20:22])),
            RootY:     int16(binary.LittleEndian.Uint16(buf[22:24])),
        }, nil

    case EventMotionNotify:
        return MotionEvent{
            X:     int16(binary.LittleEndian.Uint16(buf[24:26])),
            Y:     int16(binary.LittleEndian.Uint16(buf[26:28])),
            RootX: int16(binary.LittleEndian.Uint16(buf[20:22])),
            RootY: int16(binary.LittleEndian.Uint16(buf[22:24])),
            State: binary.LittleEndian.Uint16(buf[28:30]),
        }, nil

    case EventExpose:
        return ExposeEvent{
            Window: binary.LittleEndian.Uint32(buf[4:8]),
            X:      binary.LittleEndian.Uint16(buf[8:10]),
            Y:      binary.LittleEndian.Uint16(buf[10:12]),
            Width:  binary.LittleEndian.Uint16(buf[12:14]),
            Height: binary.LittleEndian.Uint16(buf[14:16]),
            Count:  binary.LittleEndian.Uint16(buf[16:18]),
        }, nil

    case EventConfigureNotify:
        return ConfigureEvent{
            Window: binary.LittleEndian.Uint32(buf[4:8]),
            X:      int16(binary.LittleEndian.Uint16(buf[16:18])),
            Y:      int16(binary.LittleEndian.Uint16(buf[18:20])),
            Width:  binary.LittleEndian.Uint16(buf[20:22]),
            Height: binary.LittleEndian.Uint16(buf[22:24]),
        }, nil

    case EventClientMessage:
        e := ClientMessageEvent{
            Window:      binary.LittleEndian.Uint32(buf[4:8]),
            Format:      buf[1],
            MessageType: binary.LittleEndian.Uint32(buf[8:12]),
        }
        copy(e.Data[:], buf[12:32])
        return e, nil

    default:
        return nil, nil  // Unknown event
    }
}
```

## 8.5 Event Loop Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/glow/internal/x11"
)

func main() {
    conn, _ := x11.Connect()
    defer conn.Close()

    windowID, _ := conn.CreateWindow(100, 100, 800, 600)
    conn.MapWindow(windowID)

    fmt.Println("Press Q to quit")

    running := true
    for running {
        event, err := conn.NextEvent()
        if err != nil {
            break
        }

        switch e := event.(type) {
        case x11.KeyEvent:
            if e.EventType == x11.EventKeyPress {
                fmt.Printf("Key pressed: %d\n", e.Keycode)
                if e.Keycode == 24 {  // Q key
                    running = false
                }
            }

        case x11.ButtonEvent:
            if e.EventType == x11.EventButtonPress {
                fmt.Printf("Mouse button %d at (%d, %d)\n",
                    e.Button, e.X, e.Y)
            }

        case x11.MotionEvent:
            // Mouse moved - often ignored (too many events)

        case x11.ExposeEvent:
            fmt.Printf("Window needs redraw\n")

        case x11.ConfigureEvent:
            fmt.Printf("Window resized to %dx%d\n", e.Width, e.Height)
        }
    }
}
```

## 8.6 Understanding Key Codes

Key codes are hardware-dependent. Common values on QWERTY keyboards:

| Key | Code | Key | Code |
|-----|------|-----|------|
| Escape | 9 | Space | 65 |
| Q | 24 | A | 38 |
| W | 25 | S | 39 |
| E | 26 | D | 40 |
| Up | 111 | Down | 116 |
| Left | 113 | Right | 114 |

Use `xev` to discover key codes:
```bash
xev | grep keycode
```

## 8.7 Modifier State

The `State` field shows which modifiers are held:

```go
const (
    ShiftMask   = 1 << 0
    LockMask    = 1 << 1  // Caps lock
    ControlMask = 1 << 2
    Mod1Mask    = 1 << 3  // Alt
    Mod4Mask    = 1 << 6  // Super/Win
)

if e.State & ControlMask != 0 {
    fmt.Println("Ctrl is held")
}
```

## Next Step
Continue to Step 9 to make events non-blocking.
