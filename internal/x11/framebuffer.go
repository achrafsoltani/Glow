package x11

// Framebuffer is a software pixel buffer for rendering
// Pixels are stored in BGRA format (Blue, Green, Red, Alpha)
// This matches X11's 24-bit depth format on little-endian systems
type Framebuffer struct {
	Width  int
	Height int
	Pixels []byte // BGRA format, 4 bytes per pixel
}

// NewFramebuffer creates a new framebuffer
func NewFramebuffer(width, height int) *Framebuffer {
	return &Framebuffer{
		Width:  width,
		Height: height,
		Pixels: make([]byte, width*height*4),
	}
}

// Clear fills the entire framebuffer with a color
func (fb *Framebuffer) Clear(r, g, b uint8) {
	for i := 0; i < len(fb.Pixels); i += 4 {
		fb.Pixels[i] = b   // Blue
		fb.Pixels[i+1] = g // Green
		fb.Pixels[i+2] = r // Red
		fb.Pixels[i+3] = 0 // Alpha (unused)
	}
}

// SetPixel sets a single pixel
func (fb *Framebuffer) SetPixel(x, y int, r, g, b uint8) {
	if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return // Clipping
	}
	offset := (y*fb.Width + x) * 4
	fb.Pixels[offset] = b
	fb.Pixels[offset+1] = g
	fb.Pixels[offset+2] = r
	fb.Pixels[offset+3] = 0
}

// GetPixel returns the color at (x, y)
func (fb *Framebuffer) GetPixel(x, y int) (r, g, b uint8) {
	if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return 0, 0, 0
	}
	offset := (y*fb.Width + x) * 4
	return fb.Pixels[offset+2], fb.Pixels[offset+1], fb.Pixels[offset]
}

// DrawRect draws a filled rectangle
func (fb *Framebuffer) DrawRect(x, y, width, height int, r, g, b uint8) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			fb.SetPixel(x+dx, y+dy, r, g, b)
		}
	}
}

// DrawRectOutline draws a rectangle outline
func (fb *Framebuffer) DrawRectOutline(x, y, width, height int, r, g, b uint8) {
	// Top and bottom
	for dx := 0; dx < width; dx++ {
		fb.SetPixel(x+dx, y, r, g, b)
		fb.SetPixel(x+dx, y+height-1, r, g, b)
	}
	// Left and right
	for dy := 0; dy < height; dy++ {
		fb.SetPixel(x, y+dy, r, g, b)
		fb.SetPixel(x+width-1, y+dy, r, g, b)
	}
}

// DrawLine draws a line using Bresenham's algorithm
func (fb *Framebuffer) DrawLine(x0, y0, x1, y1 int, r, g, b uint8) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
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

// DrawCircle draws a circle outline using midpoint algorithm
func (fb *Framebuffer) DrawCircle(cx, cy, radius int, r, g, b uint8) {
	x := radius
	y := 0
	err := 0

	for x >= y {
		// Draw 8 octants
		fb.SetPixel(cx+x, cy+y, r, g, b)
		fb.SetPixel(cx+y, cy+x, r, g, b)
		fb.SetPixel(cx-y, cy+x, r, g, b)
		fb.SetPixel(cx-x, cy+y, r, g, b)
		fb.SetPixel(cx-x, cy-y, r, g, b)
		fb.SetPixel(cx-y, cy-x, r, g, b)
		fb.SetPixel(cx+y, cy-x, r, g, b)
		fb.SetPixel(cx+x, cy-y, r, g, b)

		y++
		err += 1 + 2*y
		if 2*(err-x)+1 > 0 {
			x--
			err += 1 - 2*x
		}
	}
}

// FillCircle draws a filled circle
func (fb *Framebuffer) FillCircle(cx, cy, radius int, r, g, b uint8) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				fb.SetPixel(cx+x, cy+y, r, g, b)
			}
		}
	}
}

// DrawTriangle draws a triangle outline
func (fb *Framebuffer) DrawTriangle(x0, y0, x1, y1, x2, y2 int, r, g, b uint8) {
	fb.DrawLine(x0, y0, x1, y1, r, g, b)
	fb.DrawLine(x1, y1, x2, y2, r, g, b)
	fb.DrawLine(x2, y2, x0, y0, r, g, b)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
