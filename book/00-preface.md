# Preface

## Why This Book?

Every programmer who has built a game or graphical application has used a graphics library. SDL, SFML, raylib - these tools abstract away the complexity of talking to the operating system. But have you ever wondered what happens beneath the surface?

This book takes you on a journey to build a 2D graphics library from scratch in Go. No external dependencies, no C bindings - just pure Go communicating directly with the X11 display server.

## What You'll Learn

By the end of this book, you will understand:

- **Binary protocols**: How to encode and decode binary data to communicate with system services
- **X11 architecture**: The client-server model that powers Linux graphics
- **Window management**: Creating, positioning, and decorating windows
- **Event systems**: Handling keyboard, mouse, and window events
- **Software rendering**: Building a framebuffer and implementing drawing primitives
- **API design**: Creating clean, user-friendly interfaces
- **Systems programming**: Working with sockets, file descriptors, and system calls

## Who This Book Is For

This book is for programmers who:

- Know the basics of Go (or another C-like language)
- Are curious about how things work "under the hood"
- Want to understand graphics programming at a fundamental level
- Enjoy building things from first principles

You don't need prior experience with graphics programming or X11.

## How to Read This Book

The book is structured as a progressive tutorial. Each chapter builds on the previous one, and by the end, you'll have a working graphics library capable of running games.

**Part I: Foundation** covers Go fundamentals relevant to systems programming - binary encoding, byte manipulation, and network sockets.

**Part II: X11 Protocol** dives deep into the X Window System, from connection handshakes to authentication.

**Part III: Window Management** teaches you to create and control windows.

**Part IV: Event Handling** covers input processing and the event loop.

**Part V: Graphics** builds the rendering pipeline, from pixels to primitives.

**Part VI: API Design** wraps everything in a clean, user-friendly interface.

**Part VII: Projects** puts it all together with real applications.

**Part VIII: Extensions** explores advanced topics for those who want to go further.

## The Library We'll Build

Our library, called **Glow**, will support:

- Window creation with titles and close buttons
- Keyboard and mouse input
- A canvas for drawing pixels, lines, rectangles, circles, and triangles
- Predefined and custom colors
- Non-blocking event handling

By the final chapter, you'll use Glow to build Pong, a paint application, and a particle system.

## Source Code

All code from this book is available at:

```
github.com/AchrafSoltani/glow
```

Each chapter corresponds to a step in the `tutorial/` folder.

## Conventions

Code examples appear in monospace:

```go
func main() {
    fmt.Println("Hello, Glow!")
}
```

Terminal commands are prefixed with `$`:

```bash
$ go run main.go
```

Important notes appear in callout boxes:

> **Note**: X11 uses little-endian byte order on most systems.

## Acknowledgments

This book was written with assistance from Claude, an AI assistant by Anthropic.

---

Let's begin.
