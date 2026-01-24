// Package glow is a pure Go 2D graphics library inspired by SDL.
// It provides window management, input handling, and software rendering.
package glow

import (
	"github.com/AchrafSoltani/glow/internal/x11"
)

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// Predefined colors
var (
	Black   = Color{0, 0, 0}
	White   = Color{255, 255, 255}
	Red     = Color{255, 0, 0}
	Green   = Color{0, 255, 0}
	Blue    = Color{0, 0, 255}
	Yellow  = Color{255, 255, 0}
	Cyan    = Color{0, 255, 255}
	Magenta = Color{255, 0, 255}
	Orange  = Color{255, 165, 0}
	Purple  = Color{128, 0, 128}
	Gray    = Color{128, 128, 128}
)

// RGB creates a color from red, green, blue components
func RGB(r, g, b uint8) Color {
	return Color{r, g, b}
}

// Hex creates a color from a hex value (0xRRGGBB)
func Hex(hex uint32) Color {
	return Color{
		R: uint8((hex >> 16) & 0xFF),
		G: uint8((hex >> 8) & 0xFF),
		B: uint8(hex & 0xFF),
	}
}

// Window represents a graphics window
type Window struct {
	conn     *x11.Connection
	windowID uint32
	gcID     uint32
	canvas   *Canvas
	width    int
	height   int
	closed   bool

	// Event handling
	eventChan chan Event
	quitChan  chan struct{}
}

// Canvas is the drawing surface
type Canvas struct {
	fb *x11.Framebuffer
}

// NewWindow creates a new window with the given title and dimensions
func NewWindow(title string, width, height int) (*Window, error) {
	conn, err := x11.Connect()
	if err != nil {
		return nil, err
	}

	windowID, err := conn.CreateWindow(100, 100, uint16(width), uint16(height))
	if err != nil {
		conn.Close()
		return nil, err
	}

	gcID, err := conn.CreateGC(windowID)
	if err != nil {
		conn.DestroyWindow(windowID)
		conn.Close()
		return nil, err
	}

	// Set window title
	if err := conn.SetWindowTitle(windowID, title); err != nil {
		conn.FreeGC(gcID)
		conn.DestroyWindow(windowID)
		conn.Close()
		return nil, err
	}

	// Enable close button
	if err := conn.EnableCloseButton(windowID); err != nil {
		conn.FreeGC(gcID)
		conn.DestroyWindow(windowID)
		conn.Close()
		return nil, err
	}

	if err := conn.MapWindow(windowID); err != nil {
		conn.FreeGC(gcID)
		conn.DestroyWindow(windowID)
		conn.Close()
		return nil, err
	}

	fb := x11.NewFramebuffer(width, height)

	w := &Window{
		conn:      conn,
		windowID:  windowID,
		gcID:      gcID,
		canvas:    &Canvas{fb: fb},
		width:     width,
		height:    height,
		eventChan: make(chan Event, 256),
		quitChan:  make(chan struct{}),
	}

	// Start event polling goroutine
	go w.pollEvents()

	return w, nil
}

// Close closes the window and releases resources
func (w *Window) Close() {
	if w.closed {
		return
	}
	w.closed = true

	// Signal event goroutine to stop
	close(w.quitChan)

	w.conn.FreeGC(w.gcID)
	w.conn.DestroyWindow(w.windowID)
	w.conn.Close()
}

// Width returns the window width
func (w *Window) Width() int { return w.width }

// Height returns the window height
func (w *Window) Height() int { return w.height }

// Canvas returns the drawing canvas
func (w *Window) Canvas() *Canvas { return w.canvas }

// Present copies the canvas to the screen
func (w *Window) Present() error {
	return w.conn.PutImage(w.windowID, w.gcID,
		uint16(w.width), uint16(w.height), 0, 0,
		w.conn.RootDepth, w.canvas.fb.Pixels)
}

// --- Canvas Drawing Methods ---

// Clear fills the canvas with a solid color
func (c *Canvas) Clear(color Color) {
	c.fb.Clear(color.R, color.G, color.B)
}

// SetPixel sets a single pixel
func (c *Canvas) SetPixel(x, y int, color Color) {
	c.fb.SetPixel(x, y, color.R, color.G, color.B)
}

// GetPixel returns the color at (x, y)
func (c *Canvas) GetPixel(x, y int) Color {
	r, g, b := c.fb.GetPixel(x, y)
	return Color{r, g, b}
}

// DrawRect draws a filled rectangle
func (c *Canvas) DrawRect(x, y, width, height int, color Color) {
	c.fb.DrawRect(x, y, width, height, color.R, color.G, color.B)
}

// DrawRectOutline draws a rectangle outline
func (c *Canvas) DrawRectOutline(x, y, width, height int, color Color) {
	c.fb.DrawRectOutline(x, y, width, height, color.R, color.G, color.B)
}

// DrawLine draws a line between two points
func (c *Canvas) DrawLine(x0, y0, x1, y1 int, color Color) {
	c.fb.DrawLine(x0, y0, x1, y1, color.R, color.G, color.B)
}

// DrawCircle draws a circle outline
func (c *Canvas) DrawCircle(x, y, radius int, color Color) {
	c.fb.DrawCircle(x, y, radius, color.R, color.G, color.B)
}

// FillCircle draws a filled circle
func (c *Canvas) FillCircle(x, y, radius int, color Color) {
	c.fb.FillCircle(x, y, radius, color.R, color.G, color.B)
}

// DrawTriangle draws a triangle outline
func (c *Canvas) DrawTriangle(x0, y0, x1, y1, x2, y2 int, color Color) {
	c.fb.DrawTriangle(x0, y0, x1, y1, x2, y2, color.R, color.G, color.B)
}

// Width returns the canvas width
func (c *Canvas) Width() int { return c.fb.Width }

// Height returns the canvas height
func (c *Canvas) Height() int { return c.fb.Height }
