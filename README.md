# Glow

A pure Go 2D graphics library inspired by SDL. No cgo, no external dependencies - just Go and the X11 protocol.

## Features

- **Pure Go** - No cgo, communicates directly with X11 via Unix sockets
- **Window Management** - Create windows with titles and close buttons
- **Input Handling** - Keyboard, mouse buttons, and mouse motion events
- **Software Rendering** - Pixel-level control with a framebuffer
- **Drawing Primitives** - Lines, rectangles, circles, triangles
- **Non-blocking Events** - Goroutine-based event polling

## Installation

```bash
go get github.com/AchrafSoltani/glow
```

## Quick Start

```go
package main

import (
    "github.com/AchrafSoltani/glow"
)

func main() {
    // Create a window
    win, err := glow.NewWindow("Hello Glow", 800, 600)
    if err != nil {
        panic(err)
    }
    defer win.Close()

    canvas := win.Canvas()
    running := true

    for running {
        // Handle events
        for e := win.PollEvent(); e != nil; e = win.PollEvent() {
            switch e.(type) {
            case *glow.QuitEvent:
                running = false
            }
        }

        // Draw
        canvas.Clear(glow.Black)
        canvas.FillCircle(400, 300, 50, glow.Red)
        canvas.DrawRect(100, 100, 200, 150, glow.Blue)

        // Present to screen
        win.Present()
    }
}
```

## API Reference

### Window

```go
// Create a new window
win, err := glow.NewWindow(title string, width, height int) (*Window, error)

// Window methods
win.Close()                    // Close the window
win.Width() int                // Get width
win.Height() int               // Get height
win.Canvas() *Canvas           // Get drawing surface
win.Present() error            // Display the canvas
win.PollEvent() Event          // Get next event (non-blocking)
```

### Canvas Drawing

```go
canvas := win.Canvas()

canvas.Clear(color)                              // Fill with color
canvas.SetPixel(x, y, color)                     // Set single pixel
canvas.GetPixel(x, y) Color                      // Get pixel color
canvas.DrawRect(x, y, w, h, color)               // Filled rectangle
canvas.DrawRectOutline(x, y, w, h, color)        // Rectangle outline
canvas.DrawLine(x0, y0, x1, y1, color)           // Line
canvas.DrawCircle(x, y, radius, color)           // Circle outline
canvas.FillCircle(x, y, radius, color)           // Filled circle
canvas.DrawTriangle(x0, y0, x1, y1, x2, y2, color) // Triangle outline
```

### Colors

```go
// Predefined colors
glow.Black, glow.White, glow.Red, glow.Green, glow.Blue
glow.Yellow, glow.Cyan, glow.Magenta, glow.Orange, glow.Purple, glow.Gray

// Custom colors
color := glow.RGB(255, 128, 0)    // From RGB values
color := glow.Hex(0xFF8000)       // From hex value
```

### Events

```go
for e := win.PollEvent(); e != nil; e = win.PollEvent() {
    switch ev := e.(type) {
    case *glow.QuitEvent:
        // Window close button clicked
    case *glow.KeyPressEvent:
        // ev.Keycode - X11 keycode
    case *glow.KeyReleaseEvent:
        // ev.Keycode
    case *glow.MouseButtonPressEvent:
        // ev.Button, ev.X, ev.Y
    case *glow.MouseButtonReleaseEvent:
        // ev.Button, ev.X, ev.Y
    case *glow.MouseMoveEvent:
        // ev.X, ev.Y
    }
}
```

## Examples

```bash
# Basic examples
go run examples/connect/main.go    # Test X11 connection
go run examples/window/main.go     # Simple window
go run examples/events/main.go     # Event handling
go run examples/drawing/main.go    # Drawing primitives
go run examples/demo/main.go       # Interactive demo

# Projects
go run examples/pong/main.go       # Classic Pong game
go run examples/paint/main.go      # Drawing application
go run examples/particles/main.go  # Particle system
```

## Project Structure

```
glow/
├── glow.go              # Public API
├── events.go            # Event types
├── internal/x11/        # X11 implementation
│   ├── auth.go          # Xauthority parsing
│   ├── conn.go          # Connection management
│   ├── protocol.go      # X11 constants
│   ├── window.go        # Window operations
│   ├── event.go         # Event handling
│   ├── draw.go          # Graphics context
│   ├── framebuffer.go   # Software rendering
│   └── atoms.go         # Window properties
├── examples/            # Example programs
└── tutorial/            # Step-by-step tutorial
```

## Tutorial

The `tutorial/` folder contains a comprehensive guide to building this library from scratch:

- **Steps 00-02**: Foundation (project setup, Go fundamentals)
- **Steps 03-05**: X11 basics (protocol, connection, authentication)
- **Steps 06-07**: Window management (creation, properties)
- **Steps 08-09**: Event handling (parsing, non-blocking)
- **Steps 10-12**: Graphics (rendering, framebuffer, API design)
- **Steps 13-15**: Projects (Pong, Paint, Particles)
- **Steps 20-25**: Extensions (sprites, fonts, audio, gamepad, SHM, cross-platform)

See `tutorial/INDEX.md` for the complete index.

## Requirements

- Linux with X11
- Go 1.21 or later

## How It Works

Glow communicates directly with the X11 server using the X11 protocol over Unix domain sockets (`/tmp/.X11-unix/X0`). It reads authentication cookies from `~/.Xauthority` and implements the binary protocol for:

- Window creation and management
- Event subscription and parsing
- Image transfer (PutImage) for rendering

All rendering is done in software using a framebuffer, then transferred to X11 for display.

## License

MIT
