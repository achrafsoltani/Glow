# Extension: Sprite System

## Overview
Add the ability to load images and draw them as sprites.

## What You'll Need

1. Image loading (BMP or PNG)
2. Sprite structure
3. Blitting with transparency
4. Sprite sheets and animation

## Suggested Implementation

### Sprite Structure

```go
type Sprite struct {
    Width  int
    Height int
    Pixels []byte  // RGBA format

    // For sprite sheets
    FrameWidth  int
    FrameHeight int
    FrameCount  int
}
```

### Loading BMP Files

BMP is simple to parse (no compression):

```go
func LoadBMP(filename string) (*Sprite, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // BMP Header (14 bytes)
    if data[0] != 'B' || data[1] != 'M' {
        return nil, errors.New("not a BMP file")
    }

    pixelOffset := binary.LittleEndian.Uint32(data[10:14])

    // DIB Header
    width := int(binary.LittleEndian.Uint32(data[18:22]))
    height := int(binary.LittleEndian.Uint32(data[22:26]))
    bitsPerPixel := binary.LittleEndian.Uint16(data[28:30])

    // Read pixel data (note: BMP is bottom-up!)
    sprite := &Sprite{
        Width:  width,
        Height: height,
        Pixels: make([]byte, width*height*4),
    }

    // Parse pixels based on bits per pixel...
    // Handle BGR -> RGBA conversion
    // Flip vertically (BMP is upside down)

    return sprite, nil
}
```

### Loading PNG Files

Use Go's standard library:

```go
import "image/png"

func LoadPNG(filename string) (*Sprite, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    img, err := png.Decode(file)
    if err != nil {
        return nil, err
    }

    bounds := img.Bounds()
    sprite := &Sprite{
        Width:  bounds.Dx(),
        Height: bounds.Dy(),
        Pixels: make([]byte, bounds.Dx()*bounds.Dy()*4),
    }

    // Convert to RGBA
    for y := 0; y < bounds.Dy(); y++ {
        for x := 0; x < bounds.Dx(); x++ {
            r, g, b, a := img.At(x, y).RGBA()
            offset := (y*bounds.Dx() + x) * 4
            sprite.Pixels[offset] = uint8(r >> 8)
            sprite.Pixels[offset+1] = uint8(g >> 8)
            sprite.Pixels[offset+2] = uint8(b >> 8)
            sprite.Pixels[offset+3] = uint8(a >> 8)
        }
    }

    return sprite, nil
}
```

### Drawing Sprites

```go
func (c *Canvas) DrawSprite(sprite *Sprite, x, y int) {
    for sy := 0; sy < sprite.Height; sy++ {
        for sx := 0; sx < sprite.Width; sx++ {
            srcOffset := (sy*sprite.Width + sx) * 4
            a := sprite.Pixels[srcOffset+3]

            if a == 0 {
                continue  // Fully transparent
            }

            r := sprite.Pixels[srcOffset]
            g := sprite.Pixels[srcOffset+1]
            b := sprite.Pixels[srcOffset+2]

            if a == 255 {
                // Fully opaque
                c.SetPixel(x+sx, y+sy, RGB(r, g, b))
            } else {
                // Alpha blending
                c.blendPixel(x+sx, y+sy, r, g, b, a)
            }
        }
    }
}

func (c *Canvas) blendPixel(x, y int, r, g, b, a uint8) {
    // Get existing pixel
    er, eg, eb := c.GetPixel(x, y)

    // Alpha blend: result = src * alpha + dst * (1 - alpha)
    alpha := float64(a) / 255
    nr := uint8(float64(r)*alpha + float64(er)*(1-alpha))
    ng := uint8(float64(g)*alpha + float64(eg)*(1-alpha))
    nb := uint8(float64(b)*alpha + float64(eb)*(1-alpha))

    c.SetPixel(x, y, RGB(nr, ng, nb))
}
```

### Sprite Animation

```go
type Animation struct {
    Sprite      *Sprite
    FrameWidth  int
    FrameHeight int
    FrameCount  int
    FrameTime   int  // Frames per animation frame
    CurrentFrame int
    Counter     int
}

func (a *Animation) Update() {
    a.Counter++
    if a.Counter >= a.FrameTime {
        a.Counter = 0
        a.CurrentFrame = (a.CurrentFrame + 1) % a.FrameCount
    }
}

func (c *Canvas) DrawAnimation(anim *Animation, x, y int) {
    // Calculate source rectangle
    frameX := (anim.CurrentFrame % (anim.Sprite.Width / anim.FrameWidth)) * anim.FrameWidth
    frameY := (anim.CurrentFrame / (anim.Sprite.Width / anim.FrameWidth)) * anim.FrameHeight

    // Draw just that frame
    c.DrawSpriteRegion(anim.Sprite, x, y, frameX, frameY, anim.FrameWidth, anim.FrameHeight)
}
```

## Files to Create

1. `sprite.go` - Sprite type and loading
2. `animation.go` - Animation support
3. Add to `canvas.go` - DrawSprite, DrawSpriteRegion methods

## Resources

- BMP Format: https://en.wikipedia.org/wiki/BMP_file_format
- Go image package: https://pkg.go.dev/image
- Alpha blending: https://en.wikipedia.org/wiki/Alpha_compositing
