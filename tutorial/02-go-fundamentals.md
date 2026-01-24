# Step 2: Go Fundamentals for Graphics

## Goal
Master the Go concepts essential for low-level graphics programming.

## What You'll Learn
- Binary encoding and byte manipulation
- Working with slices and unsafe memory
- Network/socket programming basics

## 2.1 Binary Encoding

X11 uses a binary protocol. You'll need to pack and unpack data.

### Using encoding/binary

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

    // Write different types
    binary.Write(buf, binary.LittleEndian, uint32(12345))
    binary.Write(buf, binary.LittleEndian, uint16(100))
    binary.Write(buf, binary.LittleEndian, uint8(42))

    fmt.Printf("Bytes: % x\n", buf.Bytes())
    // Output: Bytes: 39 30 00 00 64 00 2a

    // Reading it back
    reader := bytes.NewReader(buf.Bytes())
    var num32 uint32
    var num16 uint16
    var num8 uint8

    binary.Read(reader, binary.LittleEndian, &num32)
    binary.Read(reader, binary.LittleEndian, &num16)
    binary.Read(reader, binary.LittleEndian, &num8)

    fmt.Printf("Values: %d, %d, %d\n", num32, num16, num8)
}
```

### Manual Byte Packing (Faster)

```go
// Pack uint32 into byte slice at offset
func packUint32(buf []byte, offset int, value uint32) {
    buf[offset] = byte(value)
    buf[offset+1] = byte(value >> 8)
    buf[offset+2] = byte(value >> 16)
    buf[offset+3] = byte(value >> 24)
}

// Unpack uint32 from byte slice
func unpackUint32(buf []byte, offset int) uint32 {
    return uint32(buf[offset]) |
        uint32(buf[offset+1])<<8 |
        uint32(buf[offset+2])<<16 |
        uint32(buf[offset+3])<<24
}

// Pack uint16
func packUint16(buf []byte, offset int, value uint16) {
    buf[offset] = byte(value)
    buf[offset+1] = byte(value >> 8)
}

// Unpack uint16
func unpackUint16(buf []byte, offset int) uint16 {
    return uint16(buf[offset]) | uint16(buf[offset+1])<<8
}
```

### Using binary.LittleEndian Directly

```go
// Most efficient for our use case
buf := make([]byte, 8)
binary.LittleEndian.PutUint32(buf[0:], 12345)
binary.LittleEndian.PutUint16(buf[4:], 100)

value := binary.LittleEndian.Uint32(buf[0:])
```

## 2.2 Byte Slices

Understanding slices is critical for graphics programming.

```go
// Creating pixel buffer (800x600, 4 bytes per pixel)
width, height := 800, 600
pixels := make([]byte, width*height*4)

// Setting a pixel at (x, y)
func setPixel(pixels []byte, width, x, y int, r, g, b uint8) {
    offset := (y*width + x) * 4
    pixels[offset] = b   // Blue
    pixels[offset+1] = g // Green
    pixels[offset+2] = r // Red
    pixels[offset+3] = 0 // Alpha/padding
}

// Example: fill with red
for y := 0; y < height; y++ {
    for x := 0; x < width; x++ {
        setPixel(pixels, width, x, y, 255, 0, 0)
    }
}
```

## 2.3 Unix Sockets

X11 communicates over Unix domain sockets.

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    // Connect to X11 server (display :0)
    conn, err := net.Dial("unix", "/tmp/.X11-unix/X0")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    fmt.Println("Connected!")

    // Read and write like any other connection
    // conn.Write(data)
    // conn.Read(buffer)
}
```

## 2.4 Bit Manipulation

X11 uses bitmasks extensively.

```go
// Event mask flags
const (
    KeyPressMask      = 1 << 0  // 0x0001
    KeyReleaseMask    = 1 << 1  // 0x0002
    ButtonPressMask   = 1 << 2  // 0x0004
    PointerMotionMask = 1 << 6  // 0x0040
    ExposureMask      = 1 << 15 // 0x8000
)

// Combining masks
eventMask := KeyPressMask | KeyReleaseMask | ButtonPressMask
// Result: 0x0007

// Checking if a flag is set
if eventMask & KeyPressMask != 0 {
    fmt.Println("Key press events enabled")
}
```

## 2.5 Interfaces for Events

```go
// Event interface - all events implement this
type Event interface {
    Type() int
}

// Specific event types
type KeyEvent struct {
    EventType int
    Keycode   uint8
    X, Y      int
}

func (e KeyEvent) Type() int { return e.EventType }

type MouseEvent struct {
    EventType int
    Button    uint8
    X, Y      int
}

func (e MouseEvent) Type() int { return e.EventType }

// Using type switch to handle events
func handleEvent(e Event) {
    switch ev := e.(type) {
    case KeyEvent:
        fmt.Printf("Key: %d\n", ev.Keycode)
    case MouseEvent:
        fmt.Printf("Mouse button: %d at (%d, %d)\n", ev.Button, ev.X, ev.Y)
    }
}
```

## Exercises

1. **Exercise 2.1**: Write a function that packs a struct into bytes:
   ```go
   type Point struct {
       X, Y int16
   }
   func packPoint(p Point) []byte
   ```

2. **Exercise 2.2**: Create a simple pixel buffer and fill it with a gradient.

3. **Exercise 2.3**: Connect to the X11 socket and print that you connected.

## Key Takeaways

- Use `binary.LittleEndian` for X11 (it's little-endian on most systems)
- Byte slices are your main tool for pixel data
- Unix sockets work like TCP connections
- Bitmasks combine multiple flags efficiently

## Next Step
Continue to Step 3 to understand the X11 protocol.
