# Step 12: Public API Design

## Goal
Create a clean, user-friendly public API that hides X11 complexity.

## What You'll Learn
- API design principles
- Encapsulation
- Type design

## 12.1 Design Goals

1. **Simple**: User shouldn't need to understand X11
2. **Safe**: Hide resource management
3. **Idiomatic**: Follow Go conventions
4. **Flexible**: Allow advanced use when needed

## 12.2 Color Type

Create `glow.go`:

```go
// Package glow provides a pure Go 2D graphics library.
package glow

import "github.com/yourusername/glow/internal/x11"

// Color represents an RGB color
type Color struct {
    R, G, B uint8
}

// Predefined colors
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

// RGB creates a color from components
func RGB(r, g, b uint8) Color {
    return Color{r, g, b}
}

// Hex creates a color from hex value (0xRRGGBB)
func Hex(hex uint32) Color {
    return Color{
        R: uint8((hex >> 16) & 0xFF),
        G: uint8((hex >> 8) & 0xFF),
        B: uint8(hex & 0xFF),
    }
}
```

## 12.3 Window Type

```go
// Window represents a graphics window
type Window struct {
    conn     *x11.Connection
    windowID uint32
    gcID     uint32
    canvas   *Canvas
    width    int
    height   int
    closed   bool

    // Event handling
    eventChan chan Event
    quitChan  chan struct{}
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
        conn.DestroyWindow(windowID)
        conn.Close()
        return nil, err
    }

    conn.SetWindowTitle(windowID, title)
    conn.EnableCloseButton(windowID)
    conn.MapWindow(windowID)

    fb := x11.NewFramebuffer(width, height)

    w := &Window{
        conn:      conn,
        windowID:  windowID,
        gcID:      gcID,
        canvas:    &Canvas{fb: fb},
        width:     width,
        height:    height,
        eventChan: make(chan Event, 256),
        quitChan:  make(chan struct{}),
    }

    go w.pollEvents()

    return w, nil
}

// Close closes the window
func (w *Window) Close() {
    if w.closed {
        return
    }
    w.closed = true
    close(w.quitChan)
    w.conn.FreeGC(w.gcID)
    w.conn.DestroyWindow(w.windowID)
    w.conn.Close()
}

// Width returns window width
func (w *Window) Width() int { return w.width }

// Height returns window height
func (w *Window) Height() int { return w.height }

// Canvas returns the drawing surface
func (w *Window) Canvas() *Canvas { return w.canvas }

// Present copies canvas to screen
func (w *Window) Present() error {
    return w.conn.PutImage(w.windowID, w.gcID,
        uint16(w.width), uint16(w.height), 0, 0,
        w.conn.RootDepth, w.canvas.fb.Pixels)
}
```

## 12.4 Canvas Type

```go
// Canvas is the drawing surface
type Canvas struct {
    fb *x11.Framebuffer
}

// Clear fills canvas with color
func (c *Canvas) Clear(color Color) {
    c.fb.Clear(color.R, color.G, color.B)
}

// SetPixel sets a single pixel
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

// Width returns canvas width
func (c *Canvas) Width() int { return c.fb.Width }

// Height returns canvas height
func (c *Canvas) Height() int { return c.fb.Height }
```

## 12.5 Event Types

Create `events.go`:

```go
package glow

// EventType identifies event kinds
type EventType int

const (
    EventNone EventType = iota
    EventQuit
    EventKeyDown
    EventKeyUp
    EventMouseButtonDown
    EventMouseButtonUp
    EventMouseMotion
)

// Event represents user input
type Event struct {
    Type   EventType
    Key    Key
    Button MouseButton
    X, Y   int
}

// Key represents keyboard keys
type Key uint8

const (
    KeyEscape Key = 9
    KeySpace  Key = 65
    KeyW      Key = 25
    KeyA      Key = 38
    KeyS      Key = 39
    KeyD      Key = 40
    KeyUp     Key = 111
    KeyDown   Key = 116
    KeyLeft   Key = 113
    KeyRight  Key = 114
)

// MouseButton represents mouse buttons
type MouseButton uint8

const (
    MouseLeft   MouseButton = 1
    MouseMiddle MouseButton = 2
    MouseRight  MouseButton = 3
)
```

## 12.6 Event Methods

```go
// PollEvent returns next event or nil
func (w *Window) PollEvent() *Event {
    select {
    case e := <-w.eventChan:
        return &e
    default:
        return nil
    }
}

// WaitEvent blocks for next event
func (w *Window) WaitEvent() *Event {
    e := <-w.eventChan
    return &e
}
```

## 12.7 Usage Example

```go
package main

import (
    "log"
    "time"
    "github.com/yourusername/glow"
)

func main() {
    win, err := glow.NewWindow("My App", 800, 600)
    if err != nil {
        log.Fatal(err)
    }
    defer win.Close()

    running := true
    for running {
        // Events
        for event := win.PollEvent(); event != nil; event = win.PollEvent() {
            switch event.Type {
            case glow.EventQuit:
                running = false
            case glow.EventKeyDown:
                if event.Key == glow.KeyEscape {
                    running = false
                }
            }
        }

        // Draw
        canvas := win.Canvas()
        canvas.Clear(glow.RGB(30, 30, 50))
        canvas.FillCircle(400, 300, 100, glow.Red)
        win.Present()

        time.Sleep(16 * time.Millisecond)
    }
}
```

## Design Principles Applied

1. **Hide complexity**: User never sees X11 types
2. **Resource safety**: Window.Close() handles everything
3. **Fluent API**: canvas.Clear(color).DrawRect(...)
4. **Sensible defaults**: Window position, event masks, etc.

## Next Step
Continue to Step 13-17 for project examples (Pong, Paint, Particles).
