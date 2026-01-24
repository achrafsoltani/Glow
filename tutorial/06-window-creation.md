# Step 6: Window Creation

## Goal
Create and display an X11 window.

## What You'll Learn
- CreateWindow request structure
- Window attributes and event masks
- MapWindow to make windows visible

## 6.1 Protocol Constants

Create/update `internal/x11/protocol.go`:

```go
package x11

// Request opcodes
const (
    OpCreateWindow   = 1
    OpDestroyWindow  = 4
    OpMapWindow      = 8
    OpUnmapWindow    = 10
)

// Window classes
const (
    WindowClassCopyFromParent = 0
    WindowClassInputOutput    = 1
    WindowClassInputOnly      = 2
)

// Window attribute masks (CW = Create Window)
const (
    CWBackPixmap       = 1 << 0
    CWBackPixel        = 1 << 1
    CWBorderPixmap     = 1 << 2
    CWBorderPixel      = 1 << 3
    CWBitGravity       = 1 << 4
    CWWinGravity       = 1 << 5
    CWBackingStore     = 1 << 6
    CWBackingPlanes    = 1 << 7
    CWBackingPixel     = 1 << 8
    CWOverrideRedirect = 1 << 9
    CWSaveUnder        = 1 << 10
    CWEventMask        = 1 << 11
    CWDontPropagate    = 1 << 12
    CWColormap         = 1 << 13
    CWCursor           = 1 << 14
)

// Event masks
const (
    KeyPressMask        = 1 << 0
    KeyReleaseMask      = 1 << 1
    ButtonPressMask     = 1 << 2
    ButtonReleaseMask   = 1 << 3
    EnterWindowMask     = 1 << 4
    LeaveWindowMask     = 1 << 5
    PointerMotionMask   = 1 << 6
    ExposureMask        = 1 << 15
    StructureNotifyMask = 1 << 17
    FocusChangeMask     = 1 << 21
)
```

## 6.2 CreateWindow Request

The CreateWindow request structure:

```
┌────────┬───────┬────────┬──────────┬────────┬───────┬───────┬────────┐
│ Opcode │ Depth │ Length │ WindowID │ Parent │   X   │   Y   │ Width  │
│   1    │   0   │   8+n  │  4 bytes │4 bytes │2 bytes│2 bytes│2 bytes │
├────────┴───────┴────────┴──────────┴────────┴───────┴───────┴────────┤
│ Height │ Border │ Class  │ Visual │ Value Mask │ Values...            │
│2 bytes │2 bytes │2 bytes │4 bytes │  4 bytes   │ 4 bytes each         │
└────────┴────────┴────────┴────────┴────────────┴──────────────────────┘
```

## 6.3 Implementation

Create `internal/x11/window.go`:

```go
package x11

import "encoding/binary"

// CreateWindow creates a new window
func (c *Connection) CreateWindow(x, y int16, width, height uint16) (uint32, error) {
    windowID := c.GenerateID()

    // Events we want to receive
    eventMask := uint32(
        ExposureMask |
        KeyPressMask |
        KeyReleaseMask |
        ButtonPressMask |
        ButtonReleaseMask |
        PointerMotionMask |
        StructureNotifyMask,
    )

    // Attributes: background pixel and event mask
    valueMask := uint32(CWBackPixel | CWEventMask)
    valueCount := 2

    // Request length in 4-byte units
    reqLen := 8 + valueCount
    req := make([]byte, reqLen*4)

    req[0] = OpCreateWindow                                  // Opcode
    req[1] = c.RootDepth                                     // Depth
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))   // Length
    binary.LittleEndian.PutUint32(req[4:], windowID)         // Window ID
    binary.LittleEndian.PutUint32(req[8:], c.RootWindow)     // Parent
    binary.LittleEndian.PutUint16(req[12:], uint16(x))       // X
    binary.LittleEndian.PutUint16(req[14:], uint16(y))       // Y
    binary.LittleEndian.PutUint16(req[16:], width)           // Width
    binary.LittleEndian.PutUint16(req[18:], height)          // Height
    binary.LittleEndian.PutUint16(req[20:], 0)               // Border width
    binary.LittleEndian.PutUint16(req[22:], WindowClassInputOutput)
    binary.LittleEndian.PutUint32(req[24:], c.RootVisual)    // Visual
    binary.LittleEndian.PutUint32(req[28:], valueMask)       // Value mask

    // Values (order matches bit position in mask)
    binary.LittleEndian.PutUint32(req[32:], 0x00000000) // BackPixel (black)
    binary.LittleEndian.PutUint32(req[36:], eventMask) // EventMask

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    return windowID, nil
}
```

## 6.4 MapWindow

A window is invisible until mapped:

```go
// MapWindow makes a window visible
func (c *Connection) MapWindow(windowID uint32) error {
    req := make([]byte, 8)
    req[0] = OpMapWindow
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2)  // Length
    binary.LittleEndian.PutUint32(req[4:], windowID)

    _, err := c.conn.Write(req)
    return err
}

// UnmapWindow hides a window
func (c *Connection) UnmapWindow(windowID uint32) error {
    req := make([]byte, 8)
    req[0] = OpUnmapWindow
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2)
    binary.LittleEndian.PutUint32(req[4:], windowID)

    _, err := c.conn.Write(req)
    return err
}

// DestroyWindow destroys a window
func (c *Connection) DestroyWindow(windowID uint32) error {
    req := make([]byte, 8)
    req[0] = OpDestroyWindow
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2)
    binary.LittleEndian.PutUint32(req[4:], windowID)

    _, err := c.conn.Write(req)
    return err
}
```

## 6.5 Test It

Create `examples/window/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/yourusername/glow/internal/x11"
)

func main() {
    conn, err := x11.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create window
    windowID, err := conn.CreateWindow(100, 100, 800, 600)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created window: 0x%x\n", windowID)

    // Make it visible
    if err := conn.MapWindow(windowID); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Window mapped!")

    // Keep it open
    time.Sleep(5 * time.Second)

    // Clean up
    conn.DestroyWindow(windowID)
}
```

Run:
```bash
go run examples/window/main.go
```

You should see a black 800x600 window for 5 seconds!

## 6.6 Understanding Value Masks

The value mask is important: values are sent **in order of bit position**.

```go
// If mask is CWBackPixel | CWEventMask (bits 1 and 11)
// Values are sent as: BackPixel, EventMask

// If mask is CWEventMask | CWBackPixel (same bits, different order in code)
// Values are STILL sent as: BackPixel, EventMask
// (order is determined by bit position, not code order)
```

## 6.7 Window Hierarchy

```
Root Window (managed by X server)
├── Your Window
│   ├── Child Window (optional)
│   └── Another Child
└── Another App's Window
```

Every window has a parent. Your windows are children of the root window.

## Common Issues

### Window appears but is tiny/wrong position
- Check x, y, width, height values
- Ensure proper byte order

### Window doesn't appear
- Did you call MapWindow?
- Is the window off-screen?

### Window appears then disappears
- Program exiting too fast
- Add sleep or event loop

## Next Step
Continue to Step 7 to add window title and close button.
