# Extension: Font Rendering

## Overview
Add text rendering with bitmap or TrueType fonts.

## Approaches

1. **Bitmap fonts** - Simple, fast, fixed size
2. **TrueType fonts** - Scalable, complex

## Approach 1: Bitmap Fonts

### Font Structure

```go
type BitmapFont struct {
    CharWidth  int
    CharHeight int
    Chars      [256][]byte  // Bitmap data for each ASCII char
}
```

### Simple Pixel Font (8x8)

```go
// Define characters as 8x8 bitmaps
var font8x8 = map[rune][]byte{
    'A': {
        0b00111100,
        0b01100110,
        0b01100110,
        0b01111110,
        0b01100110,
        0b01100110,
        0b01100110,
        0b00000000,
    },
    'B': {
        0b01111100,
        0b01100110,
        0b01100110,
        0b01111100,
        0b01100110,
        0b01100110,
        0b01111100,
        0b00000000,
    },
    // ... define all characters
}

func (c *Canvas) DrawChar(char rune, x, y int, color Color) {
    bitmap, ok := font8x8[char]
    if !ok {
        return
    }

    for row := 0; row < 8; row++ {
        for col := 0; col < 8; col++ {
            if bitmap[row]&(1<<(7-col)) != 0 {
                c.SetPixel(x+col, y+row, color)
            }
        }
    }
}

func (c *Canvas) DrawText(text string, x, y int, color Color) {
    for i, char := range text {
        c.DrawChar(char, x+i*8, y, color)
    }
}
```

### Loading PSF Fonts (Linux Console Fonts)

```go
// PSF2 header
type PSF2Header struct {
    Magic       [4]byte  // 0x72, 0xb5, 0x4a, 0x86
    Version     uint32
    HeaderSize  uint32
    Flags       uint32
    NumGlyphs   uint32
    BytesPerGlyph uint32
    Height      uint32
    Width       uint32
}

func LoadPSF(filename string) (*BitmapFont, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // Parse header
    var header PSF2Header
    reader := bytes.NewReader(data)
    binary.Read(reader, binary.LittleEndian, &header)

    font := &BitmapFont{
        CharWidth:  int(header.Width),
        CharHeight: int(header.Height),
    }

    // Read glyphs
    glyphData := data[header.HeaderSize:]
    for i := 0; i < int(header.NumGlyphs) && i < 256; i++ {
        offset := i * int(header.BytesPerGlyph)
        font.Chars[i] = glyphData[offset : offset+int(header.BytesPerGlyph)]
    }

    return font, nil
}
```

## Approach 2: TrueType Fonts

Use the `golang.org/x/image/font` package:

```go
import (
    "golang.org/x/image/font"
    "golang.org/x/image/font/opentype"
    "golang.org/x/image/math/fixed"
)

type TTFont struct {
    face font.Face
}

func LoadTTF(filename string, size float64) (*TTFont, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    f, err := opentype.Parse(data)
    if err != nil {
        return nil, err
    }

    face, err := opentype.NewFace(f, &opentype.FaceOptions{
        Size:    size,
        DPI:     72,
        Hinting: font.HintingFull,
    })
    if err != nil {
        return nil, err
    }

    return &TTFont{face: face}, nil
}

func (c *Canvas) DrawTextTTF(font *TTFont, text string, x, y int, color Color) {
    // Create an image.RGBA to draw into
    img := image.NewRGBA(image.Rect(0, 0, c.Width(), c.Height()))

    // Use font.Drawer
    d := &font.Drawer{
        Dst:  img,
        Src:  image.NewUniform(color.toRGBA()),
        Face: font.face,
        Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
    }
    d.DrawString(text)

    // Copy to canvas (only non-transparent pixels)
    // ...
}
```

## Text Alignment

```go
type TextAlign int

const (
    AlignLeft TextAlign = iota
    AlignCenter
    AlignRight
)

func (c *Canvas) DrawTextAligned(text string, x, y, width int, align TextAlign, color Color) {
    textWidth := len(text) * 8  // For 8-pixel-wide font

    switch align {
    case AlignCenter:
        x = x + (width-textWidth)/2
    case AlignRight:
        x = x + width - textWidth
    }

    c.DrawText(text, x, y, color)
}
```

## Formatted Text

```go
func (c *Canvas) DrawTextf(x, y int, color Color, format string, args ...interface{}) {
    text := fmt.Sprintf(format, args...)
    c.DrawText(text, x, y, color)
}

// Usage:
canvas.DrawTextf(10, 10, glow.White, "Score: %d", score)
canvas.DrawTextf(10, 30, glow.White, "FPS: %.1f", fps)
```

## Files to Create

1. `font.go` - BitmapFont type, loading, rendering
2. `font_builtin.go` - Built-in 8x8 font data
3. Optional: `font_ttf.go` - TrueType support

## Resources

- PSF format: https://www.win.tue.nl/~aeb/linux/kbd/font-formats-1.html
- x/image/font: https://pkg.go.dev/golang.org/x/image/font
- Free fonts: https://www.dafont.com/bitmap.php
