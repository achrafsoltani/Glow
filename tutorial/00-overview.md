# Glow Tutorial Overview

A step-by-step guide to building a 2D graphics library in pure Go.

## What is Glow?

Glow is a pure Go 2D graphics library inspired by SDL. It provides:
- Window management
- Keyboard and mouse input
- Software rendering (pixel manipulation)
- No cgo dependencies

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Your Application                   │
├─────────────────────────────────────────────────────┤
│                   glow (Public API)                 │
│         Window | Canvas | Events | Colors           │
├─────────────────────────────────────────────────────┤
│                  internal/x11                       │
│   Connection | Protocol | Window | Events | Draw    │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
              X11 Server (Unix Socket)
```

## Tutorial Structure

### Part 1: Foundation (Steps 01-02)
- Project setup
- Go fundamentals for graphics programming

### Part 2: X11 Basics (Steps 03-05)
- Understanding X11 protocol
- Connecting to X11 server
- Authentication

### Part 3: Window Management (Steps 06-07)
- Creating windows
- Window properties and titles

### Part 4: Event Handling (Steps 08-09)
- Reading X11 events
- Non-blocking event loop with goroutines

### Part 5: Graphics (Steps 10-12)
- Graphics context
- Framebuffer and pixel manipulation
- Drawing primitives

### Part 6: Public API (Steps 13-14)
- Clean API design
- Complete event system

### Part 7: Projects (Steps 15-17)
- Pong game
- Paint program
- Particle system

### Part 8: Extensions (Steps 18-25)
- Possible future additions
- Implementation guidance

## Prerequisites

- Linux with X11
- Go 1.21+
- Basic programming knowledge

## File Structure After Completion

```
glow/
├── go.mod
├── glow.go              # Public API
├── events.go            # Event types and handling
├── internal/x11/
│   ├── auth.go          # Xauthority
│   ├── conn.go          # Connection
│   ├── protocol.go      # Constants
│   ├── window.go        # Window management
│   ├── event.go         # Event parsing
│   ├── draw.go          # Drawing operations
│   ├── framebuffer.go   # Pixel buffer
│   └── atoms.go         # Window manager atoms
└── examples/
    ├── connect/
    ├── window/
    ├── events/
    ├── drawing/
    ├── demo/
    ├── pong/
    ├── paint/
    └── particles/
```
