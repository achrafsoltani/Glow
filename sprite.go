package glow

import (
	"image"
	_ "image/png"
	"io"
	"os"

	"github.com/AchrafSoltani/glow/internal/x11"
)

// Sprite holds pre-converted BGRA pixel data ready for fast blitting.
type Sprite struct {
	data *x11.SpriteData
}

// Width returns the sprite width in pixels.
func (s *Sprite) Width() int { return s.data.Width }

// Height returns the sprite height in pixels.
func (s *Sprite) Height() int { return s.data.Height }

// LoadPNG loads a PNG file from disk and returns a Sprite.
func LoadPNG(path string) (*Sprite, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadPNGFromReader(f)
}

// LoadPNGFromReader decodes a PNG from a reader and returns a Sprite.
func LoadPNGFromReader(r io.Reader) (*Sprite, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return NewSpriteFromImage(img), nil
}

// NewSpriteFromImage converts any image.Image to a Sprite with BGRA pixel data.
// It uses straight (non-premultiplied) alpha.
func NewSpriteFromImage(img image.Image) *Sprite {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	pixels := make([]byte, w*h*4)

	// Fast path for *image.NRGBA — direct byte reorder, no interface overhead
	if nrgba, ok := img.(*image.NRGBA); ok {
		for y := 0; y < h; y++ {
			srcOff := (y+bounds.Min.Y-nrgba.Rect.Min.Y)*nrgba.Stride + (bounds.Min.X-nrgba.Rect.Min.X)*4
			dstOff := y * w * 4
			for x := 0; x < w; x++ {
				// NRGBA is R, G, B, A → convert to BGRA
				pixels[dstOff] = nrgba.Pix[srcOff+2]   // B
				pixels[dstOff+1] = nrgba.Pix[srcOff+1] // G
				pixels[dstOff+2] = nrgba.Pix[srcOff]   // R
				pixels[dstOff+3] = nrgba.Pix[srcOff+3] // A
				srcOff += 4
				dstOff += 4
			}
		}
	} else {
		// Generic path — handles all image types, un-premultiplies alpha
		for y := 0; y < h; y++ {
			dstOff := y * w * 4
			for x := 0; x < w; x++ {
				r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
				if a == 0 {
					dstOff += 4
					continue
				}
				// Un-premultiply: color.RGBA returns premultiplied 16-bit values
				a8 := uint8(a >> 8)
				if a == 0xFFFF {
					pixels[dstOff] = uint8(b >> 8)
					pixels[dstOff+1] = uint8(g >> 8)
					pixels[dstOff+2] = uint8(r >> 8)
				} else {
					pixels[dstOff] = uint8((b * 0xFFFF / a) >> 8)
					pixels[dstOff+1] = uint8((g * 0xFFFF / a) >> 8)
					pixels[dstOff+2] = uint8((r * 0xFFFF / a) >> 8)
				}
				pixels[dstOff+3] = a8
				dstOff += 4
			}
		}
	}

	return &Sprite{
		data: &x11.SpriteData{
			Width:  w,
			Height: h,
			Pixels: pixels,
		},
	}
}

// DrawSprite draws an entire sprite at (x, y) on the canvas with alpha blending.
func (c *Canvas) DrawSprite(s *Sprite, x, y int) {
	c.fb.BlitSprite(s.data, x, y)
}

// DrawSpriteRegion draws a sub-region of a sprite at (x, y) on the canvas.
// The source region is defined by (srcX, srcY, srcW, srcH) within the sprite.
func (c *Canvas) DrawSpriteRegion(s *Sprite, x, y, srcX, srcY, srcW, srcH int) {
	c.fb.BlitSpriteRegion(s.data, x, y, srcX, srcY, srcW, srcH)
}
