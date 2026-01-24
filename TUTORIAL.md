# Building a 2D Graphics Library in Pure Go

A hands-on tutorial to master Go and low-level systems programming by building an SDL-inspired graphics library from scratch.

## What You'll Learn

- Go language fundamentals (structs, interfaces, concurrency, binary encoding)
- Systems programming (sockets, syscalls, binary protocols)
- X11 window system internals
- Graphics programming (framebuffers, pixel manipulation, blitting)
- Software architecture and API design

## Prerequisites

- Basic programming knowledge (any language)
- Linux system with X11 (most desktop Linux distros)
- Go 1.21+ installed
- A text editor or IDE

## Project Overview

```
Final Architecture:

┌─────────────────────────────────────────────────────┐
│                  Your Application                   │
├─────────────────────────────────────────────────────┤
│                   glow (Public API)                 │
│         Window | Canvas | Events | Timing          │
├─────────────────────────────────────────────────────┤
│                  internal/x11                       │
│     Connection | Protocol | Atoms | SHM            │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
              X11 Server (Unix Socket)
```

---

# Part 1: Foundations

## Chapter 1.1: Project Setup

Create the project structure:

```bash
cd /home/achraf/Documents/GolandProjects/AchrafSoltani/glow
go mod init github.com/AchrafSoltani/glow
```

Create the directory structure:

```bash
mkdir -p internal/x11
mkdir -p examples
```

Your structure should look like:

```
glow/
├── go.mod
├── glow.go             # Public API
├── internal/
│   └── x11/
│       ├── conn.go     # X11 connection
│       ├── protocol.go # X11 message types
│       ├── window.go   # Window management
│       ├── event.go    # Event handling
│       └── draw.go     # Drawing operations
└── examples/
    └── hello/
        └── main.go     # First example
```

## Chapter 1.2: Go Fundamentals You'll Need

Before diving in, make sure you understand these Go concepts:

### Binary Encoding

X11 uses binary protocols. Go's `encoding/binary` package is essential:

```go
package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

func main() {
    // Writing binary data
    buf := new(bytes.Buffer)

    // Write a uint32 in little-endian
    binary.Write(buf, binary.LittleEndian, uint32(12345))

    fmt.Printf("Bytes: % x\n", buf.Bytes())
    // Output: Bytes: 39 30 00 00

    // Reading binary data
    var value uint32
    binary.Read(buf, binary.LittleEndian, &value)
    fmt.Printf("Value: %d\n", value)
}
```

**Exercise 1.2.1**: Write a program that:
1. Creates a struct with mixed types (uint8, uint16, uint32)
2. Serializes it to bytes
3. Deserializes it back

### Unix Sockets

X11 communicates over Unix domain sockets:

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    // Connect to X11 server
    conn, err := net.Dial("unix", "/tmp/.X11-unix/X0")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    fmt.Println("Connected to X11 server!")
}
```

### Byte Slices and Manual Packing

Sometimes you need manual control:

```go
func packUint32(buf []byte, offset int, value uint32) {
    buf[offset] = byte(value)
    buf[offset+1] = byte(value >> 8)
    buf[offset+2] = byte(value >> 16)
    buf[offset+3] = byte(value >> 24)
}

func unpackUint32(buf []byte, offset int) uint32 {
    return uint32(buf[offset]) |
        uint32(buf[offset+1])<<8 |
        uint32(buf[offset+2])<<16 |
        uint32(buf[offset+3])<<24
}
```

**Exercise 1.2.2**: Implement `packUint16` and `unpackUint16`.

---

# Part 2: X11 Connection

## Chapter 2.1: Understanding X11

X11 is a client-server protocol:
- **X Server**: Manages the display, keyboard, mouse
- **X Client**: Your application (yes, the naming is backwards!)

Communication happens via:
1. Unix socket: `/tmp/.X11-unix/X{display_number}`
2. TCP socket: Port 6000 + display_number

The `DISPLAY` environment variable tells you which display to use:
- `:0` = Local display 0, Unix socket
- `localhost:0` = Display 0 via TCP
- `192.168.1.5:0` = Remote display

## Chapter 2.2: Connection Handshake

Create `internal/x11/conn.go`:

```go
package x11

import (
    "encoding/binary"
    "errors"
    "fmt"
    "net"
    "os"
    "strings"
)

// Connection represents a connection to the X11 server
type Connection struct {
    conn   net.Conn

    // Setup information from server
    ResourceIDBase uint32
    ResourceIDMask uint32
    RootWindow     uint32
    RootVisual     uint32
    RootDepth      uint8
    ScreenWidth    uint16
    ScreenHeight   uint16

    // ID generation
    nextID uint32
}

// Connect establishes a connection to the X11 server
func Connect() (*Connection, error) {
    display := os.Getenv("DISPLAY")
    if display == "" {
        display = ":0"
    }

    // Parse display string (e.g., ":0" or ":0.0")
    displayNum := "0"
    if idx := strings.Index(display, ":"); idx != -1 {
        rest := display[idx+1:]
        if dotIdx := strings.Index(rest, "."); dotIdx != -1 {
            displayNum = rest[:dotIdx]
        } else {
            displayNum = rest
        }
    }

    // Connect via Unix socket
    socketPath := fmt.Sprintf("/tmp/.X11-unix/X%s", displayNum)
    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to X11: %w", err)
    }

    c := &Connection{conn: conn}

    if err := c.handshake(); err != nil {
        conn.Close()
        return nil, err
    }

    return c, nil
}

// Close closes the connection
func (c *Connection) Close() error {
    return c.conn.Close()
}

func (c *Connection) handshake() error {
    // Send connection setup request
    // Byte order: 'l' for little-endian, 'B' for big-endian
    setup := make([]byte, 12)
    setup[0] = 'l'                                    // Little-endian
    setup[1] = 0                                      // Unused
    binary.LittleEndian.PutUint16(setup[2:], 11)     // Protocol major version
    binary.LittleEndian.PutUint16(setup[4:], 0)      // Protocol minor version
    binary.LittleEndian.PutUint16(setup[6:], 0)      // Auth protocol name length
    binary.LittleEndian.PutUint16(setup[8:], 0)      // Auth data length
    binary.LittleEndian.PutUint16(setup[10:], 0)     // Unused

    if _, err := c.conn.Write(setup); err != nil {
        return fmt.Errorf("failed to send setup: %w", err)
    }

    // Read response header (8 bytes minimum)
    header := make([]byte, 8)
    if _, err := c.conn.Read(header); err != nil {
        return fmt.Errorf("failed to read response: %w", err)
    }

    status := header[0]
    switch status {
    case 0: // Failed
        reasonLen := header[1]
        reason := make([]byte, reasonLen)
        c.conn.Read(reason)
        return fmt.Errorf("connection failed: %s", string(reason))
    case 1: // Success
        return c.parseSetupSuccess(header)
    case 2: // Authenticate
        return errors.New("authentication required (not implemented)")
    default:
        return fmt.Errorf("unknown status: %d", status)
    }
}

func (c *Connection) parseSetupSuccess(header []byte) error {
    // Additional data length is in header[6:8] (in 4-byte units)
    additionalLen := binary.LittleEndian.Uint16(header[6:]) * 4

    // Read the rest of the setup response
    data := make([]byte, additionalLen)
    if _, err := c.conn.Read(data); err != nil {
        return fmt.Errorf("failed to read setup data: %w", err)
    }

    // Parse the setup response
    // Offset 0-3: Release number
    // Offset 4-7: Resource ID base
    // Offset 8-11: Resource ID mask
    // ... (many more fields)

    c.ResourceIDBase = binary.LittleEndian.Uint32(data[4:8])
    c.ResourceIDMask = binary.LittleEndian.Uint32(data[8:12])

    // Skip to screen info
    // This is simplified - real parsing needs to handle variable-length data
    vendorLen := binary.LittleEndian.Uint16(data[16:18])
    numFormats := data[21]
    numScreens := data[20]

    if numScreens == 0 {
        return errors.New("no screens available")
    }

    // Calculate offset to first screen
    // Vendor string is padded to 4-byte boundary
    vendorPadded := (vendorLen + 3) &^ 3
    formatSize := uint16(numFormats) * 8
    screenOffset := 32 + vendorPadded + formatSize

    // Parse first screen
    screen := data[screenOffset:]
    c.RootWindow = binary.LittleEndian.Uint32(screen[0:4])
    c.ScreenWidth = binary.LittleEndian.Uint16(screen[20:22])
    c.ScreenHeight = binary.LittleEndian.Uint16(screen[22:24])
    c.RootDepth = screen[38]
    c.RootVisual = binary.LittleEndian.Uint32(screen[32:36])

    // Initialize ID generator
    c.nextID = c.ResourceIDBase

    return nil
}

// GenerateID generates a new resource ID
func (c *Connection) GenerateID() uint32 {
    id := c.nextID
    c.nextID++
    return (id & c.ResourceIDMask) | c.ResourceIDBase
}
```

## Chapter 2.3: Test the Connection

Create `examples/connect/main.go`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
    conn, err := x11.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    fmt.Println("Successfully connected to X11!")
    fmt.Printf("Screen size: %dx%d\n", conn.ScreenWidth, conn.ScreenHeight)
    fmt.Printf("Root window: 0x%x\n", conn.RootWindow)
    fmt.Printf("Root depth: %d\n", conn.RootDepth)
}
```

Run it:

```bash
go run examples/connect/main.go
```

**Expected output**:
```
Successfully connected to X11!
Screen size: 1920x1080
Root window: 0x299
Root depth: 24
```

**Exercise 2.3.1**: Add error handling for when DISPLAY is not set.

**Exercise 2.3.2**: Parse and print more information from the setup response (vendor name, protocol version).

---

# Part 3: Creating a Window

## Chapter 3.1: X11 Protocol Basics

Every X11 request has this structure:

```
┌─────────┬─────────┬─────────────────┬─────────────────┐
│ Opcode  │  Data   │ Request Length  │   Parameters    │
│ 1 byte  │ 1 byte  │    2 bytes      │    Variable     │
└─────────┴─────────┴─────────────────┴─────────────────┘
```

Create `internal/x11/protocol.go`:

```go
package x11

// X11 Request Opcodes
const (
    OpCreateWindow     = 1
    OpMapWindow        = 8
    OpUnmapWindow      = 10
    OpConfigureWindow  = 12
    OpDestroyWindow    = 4
    OpChangeProperty   = 18
    OpDeleteProperty   = 19
    OpGetProperty      = 20
    OpInternAtom       = 16
    OpCreateGC         = 55
    OpPutImage         = 72
    OpPolyFillRect     = 70
    OpFreeGC           = 60
)

// Window classes
const (
    WindowClassCopyFromParent = 0
    WindowClassInputOutput    = 1
    WindowClassInputOnly      = 2
)

// Window attributes mask
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

// Event types
const (
    EventKeyPress         = 2
    EventKeyRelease       = 3
    EventButtonPress      = 4
    EventButtonRelease    = 5
    EventMotionNotify     = 6
    EventEnterNotify      = 7
    EventLeaveNotify      = 8
    EventFocusIn          = 9
    EventFocusOut         = 10
    EventExpose           = 12
    EventDestroyNotify    = 17
    EventUnmapNotify      = 18
    EventMapNotify        = 19
    EventConfigureNotify  = 22
    EventClientMessage    = 33
)

// Image formats
const (
    ImageFormatBitmap = 0
    ImageFormatXYPixmap = 1
    ImageFormatZPixmap = 2
)
```

## Chapter 3.2: Window Creation

Create `internal/x11/window.go`:

```go
package x11

import (
    "encoding/binary"
)

// CreateWindow creates a new window
func (c *Connection) CreateWindow(x, y int16, width, height uint16) (uint32, error) {
    windowID := c.GenerateID()

    // CreateWindow request
    // Opcode: 1
    // Depth: copy from parent (0)
    // Length: 8 + number of value items (we have 2: back pixel and event mask)

    eventMask := uint32(ExposureMask | KeyPressMask | KeyReleaseMask |
        ButtonPressMask | ButtonReleaseMask | PointerMotionMask |
        StructureNotifyMask)

    valueMask := uint32(CWBackPixel | CWEventMask)
    valueCount := 2 // back pixel + event mask

    reqLen := 8 + valueCount // in 4-byte units
    req := make([]byte, reqLen*4)

    req[0] = OpCreateWindow                              // Opcode
    req[1] = c.RootDepth                                 // Depth
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen)) // Request length
    binary.LittleEndian.PutUint32(req[4:], windowID)     // Window ID
    binary.LittleEndian.PutUint32(req[8:], c.RootWindow) // Parent
    binary.LittleEndian.PutUint16(req[12:], uint16(x))   // X
    binary.LittleEndian.PutUint16(req[14:], uint16(y))   // Y
    binary.LittleEndian.PutUint16(req[16:], width)       // Width
    binary.LittleEndian.PutUint16(req[18:], height)      // Height
    binary.LittleEndian.PutUint16(req[20:], 0)           // Border width
    binary.LittleEndian.PutUint16(req[22:], WindowClassInputOutput) // Class
    binary.LittleEndian.PutUint32(req[24:], c.RootVisual) // Visual
    binary.LittleEndian.PutUint32(req[28:], valueMask)   // Value mask

    // Values (in order of bit position in mask)
    binary.LittleEndian.PutUint32(req[32:], 0x00000000) // Back pixel (black)
    binary.LittleEndian.PutUint32(req[36:], eventMask) // Event mask

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    return windowID, nil
}

// MapWindow makes a window visible
func (c *Connection) MapWindow(windowID uint32) error {
    req := make([]byte, 8)
    req[0] = OpMapWindow
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2) // Length
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

## Chapter 3.3: Display a Window

Create `examples/window/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
    conn, err := x11.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    windowID, err := conn.CreateWindow(100, 100, 800, 600)
    if err != nil {
        log.Fatal(err)
    }

    if err := conn.MapWindow(windowID); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Window created: 0x%x\n", windowID)
    fmt.Println("Sleeping for 5 seconds...")

    // Keep the program running so we can see the window
    time.Sleep(5 * time.Second)

    conn.DestroyWindow(windowID)
    fmt.Println("Window destroyed")
}
```

Run it:

```bash
go run examples/window/main.go
```

You should see a black window appear for 5 seconds!

**Exercise 3.3.1**: Modify the background color (hint: change the back pixel value).

**Exercise 3.3.2**: Create multiple windows at different positions.

---

# Part 4: Event Handling

## Chapter 4.1: Understanding X11 Events

X11 sends events as 32-byte packets. The first byte indicates the event type.

Create `internal/x11/event.go`:

```go
package x11

import (
    "encoding/binary"
    "io"
)

// Event represents an X11 event
type Event interface {
    Type() int
}

// KeyEvent represents a key press or release
type KeyEvent struct {
    EventType int
    Keycode   uint8
    State     uint16 // Modifier state
    X, Y      int16
    RootX     int16
    RootY     int16
}

func (e KeyEvent) Type() int { return e.EventType }

// ButtonEvent represents a mouse button press or release
type ButtonEvent struct {
    EventType int
    Button    uint8
    State     uint16
    X, Y      int16
    RootX     int16
    RootY     int16
}

func (e ButtonEvent) Type() int { return e.EventType }

// MotionEvent represents mouse movement
type MotionEvent struct {
    X, Y  int16
    RootX int16
    RootY int16
    State uint16
}

func (e MotionEvent) Type() int { return EventMotionNotify }

// ExposeEvent represents a window expose (needs redraw)
type ExposeEvent struct {
    Window uint32
    X, Y   uint16
    Width  uint16
    Height uint16
    Count  uint16 // Number of expose events to follow
}

func (e ExposeEvent) Type() int { return EventExpose }

// ConfigureEvent represents window resize/move
type ConfigureEvent struct {
    Window uint32
    X, Y   int16
    Width  uint16
    Height uint16
}

func (e ConfigureEvent) Type() int { return EventConfigureNotify }

// ClientMessageEvent represents a client message (like WM_DELETE_WINDOW)
type ClientMessageEvent struct {
    Window uint32
    Format uint8
    Type   uint32
    Data   [20]byte
}

func (e ClientMessageEvent) Type() int { return EventClientMessage }

// UnknownEvent for events we don't handle
type UnknownEvent struct {
    EventType int
    Data      [32]byte
}

func (e UnknownEvent) Type() int { return e.EventType }

// NextEvent reads the next event from the connection
func (c *Connection) NextEvent() (Event, error) {
    buf := make([]byte, 32)
    _, err := io.ReadFull(c.conn, buf)
    if err != nil {
        return nil, err
    }

    // Event type is in the first byte (mask off the high bit)
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
            Window: binary.LittleEndian.Uint32(buf[4:8]),
            Format: buf[1],
            Type:   binary.LittleEndian.Uint32(buf[8:12]),
        }
        copy(e.Data[:], buf[12:32])
        return e, nil

    default:
        e := UnknownEvent{EventType: eventType}
        copy(e.Data[:], buf)
        return e, nil
    }
}
```

## Chapter 4.2: Event Loop Example

Create `examples/events/main.go`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
    conn, err := x11.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    windowID, err := conn.CreateWindow(100, 100, 800, 600)
    if err != nil {
        log.Fatal(err)
    }

    if err := conn.MapWindow(windowID); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Window created. Press keys, click mouse, or close window...")
    fmt.Println("Press 'q' to quit (keycode 24 on most systems)")

    // Event loop
    for {
        event, err := conn.NextEvent()
        if err != nil {
            log.Printf("Error reading event: %v", err)
            break
        }

        switch e := event.(type) {
        case x11.KeyEvent:
            action := "pressed"
            if e.EventType == x11.EventKeyRelease {
                action = "released"
            }
            fmt.Printf("Key %s: keycode=%d at (%d, %d)\n",
                action, e.Keycode, e.X, e.Y)

            // 'q' is usually keycode 24
            if e.EventType == x11.EventKeyPress && e.Keycode == 24 {
                fmt.Println("Quitting...")
                return
            }

        case x11.ButtonEvent:
            action := "pressed"
            if e.EventType == x11.EventButtonRelease {
                action = "released"
            }
            fmt.Printf("Button %d %s at (%d, %d)\n",
                e.Button, action, e.X, e.Y)

        case x11.MotionEvent:
            fmt.Printf("Mouse moved to (%d, %d)\n", e.X, e.Y)

        case x11.ExposeEvent:
            fmt.Printf("Expose: region (%d,%d) %dx%d\n",
                e.X, e.Y, e.Width, e.Height)

        case x11.ConfigureEvent:
            fmt.Printf("Window resized to %dx%d\n", e.Width, e.Height)
        }
    }
}
```

**Exercise 4.2.1**: Modify the program to track mouse position only when a button is held down.

**Exercise 4.2.2**: Count how many times each key is pressed and print statistics on quit.

---

# Part 5: Drawing Graphics

## Chapter 5.1: Graphics Context

X11 requires a Graphics Context (GC) for drawing operations.

Add to `internal/x11/draw.go`:

```go
package x11

import (
    "encoding/binary"
)

// Graphics Context value masks
const (
    GCFunction   = 1 << 0
    GCForeground = 1 << 2
    GCBackground = 1 << 3
    GCLineWidth  = 1 << 4
    GCGraphicsExposures = 1 << 16
)

// CreateGC creates a graphics context
func (c *Connection) CreateGC(drawable uint32) (uint32, error) {
    gcID := c.GenerateID()

    valueMask := uint32(GCForeground | GCBackground | GCGraphicsExposures)
    valueCount := 3

    reqLen := 4 + valueCount
    req := make([]byte, reqLen*4)

    req[0] = OpCreateGC
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], gcID)
    binary.LittleEndian.PutUint32(req[8:], drawable)
    binary.LittleEndian.PutUint32(req[12:], valueMask)
    binary.LittleEndian.PutUint32(req[16:], 0xFFFFFF) // Foreground (white)
    binary.LittleEndian.PutUint32(req[20:], 0x000000) // Background (black)
    binary.LittleEndian.PutUint32(req[24:], 0)        // Graphics exposures off

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    return gcID, nil
}

// FreeGC frees a graphics context
func (c *Connection) FreeGC(gcID uint32) error {
    req := make([]byte, 8)
    req[0] = OpFreeGC
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2)
    binary.LittleEndian.PutUint32(req[4:], gcID)

    _, err := c.conn.Write(req)
    return err
}

// Rectangle represents a rectangle for drawing
type Rectangle struct {
    X, Y          int16
    Width, Height uint16
}

// FillRectangles fills rectangles with the current foreground color
func (c *Connection) FillRectangles(drawable, gc uint32, rects []Rectangle) error {
    reqLen := 3 + len(rects)*2 // header + rectangles (2 units each)
    req := make([]byte, reqLen*4)

    req[0] = OpPolyFillRect
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], drawable)
    binary.LittleEndian.PutUint32(req[8:], gc)

    offset := 12
    for _, r := range rects {
        binary.LittleEndian.PutUint16(req[offset:], uint16(r.X))
        binary.LittleEndian.PutUint16(req[offset+2:], uint16(r.Y))
        binary.LittleEndian.PutUint16(req[offset+4:], r.Width)
        binary.LittleEndian.PutUint16(req[offset+6:], r.Height)
        offset += 8
    }

    _, err := c.conn.Write(req)
    return err
}

// PutImage sends pixel data to a drawable
func (c *Connection) PutImage(drawable, gc uint32, width, height uint16,
    dstX, dstY int16, depth uint8, data []byte) error {

    // Calculate padding to 4-byte boundary
    dataLen := len(data)
    padding := (4 - (dataLen % 4)) % 4

    reqLen := 6 + (dataLen+padding)/4
    req := make([]byte, reqLen*4)

    req[0] = OpPutImage
    req[1] = ImageFormatZPixmap // Format
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], drawable)
    binary.LittleEndian.PutUint32(req[8:], gc)
    binary.LittleEndian.PutUint16(req[12:], width)
    binary.LittleEndian.PutUint16(req[14:], height)
    binary.LittleEndian.PutUint16(req[16:], uint16(dstX))
    binary.LittleEndian.PutUint16(req[18:], uint16(dstY))
    req[20] = 0 // Left pad
    req[21] = depth
    binary.LittleEndian.PutUint16(req[22:], 0) // Unused

    copy(req[24:], data)

    _, err := c.conn.Write(req)
    return err
}
```

## Chapter 5.2: Framebuffer

Create `internal/x11/framebuffer.go`:

```go
package x11

// Framebuffer represents a pixel buffer
type Framebuffer struct {
    Width  int
    Height int
    Pixels []byte // BGRA format (4 bytes per pixel for 24-bit depth)
}

// NewFramebuffer creates a new framebuffer
func NewFramebuffer(width, height int) *Framebuffer {
    return &Framebuffer{
        Width:  width,
        Height: height,
        Pixels: make([]byte, width*height*4),
    }
}

// Clear fills the framebuffer with a color
func (fb *Framebuffer) Clear(r, g, b uint8) {
    for i := 0; i < len(fb.Pixels); i += 4 {
        fb.Pixels[i] = b   // Blue
        fb.Pixels[i+1] = g // Green
        fb.Pixels[i+2] = r // Red
        fb.Pixels[i+3] = 0 // Unused (alpha)
    }
}

// SetPixel sets a pixel at (x, y)
func (fb *Framebuffer) SetPixel(x, y int, r, g, b uint8) {
    if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
        return
    }
    offset := (y*fb.Width + x) * 4
    fb.Pixels[offset] = b
    fb.Pixels[offset+1] = g
    fb.Pixels[offset+2] = r
    fb.Pixels[offset+3] = 0
}

// GetPixel gets the color at (x, y)
func (fb *Framebuffer) GetPixel(x, y int) (r, g, b uint8) {
    if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
        return 0, 0, 0
    }
    offset := (y*fb.Width + x) * 4
    return fb.Pixels[offset+2], fb.Pixels[offset+1], fb.Pixels[offset]
}

// DrawRect draws a filled rectangle
func (fb *Framebuffer) DrawRect(x, y, width, height int, r, g, b uint8) {
    for dy := 0; dy < height; dy++ {
        for dx := 0; dx < width; dx++ {
            fb.SetPixel(x+dx, y+dy, r, g, b)
        }
    }
}

// DrawLine draws a line using Bresenham's algorithm
func (fb *Framebuffer) DrawLine(x0, y0, x1, y1 int, r, g, b uint8) {
    dx := abs(x1 - x0)
    dy := -abs(y1 - y0)
    sx := 1
    if x0 > x1 {
        sx = -1
    }
    sy := 1
    if y0 > y1 {
        sy = -1
    }
    err := dx + dy

    for {
        fb.SetPixel(x0, y0, r, g, b)
        if x0 == x1 && y0 == y1 {
            break
        }
        e2 := 2 * err
        if e2 >= dy {
            err += dy
            x0 += sx
        }
        if e2 <= dx {
            err += dx
            y0 += sy
        }
    }
}

// DrawCircle draws a circle using the midpoint algorithm
func (fb *Framebuffer) DrawCircle(cx, cy, radius int, r, g, b uint8) {
    x := radius
    y := 0
    err := 0

    for x >= y {
        fb.SetPixel(cx+x, cy+y, r, g, b)
        fb.SetPixel(cx+y, cy+x, r, g, b)
        fb.SetPixel(cx-y, cy+x, r, g, b)
        fb.SetPixel(cx-x, cy+y, r, g, b)
        fb.SetPixel(cx-x, cy-y, r, g, b)
        fb.SetPixel(cx-y, cy-x, r, g, b)
        fb.SetPixel(cx+y, cy-x, r, g, b)
        fb.SetPixel(cx+x, cy-y, r, g, b)

        y++
        err += 1 + 2*y
        if 2*(err-x)+1 > 0 {
            x--
            err += 1 - 2*x
        }
    }
}

// FillCircle draws a filled circle
func (fb *Framebuffer) FillCircle(cx, cy, radius int, r, g, b uint8) {
    for y := -radius; y <= radius; y++ {
        for x := -radius; x <= radius; x++ {
            if x*x+y*y <= radius*radius {
                fb.SetPixel(cx+x, cy+y, r, g, b)
            }
        }
    }
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

## Chapter 5.3: Drawing Example

Create `examples/drawing/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "math"
    "time"

    "github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
    conn, err := x11.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    width, height := uint16(800), uint16(600)

    windowID, err := conn.CreateWindow(100, 100, width, height)
    if err != nil {
        log.Fatal(err)
    }

    gcID, err := conn.CreateGC(windowID)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.FreeGC(gcID)

    if err := conn.MapWindow(windowID); err != nil {
        log.Fatal(err)
    }

    // Create framebuffer
    fb := x11.NewFramebuffer(int(width), int(height))

    fmt.Println("Drawing animation. Press Ctrl+C to quit.")

    // Animation loop
    frame := 0
    for {
        // Clear
        fb.Clear(0, 0, 0)

        // Animated circle
        cx := int(width)/2 + int(100*math.Cos(float64(frame)*0.02))
        cy := int(height)/2 + int(100*math.Sin(float64(frame)*0.02))
        fb.FillCircle(cx, cy, 50, 255, 100, 100)

        // Static shapes
        fb.DrawRect(50, 50, 100, 100, 0, 255, 0)
        fb.DrawLine(0, 0, int(width), int(height), 255, 255, 0)
        fb.DrawCircle(600, 100, 80, 0, 100, 255)

        // Some moving lines
        for i := 0; i < 10; i++ {
            x := (frame + i*50) % int(width)
            fb.DrawLine(x, 0, x, int(height), 50, 50, 50)
        }

        // Blit to screen
        err := conn.PutImage(windowID, gcID, width, height, 0, 0,
            conn.RootDepth, fb.Pixels)
        if err != nil {
            log.Printf("PutImage error: %v", err)
        }

        frame++
        time.Sleep(16 * time.Millisecond) // ~60 FPS
    }
}
```

**Exercise 5.3.1**: Add more shapes: triangles, polygons.

**Exercise 5.3.2**: Make the animation interactive - change colors with key presses.

---

# Part 6: Building the Public API

## Chapter 6.1: Clean API Design

Create `glow.go` in the root directory:

```go
package glow

import (
    "github.com/AchrafSoltani/glow/internal/x11"
)

// Color represents an RGB color
type Color struct {
    R, G, B uint8
}

// Common colors
var (
    Black   = Color{0, 0, 0}
    White   = Color{255, 255, 255}
    Red     = Color{255, 0, 0}
    Green   = Color{0, 255, 0}
    Blue    = Color{0, 0, 255}
    Yellow  = Color{255, 255, 0}
    Cyan    = Color{0, 255, 255}
    Magenta = Color{255, 0, 255}
)

// RGB creates a color from RGB values
func RGB(r, g, b uint8) Color {
    return Color{r, g, b}
}

// Hex creates a color from a hex value (0xRRGGBB)
func Hex(hex uint32) Color {
    return Color{
        R: uint8((hex >> 16) & 0xFF),
        G: uint8((hex >> 8) & 0xFF),
        B: uint8(hex & 0xFF),
    }
}

// Window represents a graphics window
type Window struct {
    conn     *x11.Connection
    windowID uint32
    gcID     uint32
    canvas   *Canvas
    width    int
    height   int
    closed   bool
}

// Canvas represents a drawing surface
type Canvas struct {
    fb *x11.Framebuffer
}

// NewWindow creates a new window
func NewWindow(title string, width, height int) (*Window, error) {
    conn, err := x11.Connect()
    if err != nil {
        return nil, err
    }

    windowID, err := conn.CreateWindow(100, 100, uint16(width), uint16(height))
    if err != nil {
        conn.Close()
        return nil, err
    }

    gcID, err := conn.CreateGC(windowID)
    if err != nil {
        conn.Close()
        return nil, err
    }

    if err := conn.MapWindow(windowID); err != nil {
        conn.Close()
        return nil, err
    }

    fb := x11.NewFramebuffer(width, height)

    return &Window{
        conn:     conn,
        windowID: windowID,
        gcID:     gcID,
        canvas:   &Canvas{fb: fb},
        width:    width,
        height:   height,
    }, nil
}

// Close closes the window
func (w *Window) Close() {
    if !w.closed {
        w.conn.FreeGC(w.gcID)
        w.conn.DestroyWindow(w.windowID)
        w.conn.Close()
        w.closed = true
    }
}

// Canvas returns the drawing canvas
func (w *Window) Canvas() *Canvas {
    return w.canvas
}

// Present displays the canvas contents
func (w *Window) Present() error {
    return w.conn.PutImage(w.windowID, w.gcID,
        uint16(w.width), uint16(w.height), 0, 0,
        w.conn.RootDepth, w.canvas.fb.Pixels)
}

// Width returns the window width
func (w *Window) Width() int { return w.width }

// Height returns the window height
func (w *Window) Height() int { return w.height }

// Canvas drawing methods

// Clear fills the canvas with a color
func (c *Canvas) Clear(color Color) {
    c.fb.Clear(color.R, color.G, color.B)
}

// SetPixel sets a pixel
func (c *Canvas) SetPixel(x, y int, color Color) {
    c.fb.SetPixel(x, y, color.R, color.G, color.B)
}

// DrawRect draws a filled rectangle
func (c *Canvas) DrawRect(x, y, width, height int, color Color) {
    c.fb.DrawRect(x, y, width, height, color.R, color.G, color.B)
}

// DrawLine draws a line
func (c *Canvas) DrawLine(x0, y0, x1, y1 int, color Color) {
    c.fb.DrawLine(x0, y0, x1, y1, color.R, color.G, color.B)
}

// DrawCircle draws a circle outline
func (c *Canvas) DrawCircle(x, y, radius int, color Color) {
    c.fb.DrawCircle(x, y, radius, color.R, color.G, color.B)
}

// FillCircle draws a filled circle
func (c *Canvas) FillCircle(x, y, radius int, color Color) {
    c.fb.FillCircle(x, y, radius, color.R, color.G, color.B)
}
```

## Chapter 6.2: Event System

Add to `glow.go`:

```go
// Event types
type EventType int

const (
    EventQuit EventType = iota
    EventKeyDown
    EventKeyUp
    EventMouseButtonDown
    EventMouseButtonUp
    EventMouseMotion
    EventWindowResize
    EventWindowExpose
)

// Event represents an input event
type Event struct {
    Type    EventType
    Key     Key
    Button  MouseButton
    X, Y    int
    Width   int
    Height  int
}

// Key represents a keyboard key
type Key uint8

// Common key codes (X11 keycodes, will vary by system)
const (
    KeyEscape Key = 9
    KeyQ      Key = 24
    KeyW      Key = 25
    KeyA      Key = 38
    KeyS      Key = 39
    KeyD      Key = 40
    KeySpace  Key = 65
    KeyLeft   Key = 113
    KeyUp     Key = 111
    KeyRight  Key = 114
    KeyDown   Key = 116
)

// MouseButton represents a mouse button
type MouseButton uint8

const (
    MouseLeft   MouseButton = 1
    MouseMiddle MouseButton = 2
    MouseRight  MouseButton = 3
    MouseWheelUp   MouseButton = 4
    MouseWheelDown MouseButton = 5
)

// PollEvent polls for the next event (non-blocking would require select/goroutines)
func (w *Window) PollEvent() (*Event, error) {
    xEvent, err := w.conn.NextEvent()
    if err != nil {
        return nil, err
    }

    switch e := xEvent.(type) {
    case x11.KeyEvent:
        evType := EventKeyDown
        if e.EventType == x11.EventKeyRelease {
            evType = EventKeyUp
        }
        return &Event{
            Type: evType,
            Key:  Key(e.Keycode),
            X:    int(e.X),
            Y:    int(e.Y),
        }, nil

    case x11.ButtonEvent:
        evType := EventMouseButtonDown
        if e.EventType == x11.EventButtonRelease {
            evType = EventMouseButtonUp
        }
        return &Event{
            Type:   evType,
            Button: MouseButton(e.Button),
            X:      int(e.X),
            Y:      int(e.Y),
        }, nil

    case x11.MotionEvent:
        return &Event{
            Type: EventMouseMotion,
            X:    int(e.X),
            Y:    int(e.Y),
        }, nil

    case x11.ExposeEvent:
        return &Event{
            Type: EventWindowExpose,
        }, nil

    case x11.ConfigureEvent:
        return &Event{
            Type:   EventWindowResize,
            Width:  int(e.Width),
            Height: int(e.Height),
        }, nil
    }

    return nil, nil // Unknown event
}
```

## Chapter 6.3: Complete Example

Create `examples/game/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/AchrafSoltani/glow"
)

func main() {
    win, err := glow.NewWindow("My Game", 800, 600)
    if err != nil {
        log.Fatal(err)
    }
    defer win.Close()

    // Player state
    playerX, playerY := 400, 300
    playerSpeed := 5

    // Game loop
    running := true
    lastTime := time.Now()

    for running {
        // Delta time (for frame-independent movement)
        now := time.Now()
        _ = now.Sub(lastTime).Seconds()
        lastTime = now

        // Handle events (this is blocking - see next chapter for improvement)
        // For now, we'll use a simple approach

        // Clear screen
        canvas := win.Canvas()
        canvas.Clear(glow.RGB(20, 20, 40))

        // Draw player
        canvas.FillCircle(playerX, playerY, 20, glow.Red)

        // Draw some scenery
        canvas.DrawRect(100, 500, 600, 50, glow.Green)
        canvas.DrawCircle(700, 100, 50, glow.Yellow)

        // Present
        if err := win.Present(); err != nil {
            log.Printf("Present error: %v", err)
        }

        // Frame timing
        time.Sleep(16 * time.Millisecond)
    }

    fmt.Println("Game ended")
}
```

---

# Part 7: Advanced Topics

## Chapter 7.1: Non-Blocking Events

The current event system blocks on `NextEvent()`. For a proper game loop, we need non-blocking I/O.

Add to `internal/x11/conn.go`:

```go
import (
    "time"
)

// SetReadDeadline sets a deadline for reading
func (c *Connection) SetReadDeadline(t time.Time) error {
    return c.conn.SetReadDeadline(t)
}

// HasPendingEvent checks if there's an event waiting (non-blocking)
func (c *Connection) HasPendingEvent() bool {
    // Set a very short deadline
    c.conn.SetReadDeadline(time.Now().Add(time.Microsecond))

    buf := make([]byte, 1)
    _, err := c.conn.Read(buf)

    // Reset deadline
    c.conn.SetReadDeadline(time.Time{})

    if err != nil {
        return false
    }

    // We read a byte, need to handle it
    // This is a simplified approach - real implementation would use select/poll
    return true
}
```

A better approach uses goroutines:

```go
// Add to Window struct
type Window struct {
    // ... existing fields
    events chan Event
    quit   chan struct{}
}

// Start event polling goroutine in NewWindow:
func NewWindow(title string, width, height int) (*Window, error) {
    // ... existing code ...

    w := &Window{
        // ... existing fields ...
        events: make(chan Event, 100),
        quit:   make(chan struct{}),
    }

    // Start event goroutine
    go w.pollEvents()

    return w, nil
}

func (w *Window) pollEvents() {
    for {
        select {
        case <-w.quit:
            return
        default:
            xEvent, err := w.conn.NextEvent()
            if err != nil {
                continue
            }
            // Convert and send to channel
            if event := w.convertEvent(xEvent); event != nil {
                select {
                case w.events <- *event:
                default: // Channel full, drop event
                }
            }
        }
    }
}

// PollEvent returns next event or nil if none available
func (w *Window) PollEvent() *Event {
    select {
    case e := <-w.events:
        return &e
    default:
        return nil
    }
}
```

## Chapter 7.2: Window Properties and WM Integration

For proper window manager integration (close button, title, etc.):

Add to `internal/x11/window.go`:

```go
// Atom represents an X11 atom (interned string)
type Atom uint32

// InternAtom gets or creates an atom
func (c *Connection) InternAtom(name string, onlyIfExists bool) (Atom, error) {
    nameBytes := []byte(name)
    nameLen := len(nameBytes)
    padding := (4 - (nameLen % 4)) % 4

    reqLen := 2 + (nameLen+padding)/4
    req := make([]byte, reqLen*4)

    req[0] = OpInternAtom
    if onlyIfExists {
        req[1] = 1
    }
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint16(req[4:], uint16(nameLen))
    binary.LittleEndian.PutUint16(req[6:], 0) // Unused
    copy(req[8:], nameBytes)

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    // Read reply
    reply := make([]byte, 32)
    if _, err := c.conn.Read(reply); err != nil {
        return 0, err
    }

    atom := binary.LittleEndian.Uint32(reply[8:12])
    return Atom(atom), nil
}

// ChangeProperty sets a window property
func (c *Connection) ChangeProperty(window uint32, property, propType Atom,
    format uint8, data []byte) error {

    dataLen := len(data)
    padding := (4 - (dataLen % 4)) % 4

    reqLen := 6 + (dataLen+padding)/4
    req := make([]byte, reqLen*4)

    req[0] = OpChangeProperty
    req[1] = 0 // Replace mode
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], window)
    binary.LittleEndian.PutUint32(req[8:], uint32(property))
    binary.LittleEndian.PutUint32(req[12:], uint32(propType))
    req[16] = format
    // req[17:20] unused
    binary.LittleEndian.PutUint32(req[20:], uint32(dataLen/(int(format)/8)))
    copy(req[24:], data)

    _, err := c.conn.Write(req)
    return err
}

// SetWindowTitle sets the window title
func (c *Connection) SetWindowTitle(window uint32, title string) error {
    // WM_NAME property
    wmName, err := c.InternAtom("WM_NAME", false)
    if err != nil {
        return err
    }

    stringAtom, err := c.InternAtom("STRING", false)
    if err != nil {
        return err
    }

    return c.ChangeProperty(window, wmName, stringAtom, 8, []byte(title))
}

// EnableCloseEvent sets up WM_DELETE_WINDOW handling
func (c *Connection) EnableCloseEvent(window uint32) (Atom, error) {
    wmProtocols, err := c.InternAtom("WM_PROTOCOLS", false)
    if err != nil {
        return 0, err
    }

    wmDeleteWindow, err := c.InternAtom("WM_DELETE_WINDOW", false)
    if err != nil {
        return 0, err
    }

    atomAtom, err := c.InternAtom("ATOM", false)
    if err != nil {
        return 0, err
    }

    data := make([]byte, 4)
    binary.LittleEndian.PutUint32(data, uint32(wmDeleteWindow))

    err = c.ChangeProperty(window, wmProtocols, atomAtom, 32, data)
    return wmDeleteWindow, err
}
```

## Chapter 7.3: Performance Optimization

### MIT-SHM Extension (Shared Memory)

For faster blitting, X11 supports shared memory. This is more complex but much faster:

```go
// This requires the MIT-SHM extension and syscalls
// Simplified concept:

import "golang.org/x/sys/unix"

func createSharedMemory(size int) ([]byte, int, error) {
    // Create shared memory segment
    shmid, _, errno := unix.Syscall(unix.SYS_SHMGET,
        unix.IPC_PRIVATE,
        uintptr(size),
        0600|unix.IPC_CREAT)
    if errno != 0 {
        return nil, 0, errno
    }

    // Attach shared memory
    addr, _, errno := unix.Syscall(unix.SYS_SHMAT, shmid, 0, 0)
    if errno != 0 {
        return nil, 0, errno
    }

    // Create byte slice from memory
    // (careful with unsafe here)

    return data, int(shmid), nil
}
```

### Double Buffering

Create a back buffer and swap:

```go
type DoubleBufferedWindow struct {
    frontBuffer *Canvas
    backBuffer  *Canvas
}

func (w *DoubleBufferedWindow) Swap() {
    w.frontBuffer, w.backBuffer = w.backBuffer, w.frontBuffer
}
```

---

# Part 8: Projects and Exercises

## Project 1: Pong Clone

Build a simple Pong game:
- Two paddles controlled by keys
- Ball that bounces
- Score display (draw numbers with rectangles)

## Project 2: Paint Program

Build a simple paint program:
- Draw with mouse
- Color selection (keyboard shortcuts)
- Clear canvas
- Line and rectangle tools

## Project 3: Particle System

Create a particle system:
- Spawn particles on click
- Physics (gravity, velocity)
- Color fading
- Particle pooling for performance

## Project 4: Sprite System

Implement sprite loading and rendering:
- Load BMP or TGA files (simple formats)
- Draw sprites with transparency
- Sprite animation

---

# Appendix A: X11 Protocol Reference

## Request Format

```
┌────────┬────────┬────────────────┬─────────────────────────────┐
│ Opcode │  Data  │ Request Length │          Payload            │
│ 1 byte │ 1 byte │    2 bytes     │    (length-1)*4 bytes       │
└────────┴────────┴────────────────┴─────────────────────────────┘
```

## Event Format

All events are 32 bytes:

```
┌────────┬────────────────────────────────────────────────────────┐
│  Type  │                    Event Data                          │
│ 1 byte │                     31 bytes                           │
└────────┴────────────────────────────────────────────────────────┘
```

## Common Opcodes

| Opcode | Request |
|--------|---------|
| 1 | CreateWindow |
| 4 | DestroyWindow |
| 8 | MapWindow |
| 10 | UnmapWindow |
| 16 | InternAtom |
| 18 | ChangeProperty |
| 55 | CreateGC |
| 60 | FreeGC |
| 70 | PolyFillRectangle |
| 72 | PutImage |

---

# Appendix B: Debugging Tips

## View X11 Traffic

Use `xtrace` to see X11 protocol messages:

```bash
sudo apt install xtrace
xtrace -n ./your_program
```

## Check X11 Connection

```bash
xdpyinfo | head -20
```

## Common Errors

1. **"Connection refused"**: X server not running or DISPLAY not set
2. **"Bad window"**: Window ID is invalid or window was destroyed
3. **"Bad GC"**: Graphics context is invalid
4. **Black window**: Forgot to set event mask or handle Expose

---

# Appendix C: Resources

## X11 Documentation
- X11 Protocol Specification: https://x.org/releases/current/doc/xproto/x11protocol.html
- Xlib Manual: https://tronche.com/gui/x/xlib/

## Go Resources
- Effective Go: https://go.dev/doc/effective_go
- Go Binary Package: https://pkg.go.dev/encoding/binary

## Reference Implementations
- XGB (Pure Go X11): https://github.com/BurntSushi/xgb
- Shiny: https://github.com/golang/exp/tree/master/shiny

---

# Changelog / Progress Tracker

Use this section to track your progress:

- [ ] Chapter 1: Project Setup
- [ ] Chapter 2: X11 Connection
- [ ] Chapter 3: Window Creation
- [ ] Chapter 4: Event Handling
- [ ] Chapter 5: Drawing Graphics
- [ ] Chapter 6: Public API
- [ ] Chapter 7: Advanced Topics
- [ ] Project 1: Pong
- [ ] Project 2: Paint
- [ ] Project 3: Particles
- [ ] Project 4: Sprites
