# Glow Tutorial Index

## Completed Steps (Implemented)

### Part 1: Foundation
- [00 - Overview](00-overview.md) - Introduction and architecture
- [01 - Project Setup](01-project-setup.md) - Go modules and directory structure
- [02 - Go Fundamentals](02-go-fundamentals.md) - Binary encoding, slices, sockets

### Part 2: X11 Basics
- [03 - X11 Protocol](03-x11-protocol.md) - Understanding the protocol
- [04 - X11 Connection](04-x11-connection.md) - Connecting and handshake
- [05 - Authentication](05-authentication.md) - Xauthority and MIT-MAGIC-COOKIE

### Part 3: Window Management
- [06 - Window Creation](06-window-creation.md) - Creating and mapping windows
- [07 - Window Properties](07-window-properties.md) - Title, close button, atoms

### Part 4: Event Handling
- [08 - Event Handling](08-event-handling.md) - Reading and parsing events
- [09 - Non-Blocking Events](09-nonblocking-events.md) - Goroutines and channels

### Part 5: Graphics
- [10 - Graphics Context](10-graphics-context.md) - GC and PutImage
- [11 - Framebuffer](11-framebuffer.md) - Pixel buffer and drawing primitives
- [12 - Public API](12-public-api.md) - Clean user-facing API design

### Part 6: Projects
- [13 - Pong](13-project-pong.md) - Classic arcade game
- [14 - Paint](14-project-paint.md) - Drawing application
- [15 - Particles](15-project-particles.md) - Physics-based particle system

---

## Extensions (Not Implemented - Guidance Only)

These files explain how to add features but don't provide complete implementations.

### Graphics Extensions
- [20 - Sprites](20-extension-sprites.md) - Image loading and sprite rendering
- [21 - Fonts](21-extension-fonts.md) - Text rendering with bitmap/TTF fonts

### Input/Output
- [22 - Audio](22-extension-audio.md) - Sound effects and music
- [23 - Gamepad](23-extension-gamepad.md) - Controller support

### Performance & Portability
- [24 - MIT-SHM](24-extension-shm.md) - Shared memory for faster rendering
- [25 - Cross-Platform](25-extension-crossplatform.md) - Windows and macOS support

---

## Quick Reference

### Running Examples

```bash
cd /home/achraf/Documents/GolandProjects/AchrafSoltani/glow

# Basic examples
go run examples/connect/main.go
go run examples/window/main.go
go run examples/events/main.go
go run examples/drawing/main.go
go run examples/demo/main.go

# Projects
go run examples/pong/main.go
go run examples/paint/main.go
go run examples/particles/main.go
```

### Project Structure

```
glow/
├── go.mod
├── glow.go              # Public API
├── events.go            # Event types
├── tutorial/            # This tutorial
├── internal/x11/        # X11 implementation
│   ├── auth.go
│   ├── conn.go
│   ├── protocol.go
│   ├── window.go
│   ├── event.go
│   ├── draw.go
│   ├── framebuffer.go
│   └── atoms.go
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

### Key Concepts Learned

| Concept | Where |
|---------|-------|
| Binary encoding | Step 02 |
| Unix sockets | Step 04 |
| Protocol parsing | Step 03-05 |
| Window management | Step 06-07 |
| Event handling | Step 08-09 |
| Goroutines | Step 09 |
| Software rendering | Step 10-11 |
| API design | Step 12 |
| Game loops | Step 13-15 |

### X11 Opcodes Used

| Opcode | Name | Step |
|--------|------|------|
| 1 | CreateWindow | 06 |
| 4 | DestroyWindow | 06 |
| 8 | MapWindow | 06 |
| 10 | UnmapWindow | 06 |
| 16 | InternAtom | 07 |
| 18 | ChangeProperty | 07 |
| 55 | CreateGC | 10 |
| 60 | FreeGC | 10 |
| 72 | PutImage | 10 |
