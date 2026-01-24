# Step 14: Project - Paint Program

## Goal
Build a drawing application with multiple tools.

## Features

1. Brush tool (freehand drawing)
2. Line tool
3. Rectangle tool
4. Circle tool
5. Eraser
6. Color palette
7. Brush size control

## Application Structure

```go
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
    currentTool Tool
    brushSize   int
    color       glow.Color
    colors      []glow.Color

    // Drawing state
    drawing bool
    startX  int
    startY  int
    lastX   int
    lastY   int
}
```

## Tool Implementations

### Brush Tool
```go
func (app *PaintApp) drawBrush(x, y int) {
    color := app.color
    if app.currentTool == ToolEraser {
        color = glow.White
    }
    app.canvas.FillCircle(x, y, app.brushSize, color)
}

// Draw continuous line between points (while dragging)
func (app *PaintApp) drawLineBrush(x0, y0, x1, y1 int) {
    // Use Bresenham's algorithm with circles at each point
    dx := abs(x1 - x0)
    dy := -abs(y1 - y0)
    sx, sy := 1, 1
    if x0 > x1 { sx = -1 }
    if y0 > y1 { sy = -1 }
    err := dx + dy

    for {
        app.drawBrush(x0, y0)
        if x0 == x1 && y0 == y1 { break }
        e2 := 2 * err
        if e2 >= dy { err += dy; x0 += sx }
        if e2 <= dx { err += dx; y0 += sy }
    }
}
```

### Shape Tools (Line, Rectangle, Circle)

These draw preview while dragging, finalize on mouse up:

```go
func (app *PaintApp) onMouseDown(x, y int) {
    app.drawing = true
    app.startX, app.startY = x, y
    app.lastX, app.lastY = x, y

    if app.currentTool == ToolBrush || app.currentTool == ToolEraser {
        app.drawBrush(x, y)
    }
}

func (app *PaintApp) onMouseUp(x, y int) {
    if !app.drawing { return }
    app.drawing = false

    switch app.currentTool {
    case ToolLine:
        drawThickLine(app.canvas, app.startX, app.startY, x, y,
            app.color, app.brushSize)
    case ToolRect:
        drawRectOutline(app.canvas, app.startX, app.startY, x, y,
            app.color, app.brushSize)
    case ToolCircle:
        drawCircleOutline(app.canvas, app.startX, app.startY, x, y,
            app.color, app.brushSize)
    }
}

func (app *PaintApp) onMouseMove(x, y int) {
    if !app.drawing { return }

    if app.currentTool == ToolBrush || app.currentTool == ToolEraser {
        app.drawLineBrush(app.lastX, app.lastY, x, y)
        app.lastX, app.lastY = x, y
    }
    // Shape tools would show preview here (requires double buffering)
}
```

## Toolbar UI

```go
const toolbarHeight = 60

func drawToolbar(app *PaintApp) {
    canvas := app.canvas

    // Background
    canvas.DrawRect(0, 0, screenWidth, toolbarHeight, glow.RGB(60, 60, 70))

    // Tool buttons
    tools := []Tool{ToolBrush, ToolLine, ToolRect, ToolCircle, ToolEraser}
    for i, tool := range tools {
        x := 10 + i*70
        color := glow.RGB(80, 80, 90)
        if tool == app.currentTool {
            color = glow.RGB(100, 150, 200)
        }
        canvas.DrawRect(x, 10, 60, 40, color)
        canvas.DrawRectOutline(x, 10, 60, 40, glow.White)
        drawToolIcon(canvas, tool, x+30, 30)
    }

    // Color palette
    for i, c := range app.colors {
        x := 400 + (i%6)*30
        y := 10 + (i/6)*25
        canvas.DrawRect(x, y, 25, 20, c)
    }

    // Brush size indicator
    canvas.FillCircle(700, 30, min(app.brushSize, 20), glow.White)
}
```

## Keyboard Shortcuts

```go
case glow.EventKeyDown:
    switch event.Key {
    case glow.Key1: app.currentTool = ToolBrush
    case glow.Key2: app.currentTool = ToolLine
    case glow.Key3: app.currentTool = ToolRect
    case glow.Key4: app.currentTool = ToolCircle
    case glow.Key5: app.currentTool = ToolEraser
    case glow.KeyQ: app.brushSize = max(1, app.brushSize-2)
    case glow.KeyE: app.brushSize = min(50, app.brushSize+2)
    case glow.KeyC: clearCanvas(app)
    }
```

## Learning Outcomes

- Mouse event handling (down, up, move)
- Tool state machines
- UI layout (toolbar)
- Drawing algorithms

## Run It

```bash
go run examples/paint/main.go
```

Controls:
- Mouse: Draw
- 1-5: Select tool
- Q/E: Brush size
- C: Clear
- ESC: Quit
