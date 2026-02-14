package glow

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/AchrafSoltani/glow/internal/x11"
)

// makeTestPNG creates a 4x4 PNG in memory with known pixel values.
// Layout:
//
//	Row 0: fully red (opaque), green (opaque), blue (opaque), transparent
//	Row 1: white 50% alpha, white 100%, black 100%, black 0%
//	Row 2-3: all magenta opaque
func makeTestPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))

	img.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})
	img.SetNRGBA(1, 0, color.NRGBA{0, 255, 0, 255})
	img.SetNRGBA(2, 0, color.NRGBA{0, 0, 255, 255})
	img.SetNRGBA(3, 0, color.NRGBA{0, 0, 0, 0}) // transparent

	img.SetNRGBA(0, 1, color.NRGBA{255, 255, 255, 128})
	img.SetNRGBA(1, 1, color.NRGBA{255, 255, 255, 255})
	img.SetNRGBA(2, 1, color.NRGBA{0, 0, 0, 255})
	img.SetNRGBA(3, 1, color.NRGBA{0, 0, 0, 0})

	for y := 2; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{255, 0, 255, 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestLoadPNGFromReader(t *testing.T) {
	data := makeTestPNG()
	sprite, err := LoadPNGFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("LoadPNGFromReader failed: %v", err)
	}
	if sprite.Width() != 4 || sprite.Height() != 4 {
		t.Fatalf("expected 4x4, got %dx%d", sprite.Width(), sprite.Height())
	}

	// Check a known opaque pixel: (0,0) should be red → BGRA = [0, 0, 255, 255]
	off := 0
	if sprite.data.Pixels[off] != 0 || sprite.data.Pixels[off+1] != 0 ||
		sprite.data.Pixels[off+2] != 255 || sprite.data.Pixels[off+3] != 255 {
		t.Errorf("pixel (0,0) expected BGRA [0,0,255,255], got [%d,%d,%d,%d]",
			sprite.data.Pixels[off], sprite.data.Pixels[off+1],
			sprite.data.Pixels[off+2], sprite.data.Pixels[off+3])
	}

	// Check transparent pixel: (3,0) should have alpha=0
	off = 3 * 4
	if sprite.data.Pixels[off+3] != 0 {
		t.Errorf("pixel (3,0) expected alpha 0, got %d", sprite.data.Pixels[off+3])
	}
}

func TestNewSpriteFromImage_Generic(t *testing.T) {
	// Use image.RGBA (premultiplied) to exercise the generic path
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// Fully opaque green
	img.SetRGBA(0, 0, color.RGBA{0, 255, 0, 255})
	// 50% alpha red (premultiplied: R=128, A=128 → straight R=255)
	img.SetRGBA(1, 0, color.RGBA{128, 0, 0, 128})
	// Fully transparent
	img.SetRGBA(0, 1, color.RGBA{0, 0, 0, 0})
	// Fully opaque blue
	img.SetRGBA(1, 1, color.RGBA{0, 0, 255, 255})

	sprite := NewSpriteFromImage(img)
	if sprite.Width() != 2 || sprite.Height() != 2 {
		t.Fatalf("expected 2x2, got %dx%d", sprite.Width(), sprite.Height())
	}

	// (0,0): green opaque → BGRA [0, 255, 0, 255]
	assertPixel(t, sprite, 0, 0, 0, 255, 0, 255)
	// (1,0): straight red at ~50% alpha → BGRA [0, 0, ~255, 128]
	p := pixelAt(sprite, 1, 0)
	if p[3] != 128 {
		t.Errorf("pixel (1,0) alpha: expected 128, got %d", p[3])
	}
	if p[2] < 250 { // R channel should be ~255 after un-premultiply
		t.Errorf("pixel (1,0) R: expected ~255, got %d", p[2])
	}
	// (0,1): transparent
	if pixelAt(sprite, 0, 1)[3] != 0 {
		t.Errorf("pixel (0,1) expected transparent")
	}
	// (1,1): blue opaque → BGRA [255, 0, 0, 255]
	assertPixel(t, sprite, 1, 1, 255, 0, 0, 255)
}

func TestBlitSprite_FullyOnScreen(t *testing.T) {
	fb := x11.NewFramebuffer(8, 8)
	fb.Clear(0, 0, 0) // black background

	sprite := makeOpaqueRedSprite(2, 2)
	fb.BlitSprite(sprite.data, 3, 3)

	// Check that (3,3) is red: BGRA = [0, 0, 255, 0]
	assertFBPixel(t, fb, 3, 3, 255, 0, 0)
	assertFBPixel(t, fb, 4, 4, 255, 0, 0)
	// Check that (2,2) is still black
	assertFBPixel(t, fb, 2, 2, 0, 0, 0)
}

func TestBlitSprite_PartiallyOffScreen(t *testing.T) {
	fb := x11.NewFramebuffer(8, 8)
	fb.Clear(0, 0, 0)

	sprite := makeOpaqueRedSprite(4, 4)

	// Draw partially off the top-left corner
	fb.BlitSprite(sprite.data, -2, -2)
	// Only the bottom-right 2x2 of the sprite should appear at (0,0)-(1,1)
	assertFBPixel(t, fb, 0, 0, 255, 0, 0)
	assertFBPixel(t, fb, 1, 1, 255, 0, 0)
	assertFBPixel(t, fb, 2, 0, 0, 0, 0) // outside sprite

	// Draw partially off the bottom-right corner
	fb.Clear(0, 0, 0)
	fb.BlitSprite(sprite.data, 6, 6)
	assertFBPixel(t, fb, 6, 6, 255, 0, 0)
	assertFBPixel(t, fb, 7, 7, 255, 0, 0)
	assertFBPixel(t, fb, 5, 5, 0, 0, 0)
}

func TestBlitSprite_FullyOffScreen(t *testing.T) {
	fb := x11.NewFramebuffer(8, 8)
	fb.Clear(100, 100, 100) // grey background

	sprite := makeOpaqueRedSprite(2, 2)

	// Fully off each edge — should not modify framebuffer
	fb.BlitSprite(sprite.data, -10, -10)
	fb.BlitSprite(sprite.data, 100, 100)
	fb.BlitSprite(sprite.data, -10, 4)
	fb.BlitSprite(sprite.data, 4, -10)

	// Verify framebuffer unchanged
	assertFBPixel(t, fb, 0, 0, 100, 100, 100)
	assertFBPixel(t, fb, 4, 4, 100, 100, 100)
}

func TestBlitSpriteRegion(t *testing.T) {
	fb := x11.NewFramebuffer(8, 8)
	fb.Clear(0, 0, 0)

	// Create a 4x4 sprite: top-left 2x2 is red, rest is green
	sd := &x11.SpriteData{Width: 4, Height: 4, Pixels: make([]byte, 4*4*4)}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			off := (y*4 + x) * 4
			if x < 2 && y < 2 {
				sd.Pixels[off] = 0     // B
				sd.Pixels[off+1] = 0   // G
				sd.Pixels[off+2] = 255 // R
			} else {
				sd.Pixels[off] = 0     // B
				sd.Pixels[off+1] = 255 // G
				sd.Pixels[off+2] = 0   // R
			}
			sd.Pixels[off+3] = 255 // A
		}
	}
	sprite := &Sprite{data: sd}

	// Draw only the bottom-right 2x2 (green) region at (1,1)
	fb.BlitSpriteRegion(sprite.data, 1, 1, 2, 2, 2, 2)
	assertFBPixel(t, fb, 1, 1, 0, 255, 0) // green
	assertFBPixel(t, fb, 2, 2, 0, 255, 0) // green
	assertFBPixel(t, fb, 0, 0, 0, 0, 0)   // untouched black

	// Draw the top-left 2x2 (red) region at (5,5)
	fb.BlitSpriteRegion(sprite.data, 5, 5, 0, 0, 2, 2)
	assertFBPixel(t, fb, 5, 5, 255, 0, 0) // red
	assertFBPixel(t, fb, 6, 6, 255, 0, 0) // red
}

func TestAlphaBlending(t *testing.T) {
	fb := x11.NewFramebuffer(4, 4)
	// Fill with white background
	for i := 0; i < len(fb.Pixels); i += 4 {
		fb.Pixels[i] = 255   // B
		fb.Pixels[i+1] = 255 // G
		fb.Pixels[i+2] = 255 // R
		fb.Pixels[i+3] = 0
	}

	// Create a 1x1 sprite: pure red at 50% alpha
	sd := &x11.SpriteData{
		Width: 1, Height: 1,
		Pixels: []byte{0, 0, 255, 128}, // BGRA: blue=0, green=0, red=255, alpha=128
	}

	fb.BlitSprite(sd, 0, 0)

	// Expected: blend red onto white
	// R: (255*128 + 255*127 + 1 + ((255*128 + 255*127)>>8)) >> 8 = 255
	// G: (0*128 + 255*127 + 1 + ((0*128 + 255*127)>>8)) >> 8 ≈ 127
	// B: same as G ≈ 127
	r, g, b := fb.GetPixel(0, 0)
	if r != 255 {
		t.Errorf("R: expected 255, got %d", r)
	}
	// G and B should be approximately 127 (allow ±1 for rounding)
	if g < 126 || g > 128 {
		t.Errorf("G: expected ~127, got %d", g)
	}
	if b < 126 || b > 128 {
		t.Errorf("B: expected ~127, got %d", b)
	}
}

// --- Helpers ---

func makeOpaqueRedSprite(w, h int) *Sprite {
	pixels := make([]byte, w*h*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 0     // B
		pixels[i+1] = 0   // G
		pixels[i+2] = 255 // R
		pixels[i+3] = 255 // A
	}
	return &Sprite{data: &x11.SpriteData{Width: w, Height: h, Pixels: pixels}}
}

func pixelAt(s *Sprite, x, y int) [4]byte {
	off := (y*s.Width() + x) * 4
	return [4]byte{s.data.Pixels[off], s.data.Pixels[off+1], s.data.Pixels[off+2], s.data.Pixels[off+3]}
}

func assertPixel(t *testing.T, s *Sprite, x, y int, b, g, r, a uint8) {
	t.Helper()
	p := pixelAt(s, x, y)
	if p[0] != b || p[1] != g || p[2] != r || p[3] != a {
		t.Errorf("pixel (%d,%d): expected BGRA [%d,%d,%d,%d], got [%d,%d,%d,%d]",
			x, y, b, g, r, a, p[0], p[1], p[2], p[3])
	}
}

func assertFBPixel(t *testing.T, fb *x11.Framebuffer, x, y int, er, eg, eb uint8) {
	t.Helper()
	r, g, b := fb.GetPixel(x, y)
	if r != er || g != eg || b != eb {
		t.Errorf("fb pixel (%d,%d): expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
			x, y, er, eg, eb, r, g, b)
	}
}
