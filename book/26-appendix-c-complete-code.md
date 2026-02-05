# Appendix C: Complete Code Listing

The essential files that make up the Glow library.

## C.1 Project Structure

```
glow/
├── go.mod
├── glow.go
├── window.go
├── canvas.go
├── color.go
├── events.go
├── keys.go
└── internal/
    └── x11/
        ├── conn.go
        ├── auth.go
        ├── window.go
        ├── events.go
        ├── draw.go
        └── framebuffer.go
```

## C.2 go.mod

```go
module github.com/AchrafSoltani/glow

go 1.21
```

## C.3 glow.go

```go
// Package glow provides a simple 2D graphics library for Go.
// It creates windows, handles input, and renders pixels using software rendering.
//
// Example:
//
//	win, _ := glow.NewWindow("Hello", 800, 600)
//	defer win.Close()
//
//	for win.IsOpen() {
//	    for e := win.PollEvent(); e != nil; e = win.PollEvent() {
//	        // handle events
//	    }
//	    win.Canvas().Clear(glow.Black)
//	    win.Display()
//	}
package glow
```

## C.4 color.go

```go
package glow

// Color represents an RGBA color.
type Color struct {
    R, G, B, A uint8
}

// Predefined colors
var (
    Black       = Color{0, 0, 0, 255}
    White       = Color{255, 255, 255, 255}
    Red         = Color{255, 0, 0, 255}
    Green       = Color{0, 255, 0, 255}
    Blue        = Color{0, 0, 255, 255}
    Yellow      = Color{255, 255, 0, 255}
    Cyan        = Color{0, 255, 255, 255}
    Magenta     = Color{255, 0, 255, 255}
    Transparent = Color{0, 0, 0, 0}
)

// RGB creates an opaque color from RGB values.
func RGB(r, g, b uint8) Color {
    return Color{r, g, b, 255}
}

// RGBA creates a color with alpha.
func RGBA(r, g, b, a uint8) Color {
    return Color{r, g, b, a}
}

// Hex creates an opaque color from a hex value (0xRRGGBB).
func Hex(hex uint32) Color {
    return Color{
        R: uint8((hex >> 16) & 0xFF),
        G: uint8((hex >> 8) & 0xFF),
        B: uint8(hex & 0xFF),
        A: 255,
    }
}
```

## C.5 internal/x11/conn.go (excerpts)

```go
package x11

import (
    "encoding/binary"
    "fmt"
    "io"
    "net"
    "os"
)

type Connection struct {
    conn          net.Conn
    RootWindow    uint32
    RootDepth     uint8
    RootVisual    uint32
    resourceIDBase uint32
    resourceIDMask uint32
    nextResourceID uint32

    // Cached atoms
    atomWmProtocols    Atom
    atomWmDeleteWindow Atom
    atomNetWmName      Atom
    atomUtf8String     Atom
}

func Connect() (*Connection, error) {
    display := os.Getenv("DISPLAY")
    if display == "" {
        display = ":0"
    }

    socketPath := fmt.Sprintf("/tmp/.X11-unix/X%s", display[1:])
    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return nil, err
    }

    c := &Connection{conn: conn}

    if err := c.authenticate(); err != nil {
        conn.Close()
        return nil, err
    }

    if err := c.handshake(); err != nil {
        conn.Close()
        return nil, err
    }

    if err := c.initAtoms(); err != nil {
        conn.Close()
        return nil, err
    }

    return c, nil
}

func (c *Connection) GenerateID() uint32 {
    id := c.resourceIDBase | c.nextResourceID
    c.nextResourceID++
    return id
}

func (c *Connection) Close() error {
    return c.conn.Close()
}
```

## C.6 internal/x11/framebuffer.go

```go
package x11

type Framebuffer struct {
    Width  int
    Height int
    Pixels []byte
}

func NewFramebuffer(width, height int) *Framebuffer {
    return &Framebuffer{
        Width:  width,
        Height: height,
        Pixels: make([]byte, width*height*4),
    }
}

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

func (fb *Framebuffer) GetPixel(x, y int) (r, g, b uint8) {
    if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
        return 0, 0, 0
    }
    offset := (y*fb.Width + x) * 4
    return fb.Pixels[offset+2], fb.Pixels[offset+1], fb.Pixels[offset]
}

func (fb *Framebuffer) Clear(r, g, b uint8) {
    fb.Pixels[0] = b
    fb.Pixels[1] = g
    fb.Pixels[2] = r
    fb.Pixels[3] = 0

    for filled := 4; filled < len(fb.Pixels); filled *= 2 {
        copy(fb.Pixels[filled:], fb.Pixels[:filled])
    }
}

func (fb *Framebuffer) DrawRect(x, y, width, height int, r, g, b uint8) {
    x0, y0 := max(x, 0), max(y, 0)
    x1, y1 := min(x+width, fb.Width), min(y+height, fb.Height)

    if x0 >= x1 || y0 >= y1 {
        return
    }

    for py := y0; py < y1; py++ {
        offset := (py*fb.Width + x0) * 4
        for px := x0; px < x1; px++ {
            fb.Pixels[offset] = b
            fb.Pixels[offset+1] = g
            fb.Pixels[offset+2] = r
            fb.Pixels[offset+3] = 0
            offset += 4
        }
    }
}

func (fb *Framebuffer) DrawLine(x0, y0, x1, y1 int, r, g, b uint8) {
    dx := abs(x1 - x0)
    dy := -abs(y1 - y0)
    sx, sy := 1, 1
    if x0 > x1 {
        sx = -1
    }
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

## C.7 internal/x11/draw.go

```go
package x11

import "encoding/binary"

const (
    OpPutImage        = 72
    ImageFormatZPixmap = 2
)

func (c *Connection) PutImage(drawable, gc uint32, width, height uint16,
    dstX, dstY int16, depth uint8, data []byte) error {

    bytesPerPixel := 4
    rowBytes := int(width) * bytesPerPixel
    maxDataBytes := 262140 - 24
    rowsPerRequest := maxDataBytes / rowBytes

    if rowsPerRequest > int(height) {
        rowsPerRequest = int(height)
    }
    if rowsPerRequest < 1 {
        rowsPerRequest = 1
    }

    for y := 0; y < int(height); y += rowsPerRequest {
        stripHeight := rowsPerRequest
        if y+stripHeight > int(height) {
            stripHeight = int(height) - y
        }

        stripData := data[y*rowBytes : (y+stripHeight)*rowBytes]
        err := c.putImageStrip(drawable, gc, width, uint16(stripHeight),
            dstX, dstY+int16(y), depth, stripData)
        if err != nil {
            return err
        }
    }

    return nil
}

func (c *Connection) putImageStrip(drawable, gc uint32, width, height uint16,
    dstX, dstY int16, depth uint8, data []byte) error {

    dataLen := len(data)
    padding := (4 - (dataLen % 4)) % 4
    reqLen := 6 + (dataLen+padding)/4

    req := make([]byte, reqLen*4)

    req[0] = OpPutImage
    req[1] = ImageFormatZPixmap
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], drawable)
    binary.LittleEndian.PutUint32(req[8:], gc)
    binary.LittleEndian.PutUint16(req[12:], width)
    binary.LittleEndian.PutUint16(req[14:], height)
    binary.LittleEndian.PutUint16(req[16:], uint16(dstX))
    binary.LittleEndian.PutUint16(req[18:], uint16(dstY))
    req[20] = 0
    req[21] = depth

    copy(req[24:], data)

    _, err := c.conn.Write(req)
    return err
}
```

## C.8 Example: Minimal Window

```go
package main

import (
    "github.com/AchrafSoltani/glow"
)

func main() {
    win, err := glow.NewWindow("Minimal Example", 800, 600)
    if err != nil {
        panic(err)
    }
    defer win.Close()

    canvas := win.Canvas()

    for win.IsOpen() {
        for e := win.PollEvent(); e != nil; e = win.PollEvent() {
            switch e.(type) {
            case glow.CloseEvent:
                win.Close()
            case glow.KeyEvent:
                if e.(glow.KeyEvent).Key == glow.KeyEscape {
                    win.Close()
                }
            }
        }

        canvas.Clear(glow.RGB(30, 30, 50))
        canvas.FillCircle(400, 300, 50, glow.Red)

        win.Display()
    }
}
```

## C.9 Example: Animation

```go
package main

import (
    "math"
    "time"

    "github.com/AchrafSoltani/glow"
)

func main() {
    win, _ := glow.NewWindow("Animation", 800, 600)
    defer win.Close()

    canvas := win.Canvas()
    startTime := time.Now()

    for win.IsOpen() {
        for e := win.PollEvent(); e != nil; e = win.PollEvent() {
            if _, ok := e.(glow.CloseEvent); ok {
                win.Close()
            }
        }

        t := time.Since(startTime).Seconds()

        canvas.Clear(glow.Black)

        // Animated circles
        for i := 0; i < 5; i++ {
            angle := t + float64(i)*math.Pi*2/5
            x := 400 + int(150*math.Cos(angle))
            y := 300 + int(150*math.Sin(angle))

            hue := float64(i) / 5
            color := hsvToRGB(hue, 1, 1)
            canvas.FillCircle(x, y, 30, color)
        }

        win.Display()
        time.Sleep(16 * time.Millisecond)
    }
}

func hsvToRGB(h, s, v float64) glow.Color {
    i := int(h * 6)
    f := h*6 - float64(i)
    p := v * (1 - s)
    q := v * (1 - f*s)
    t := v * (1 - (1-f)*s)

    var r, g, b float64
    switch i % 6 {
    case 0:
        r, g, b = v, t, p
    case 1:
        r, g, b = q, v, p
    case 2:
        r, g, b = p, v, t
    case 3:
        r, g, b = p, q, v
    case 4:
        r, g, b = t, p, v
    case 5:
        r, g, b = v, p, q
    }

    return glow.RGB(uint8(r*255), uint8(g*255), uint8(b*255))
}
```

---

The complete source code is available at:
**https://github.com/AchrafSoltani/glow**
