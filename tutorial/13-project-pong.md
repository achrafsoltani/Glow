# Step 13: Project - Pong

## Goal
Build a complete Pong game to practice everything learned.

## Game Elements

1. Two paddles (player controlled)
2. Ball with physics
3. Score display
4. Win condition

## Game Structure

```go
// Game constants
const (
    screenWidth  = 800
    screenHeight = 600
    paddleWidth  = 15
    paddleHeight = 80
    paddleSpeed  = 8.0
    ballSize     = 12
    ballSpeed    = 5.0
)

// Paddle structure
type Paddle struct {
    X, Y   float64
    Width  float64
    Height float64
    Score  int
}

// Ball structure
type Ball struct {
    X, Y   float64
    VX, VY float64
    Size   float64
}
```

## Game Loop

```go
func main() {
    win, _ := glow.NewWindow("Pong", screenWidth, screenHeight)
    defer win.Close()

    paddle1 := &Paddle{X: 30, Y: screenHeight/2}
    paddle2 := &Paddle{X: screenWidth-45, Y: screenHeight/2}
    ball := &Ball{X: screenWidth/2, Y: screenHeight/2}

    keys := make(map[glow.Key]bool)

    for running {
        // 1. Handle input
        for event := win.PollEvent(); event != nil; event = win.PollEvent() {
            switch event.Type {
            case glow.EventKeyDown:
                keys[event.Key] = true
            case glow.EventKeyUp:
                keys[event.Key] = false
            case glow.EventQuit:
                running = false
            }
        }

        // 2. Update game state
        updatePaddles(paddle1, paddle2, keys)
        updateBall(ball, paddle1, paddle2)
        checkScoring(ball, paddle1, paddle2)

        // 3. Render
        canvas := win.Canvas()
        canvas.Clear(glow.RGB(20, 20, 30))
        drawPaddle(canvas, paddle1)
        drawPaddle(canvas, paddle2)
        drawBall(canvas, ball)
        drawScore(canvas, paddle1.Score, paddle2.Score)
        win.Present()

        time.Sleep(16 * time.Millisecond)
    }
}
```

## Paddle Movement

```go
func updatePaddles(p1, p2 *Paddle, keys map[glow.Key]bool) {
    // Player 1: W/S
    if keys[glow.KeyW] { p1.Y -= paddleSpeed }
    if keys[glow.KeyS] { p1.Y += paddleSpeed }

    // Player 2: Up/Down
    if keys[glow.KeyUp]   { p2.Y -= paddleSpeed }
    if keys[glow.KeyDown] { p2.Y += paddleSpeed }

    // Clamp to screen
    p1.Y = clamp(p1.Y, 0, screenHeight-paddleHeight)
    p2.Y = clamp(p2.Y, 0, screenHeight-paddleHeight)
}
```

## Ball Physics

```go
func updateBall(ball *Ball, p1, p2 *Paddle) {
    ball.X += ball.VX
    ball.Y += ball.VY

    // Bounce off top/bottom
    if ball.Y <= 0 || ball.Y >= screenHeight-ballSize {
        ball.VY = -ball.VY
    }

    // Bounce off paddles
    if ballHitsPaddle(ball, p1) {
        ball.VX = abs(ball.VX)  // Go right
        addSpin(ball, p1)
    }
    if ballHitsPaddle(ball, p2) {
        ball.VX = -abs(ball.VX)  // Go left
        addSpin(ball, p2)
    }
}

func ballHitsPaddle(ball *Ball, paddle *Paddle) bool {
    return ball.X < paddle.X+paddle.Width &&
           ball.X+ball.Size > paddle.X &&
           ball.Y < paddle.Y+paddle.Height &&
           ball.Y+ball.Size > paddle.Y
}

func addSpin(ball *Ball, paddle *Paddle) {
    // Hit position affects angle
    relY := (ball.Y + ball.Size/2) - (paddle.Y + paddle.Height/2)
    ball.VY += relY * 0.1
    ball.VX *= 1.05  // Speed up
}
```

## Scoring

```go
func checkScoring(ball *Ball, p1, p2 *Paddle) {
    if ball.X < 0 {
        p2.Score++
        resetBall(ball)
    }
    if ball.X > screenWidth {
        p1.Score++
        resetBall(ball)
    }
}

func resetBall(ball *Ball) {
    ball.X = screenWidth / 2
    ball.Y = screenHeight / 2
    angle := (rand.Float64()-0.5) * math.Pi/2
    ball.VX = ballSpeed * math.Cos(angle)
    ball.VY = ballSpeed * math.Sin(angle)
    if rand.Intn(2) == 0 {
        ball.VX = -ball.VX
    }
}
```

## Drawing Score (7-Segment Display)

```go
func drawNumber(canvas *glow.Canvas, n int, x, y int, color glow.Color) {
    // 7-segment patterns for 0-9
    segments := [][]bool{
        {true, true, true, false, true, true, true},     // 0
        {false, false, true, false, false, true, false}, // 1
        // ... etc
    }

    seg := segments[n]
    w, h, t := 30, 50, 5

    if seg[0] { canvas.DrawRect(x+t, y, w-2*t, t, color) }           // Top
    if seg[1] { canvas.DrawRect(x, y+t, t, h/2-t, color) }           // Top-left
    if seg[2] { canvas.DrawRect(x+w-t, y+t, t, h/2-t, color) }       // Top-right
    if seg[3] { canvas.DrawRect(x+t, y+h/2-t/2, w-2*t, t, color) }   // Middle
    if seg[4] { canvas.DrawRect(x, y+h/2+t/2, t, h/2-t, color) }     // Bottom-left
    if seg[5] { canvas.DrawRect(x+w-t, y+h/2+t/2, t, h/2-t, color) } // Bottom-right
    if seg[6] { canvas.DrawRect(x+t, y+h-t, w-2*t, t, color) }       // Bottom
}
```

## Learning Outcomes

- Game loop structure
- Input handling (key states)
- Simple physics (velocity, collision)
- Game state management

## Run It

```bash
go run examples/pong/main.go
```

Controls:
- Player 1: W/S
- Player 2: Up/Down arrows
- ESC: Quit
