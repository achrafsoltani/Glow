// Paint - Simple drawing program built with Glow
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/AchrafSoltani/glow"
)

const (
	screenWidth  = 900
	screenHeight = 700
	toolbarHeight = 60
)

type Tool int

const (
	ToolBrush Tool = iota
	ToolLine
	ToolRect
	ToolCircle
	ToolEraser
)

type PaintApp struct {
	canvas      *glow.Canvas
	brushSize   int
	currentTool Tool
	color       glow.Color

	// Drawing state
	drawing    bool
	lastX      int
	lastY      int
	startX     int
	startY     int

	// Color palette
	colors []glow.Color
	colorIndex int
}

func main() {
	win, err := glow.NewWindow("Glow Paint", screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer win.Close()

	fmt.Println("=== GLOW PAINT ===")
	fmt.Println("Mouse: Draw")
	fmt.Println("1-5: Select tool (Brush, Line, Rect, Circle, Eraser)")
	fmt.Println("Q/E: Decrease/Increase brush size")
	fmt.Println("A/D: Previous/Next color")
	fmt.Println("C: Clear canvas")
	fmt.Println("ESC: Quit")

	app := &PaintApp{
		canvas:    win.Canvas(),
		brushSize: 5,
		currentTool: ToolBrush,
		colors: []glow.Color{
			glow.Black,
			glow.White,
			glow.Red,
			glow.Green,
			glow.Blue,
			glow.Yellow,
			glow.Cyan,
			glow.Magenta,
			glow.Orange,
			glow.Purple,
			glow.RGB(139, 69, 19),   // Brown
			glow.RGB(255, 192, 203), // Pink
		},
		colorIndex: 2, // Start with red
	}
	app.color = app.colors[app.colorIndex]

	// Clear canvas to white
	clearCanvas(app)

	running := true
	for running {
		// Handle events
		for {
			event := win.PollEvent()
			if event == nil {
				break
			}

			switch event.Type {
			case glow.EventQuit:
				running = false

			case glow.EventKeyDown:
				switch event.Key {
				case glow.KeyEscape:
					running = false
				case glow.KeyC:
					clearCanvas(app)
				case glow.KeyQ:
					app.brushSize = max(1, app.brushSize-2)
					fmt.Printf("Brush size: %d\n", app.brushSize)
				case glow.KeyE:
					app.brushSize = min(50, app.brushSize+2)
					fmt.Printf("Brush size: %d\n", app.brushSize)
				case glow.KeyA:
					app.colorIndex = (app.colorIndex - 1 + len(app.colors)) % len(app.colors)
					app.color = app.colors[app.colorIndex]
				case glow.KeyD:
					app.colorIndex = (app.colorIndex + 1) % len(app.colors)
					app.color = app.colors[app.colorIndex]
				case glow.Key1:
					app.currentTool = ToolBrush
					fmt.Println("Tool: Brush")
				case glow.Key2:
					app.currentTool = ToolLine
					fmt.Println("Tool: Line")
				case glow.Key3:
					app.currentTool = ToolRect
					fmt.Println("Tool: Rectangle")
				case glow.Key4:
					app.currentTool = ToolCircle
					fmt.Println("Tool: Circle")
				case glow.Key5:
					app.currentTool = ToolEraser
					fmt.Println("Tool: Eraser")
				}

			case glow.EventMouseButtonDown:
				if event.Button == glow.MouseLeft && event.Y > toolbarHeight {
					app.drawing = true
					app.startX = event.X
					app.startY = event.Y
					app.lastX = event.X
					app.lastY = event.Y

					if app.currentTool == ToolBrush || app.currentTool == ToolEraser {
						drawBrush(app, event.X, event.Y)
					}
				}
				// Check toolbar clicks
				if event.Y < toolbarHeight {
					handleToolbarClick(app, event.X, event.Y)
				}

			case glow.EventMouseButtonUp:
				if event.Button == glow.MouseLeft && app.drawing {
					app.drawing = false

					// Finalize shape tools
					switch app.currentTool {
					case ToolLine:
						drawLine(app.canvas, app.startX, app.startY, event.X, event.Y, app.color, app.brushSize)
					case ToolRect:
						drawRectTool(app.canvas, app.startX, app.startY, event.X, event.Y, app.color, app.brushSize)
					case ToolCircle:
						drawCircleTool(app.canvas, app.startX, app.startY, event.X, event.Y, app.color, app.brushSize)
					}
				}

			case glow.EventMouseMotion:
				if app.drawing && event.Y > toolbarHeight {
					if app.currentTool == ToolBrush || app.currentTool == ToolEraser {
						// Draw continuous line
						drawLineBrush(app, app.lastX, app.lastY, event.X, event.Y)
						app.lastX = event.X
						app.lastY = event.Y
					}
				}
			}
		}

		// Draw toolbar
		drawToolbar(app)

		win.Present()
		time.Sleep(8 * time.Millisecond)
	}

	fmt.Println("\nPaint closed!")
}

func clearCanvas(app *PaintApp) {
	// Fill with white, leaving toolbar area
	for y := toolbarHeight; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			app.canvas.SetPixel(x, y, glow.White)
		}
	}
}

func drawBrush(app *PaintApp, x, y int) {
	color := app.color
	if app.currentTool == ToolEraser {
		color = glow.White
	}
	app.canvas.FillCircle(x, y, app.brushSize, color)
}

func drawLineBrush(app *PaintApp, x0, y0, x1, y1 int) {
	color := app.color
	if app.currentTool == ToolEraser {
		color = glow.White
	}

	// Bresenham's line with brush at each point
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
		app.canvas.FillCircle(x0, y0, app.brushSize, color)
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

func drawLine(canvas *glow.Canvas, x0, y0, x1, y1 int, color glow.Color, thickness int) {
	// Draw thick line
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
		canvas.FillCircle(x0, y0, thickness, color)
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

func drawRectTool(canvas *glow.Canvas, x0, y0, x1, y1 int, color glow.Color, thickness int) {
	minX, maxX := min(x0, x1), max(x0, x1)
	minY, maxY := min(y0, y1), max(y0, y1)

	// Draw four sides
	for t := 0; t < thickness; t++ {
		// Top
		for x := minX; x <= maxX; x++ {
			canvas.SetPixel(x, minY+t, color)
		}
		// Bottom
		for x := minX; x <= maxX; x++ {
			canvas.SetPixel(x, maxY-t, color)
		}
		// Left
		for y := minY; y <= maxY; y++ {
			canvas.SetPixel(minX+t, y, color)
		}
		// Right
		for y := minY; y <= maxY; y++ {
			canvas.SetPixel(maxX-t, y, color)
		}
	}
}

func drawCircleTool(canvas *glow.Canvas, x0, y0, x1, y1 int, color glow.Color, thickness int) {
	cx := (x0 + x1) / 2
	cy := (y0 + y1) / 2
	rx := abs(x1-x0) / 2
	ry := abs(y1-y0) / 2
	radius := (rx + ry) / 2

	for t := 0; t < thickness; t++ {
		canvas.DrawCircle(cx, cy, radius-t, color)
	}
}

func drawToolbar(app *PaintApp) {
	// Background
	app.canvas.DrawRect(0, 0, screenWidth, toolbarHeight, glow.RGB(60, 60, 70))

	// Tool buttons
	tools := []string{"Brush", "Line", "Rect", "Circle", "Eraser"}
	for i, _ := range tools {
		x := 10 + i*70
		color := glow.RGB(80, 80, 90)
		if Tool(i) == app.currentTool {
			color = glow.RGB(100, 150, 200)
		}
		app.canvas.DrawRect(x, 10, 60, 40, color)
		app.canvas.DrawRectOutline(x, 10, 60, 40, glow.White)

		// Tool indicator
		switch Tool(i) {
		case ToolBrush:
			app.canvas.FillCircle(x+30, 30, 10, glow.White)
		case ToolLine:
			app.canvas.DrawLine(x+10, 40, x+50, 20, glow.White)
		case ToolRect:
			app.canvas.DrawRectOutline(x+15, 18, 30, 24, glow.White)
		case ToolCircle:
			app.canvas.DrawCircle(x+30, 30, 12, glow.White)
		case ToolEraser:
			app.canvas.DrawRect(x+20, 20, 20, 20, glow.White)
		}
	}

	// Color palette
	for i, c := range app.colors {
		x := 380 + (i%6)*35
		y := 8 + (i/6)*25
		app.canvas.DrawRect(x, y, 30, 20, c)
		if i == app.colorIndex {
			app.canvas.DrawRectOutline(x-2, y-2, 34, 24, glow.White)
		}
	}

	// Current color preview
	app.canvas.DrawRect(600, 10, 40, 40, app.color)
	app.canvas.DrawRectOutline(600, 10, 40, 40, glow.White)

	// Brush size indicator
	app.canvas.DrawRect(660, 10, 60, 40, glow.RGB(50, 50, 60))
	app.canvas.FillCircle(690, 30, min(app.brushSize, 18), glow.White)

	// Clear button
	app.canvas.DrawRect(740, 10, 60, 40, glow.RGB(180, 60, 60))
	app.canvas.DrawRectOutline(740, 10, 60, 40, glow.White)
	// "C" letter
	app.canvas.DrawRect(760, 20, 20, 3, glow.White)
	app.canvas.DrawRect(760, 37, 20, 3, glow.White)
	app.canvas.DrawRect(760, 20, 3, 20, glow.White)

	// Separator line
	app.canvas.DrawRect(0, toolbarHeight-2, screenWidth, 2, glow.RGB(40, 40, 50))
}

func handleToolbarClick(app *PaintApp, x, y int) {
	// Tool buttons
	for i := 0; i < 5; i++ {
		bx := 10 + i*70
		if x >= bx && x < bx+60 && y >= 10 && y < 50 {
			app.currentTool = Tool(i)
			fmt.Printf("Tool: %v\n", []string{"Brush", "Line", "Rect", "Circle", "Eraser"}[i])
			return
		}
	}

	// Color palette
	for i := range app.colors {
		cx := 380 + (i%6)*35
		cy := 8 + (i/6)*25
		if x >= cx && x < cx+30 && y >= cy && y < cy+20 {
			app.colorIndex = i
			app.color = app.colors[i]
			return
		}
	}

	// Clear button
	if x >= 740 && x < 800 && y >= 10 && y < 50 {
		clearCanvas(app)
	}
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
