# Glow - Project Context

## Overview

Glow is a pure Go 2D graphics library inspired by SDL. It communicates directly with X11 via Unix sockets - no cgo, no external dependencies.

## Architecture

```
glow/
├── glow.go              # Public API (Window, Canvas, Color)
├── events.go            # Event types and polling
├── internal/x11/        # X11 protocol implementation
│   ├── auth.go          # ~/.Xauthority parsing (MIT-MAGIC-COOKIE-1)
│   ├── conn.go          # Connection, handshake, ID generation
│   ├── protocol.go      # X11 opcodes and constants
│   ├── window.go        # CreateWindow, MapWindow, DestroyWindow
│   ├── event.go         # Event parsing (key, mouse, expose)
│   ├── draw.go          # Graphics context, PutImage
│   ├── framebuffer.go   # Software rendering primitives
│   └── atoms.go         # Window properties (title, close button)
├── examples/            # Demo programs
└── tutorial/            # Step-by-step build guide
```

## Key Technical Details

### X11 Connection
- Unix socket: `/tmp/.X11-unix/X0`
- Authentication via `~/.Xauthority` (MIT-MAGIC-COOKIE-1)
- Binary protocol, little-endian

### Rendering Pipeline
1. Draw to `Framebuffer` (BGRA format, 4 bytes/pixel)
2. Transfer via `PutImage` (X11 opcode 72)
3. Images split into strips if >262KB (X11 request size limit)

### Important Fixes

**PutImage Large Image Fix** (commit 6def569):
- X11 requests limited to 262KB (16-bit length field)
- 800x600 @ 4bpp = 1.9MB - exceeds limit
- Solution: Split into horizontal strips automatically

### Wayland Compatibility
- Works via XWayland
- No special handling needed

## Running Examples

```bash
go run examples/pong/main.go       # Pong game
go run examples/paint/main.go      # Paint app
go run examples/particles/main.go  # Particle system
```

## Common Issues

1. **Black screen**: Was caused by PutImage exceeding size limit (fixed)
2. **Auth errors**: Check `~/.Xauthority` exists and matches display

## Extension Ideas (see tutorial/20-25)
- Sprites (image loading)
- Fonts (text rendering)
- Audio (oto library)
- Gamepad (/dev/input/js*)
- MIT-SHM (zero-copy rendering)
- Cross-platform (Win32, Cocoa)
