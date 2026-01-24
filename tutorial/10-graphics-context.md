# Step 10: Graphics Context

## Goal
Create a graphics context for drawing operations.

## What You'll Learn
- What a Graphics Context (GC) is
- Creating and configuring GCs
- GC value masks

## 10.1 What is a Graphics Context?

A GC holds drawing state: foreground color, background color, line width, etc. Instead of passing these with every draw call, you create a GC once and reference it.

```
GC contains:
- Foreground color (for drawing)
- Background color (for fills)
- Line width
- Font
- And 20+ other properties
```

## 10.2 Protocol Constants

Add to `protocol.go`:

```go
const (
    OpCreateGC = 55
    OpFreeGC   = 60
    OpPutImage = 72
)

// GC value masks
const (
    GCFunction          = 1 << 0
    GCPlaneMask         = 1 << 1
    GCForeground        = 1 << 2
    GCBackground        = 1 << 3
    GCLineWidth         = 1 << 4
    GCLineStyle         = 1 << 5
    GCCapStyle          = 1 << 6
    GCJoinStyle         = 1 << 7
    GCFillStyle         = 1 << 8
    GCGraphicsExposures = 1 << 16
)

// Image formats
const (
    ImageFormatBitmap  = 0
    ImageFormatXYPixmap = 1
    ImageFormatZPixmap = 2  // Raw pixels - what we use
)
```

## 10.3 CreateGC Implementation

Create `internal/x11/draw.go`:

```go
package x11

import "encoding/binary"

// CreateGC creates a graphics context for drawing
func (c *Connection) CreateGC(drawable uint32) (uint32, error) {
    gcID := c.GenerateID()

    // Set foreground, background, and disable graphics exposures
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

    // Values in mask bit order
    binary.LittleEndian.PutUint32(req[16:], 0xFFFFFF) // Foreground: white
    binary.LittleEndian.PutUint32(req[20:], 0x000000) // Background: black
    binary.LittleEndian.PutUint32(req[24:], 0)        // GraphicsExposures: off

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    return gcID, nil
}

// FreeGC releases a graphics context
func (c *Connection) FreeGC(gcID uint32) error {
    req := make([]byte, 8)
    req[0] = OpFreeGC
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 2)
    binary.LittleEndian.PutUint32(req[4:], gcID)

    _, err := c.conn.Write(req)
    return err
}
```

## 10.4 PutImage - The Key Drawing Function

PutImage sends pixel data to a window (or pixmap):

```go
// PutImage sends pixel data to a drawable
func (c *Connection) PutImage(drawable, gc uint32, width, height uint16,
    dstX, dstY int16, depth uint8, data []byte) error {

    // Pad data to 4-byte boundary
    dataLen := len(data)
    padding := (4 - (dataLen % 4)) % 4

    reqLen := 6 + (dataLen+padding)/4
    req := make([]byte, reqLen*4)

    req[0] = OpPutImage
    req[1] = ImageFormatZPixmap  // Raw pixel data
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], drawable)
    binary.LittleEndian.PutUint32(req[8:], gc)
    binary.LittleEndian.PutUint16(req[12:], width)
    binary.LittleEndian.PutUint16(req[14:], height)
    binary.LittleEndian.PutUint16(req[16:], uint16(dstX))
    binary.LittleEndian.PutUint16(req[18:], uint16(dstY))
    req[20] = 0     // Left pad
    req[21] = depth // Bits per pixel
    binary.LittleEndian.PutUint16(req[22:], 0)  // Unused

    copy(req[24:], data)

    _, err := c.conn.Write(req)
    return err
}
```

## 10.5 Understanding ZPixmap Format

For 24-bit depth (most common):
- 4 bytes per pixel: BGRA (Blue, Green, Red, Alpha)
- Alpha is usually ignored
- Little-endian byte order

```go
// Pixel layout for 24-bit depth:
// Offset 0: Blue
// Offset 1: Green
// Offset 2: Red
// Offset 3: Unused (alpha)

func setPixel(pixels []byte, width, x, y int, r, g, b uint8) {
    offset := (y*width + x) * 4
    pixels[offset] = b
    pixels[offset+1] = g
    pixels[offset+2] = r
    pixels[offset+3] = 0
}
```

## 10.6 Usage Example

```go
conn, _ := x11.Connect()
windowID, _ := conn.CreateWindow(100, 100, 800, 600)
gcID, _ := conn.CreateGC(windowID)
conn.MapWindow(windowID)

// Create pixel buffer
width, height := 800, 600
pixels := make([]byte, width*height*4)

// Fill with red
for y := 0; y < height; y++ {
    for x := 0; x < width; x++ {
        offset := (y*width + x) * 4
        pixels[offset] = 0     // Blue
        pixels[offset+1] = 0   // Green
        pixels[offset+2] = 255 // Red
        pixels[offset+3] = 0   // Alpha
    }
}

// Send to window
conn.PutImage(windowID, gcID, uint16(width), uint16(height),
    0, 0, conn.RootDepth, pixels)

// Clean up when done
conn.FreeGC(gcID)
conn.DestroyWindow(windowID)
```

## 10.7 Why Graphics Exposures Off?

We disable `GraphicsExposures` because:
- When enabled, copying causes Expose events
- We're doing software rendering, not server-side copying
- Reduces unnecessary events

## Resource Management

Always free resources when done:

```go
// Create
gcID, _ := conn.CreateGC(windowID)

// Use...

// Free
defer conn.FreeGC(gcID)
```

## 10.8 X11 Request Size Limit (Important!)

X11 requests have a **16-bit length field**, limiting each request to:
```
65535 * 4 = 262,140 bytes maximum
```

For PutImage, the header is 24 bytes, leaving ~262KB for pixel data.

### The Problem

For an 800x600 window at 4 bytes/pixel:
```
800 × 600 × 4 = 1,920,000 bytes
```

This exceeds the limit by 7x! Sending this causes an X11 `Length` error.

### The Solution: Split into Strips

Send the image in horizontal strips that fit within the limit:

```go
func (c *Connection) PutImage(drawable, gc uint32, width, height uint16,
    dstX, dstY int16, depth uint8, data []byte) error {

    bytesPerPixel := 4
    rowBytes := int(width) * bytesPerPixel

    // Max data per request (~262KB - 24 byte header)
    maxDataBytes := 262140 - 24

    // Calculate rows that fit in one request
    rowsPerRequest := maxDataBytes / rowBytes
    if rowsPerRequest > int(height) {
        rowsPerRequest = int(height)
    }

    // Send image in horizontal strips
    for y := 0; y < int(height); y += rowsPerRequest {
        stripHeight := rowsPerRequest
        if y+stripHeight > int(height) {
            stripHeight = int(height) - y
        }

        stripData := data[y*rowBytes : (y+stripHeight)*rowBytes]

        // Send this strip with adjusted Y position
        err := c.putImageStrip(drawable, gc, width, uint16(stripHeight),
            dstX, dstY+int16(y), depth, stripData)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### Example Calculation

For 800x600 at 4 bytes/pixel:
- Row size: 800 × 4 = 3,200 bytes
- Max rows per request: 262,116 / 3,200 ≈ 81 rows
- Number of requests needed: 600 / 81 ≈ 8 requests

### Alternative: BigRequests Extension

X11 has a `BigRequests` extension allowing 32-bit length fields. However, splitting into strips:
- Works on all X servers
- Requires no extension negotiation
- Is simpler to implement

## Next Step
Continue to Step 11 to create a Framebuffer with drawing primitives.
