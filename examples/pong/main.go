// Pong - Classic arcade game built with Glow
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/AchrafSoltani/glow"
)

const (
	screenWidth  = 800
	screenHeight = 600
	paddleWidth  = 15
	paddleHeight = 80
	paddleSpeed  = 8.0
	ballSize     = 12
	ballSpeed    = 5.0
	winScore     = 5
)

type Paddle struct {
	X, Y   float64
	Width  float64
	Height float64
	Score  int
}

type Ball struct {
	X, Y   float64
	VX, VY float64
	Size   float64
}

func main() {
	rand.Seed(time.Now().UnixNano())

	win, err := glow.NewWindow("Glow Pong", screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer win.Close()

	fmt.Println("=== PONG ===")
	fmt.Println("Player 1 (Left):  W/S keys")
	fmt.Println("Player 2 (Right): Up/Down arrows")
	fmt.Println("First to 5 wins!")
	fmt.Println("Press SPACE to start, ESC to quit")

	// Initialize game objects
	paddle1 := &Paddle{
		X:      30,
		Y:      float64(screenHeight)/2 - paddleHeight/2,
		Width:  paddleWidth,
		Height: paddleHeight,
	}

	paddle2 := &Paddle{
		X:      float64(screenWidth) - 30 - paddleWidth,
		Y:      float64(screenHeight)/2 - paddleHeight/2,
		Width:  paddleWidth,
		Height: paddleHeight,
	}

	ball := &Ball{
		X:    float64(screenWidth) / 2,
		Y:    float64(screenHeight) / 2,
		Size: ballSize,
	}

	// Key states
	keys := make(map[glow.Key]bool)

	// Game state
	gameStarted := false
	gameOver := false
	winner := 0

	// Main game loop
	running := true
	lastTime := time.Now()

	for running {
		// Delta time
		now := time.Now()
		dt := now.Sub(lastTime).Seconds() * 60 // Normalize to 60fps
		lastTime = now

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
				keys[event.Key] = true

				if event.Key == glow.KeyEscape {
					running = false
				}

				if event.Key == glow.KeySpace {
					if !gameStarted || gameOver {
						// Start/restart game
						resetBall(ball)
						paddle1.Score = 0
						paddle2.Score = 0
						gameStarted = true
						gameOver = false
					}
				}

			case glow.EventKeyUp:
				keys[event.Key] = false
			}
		}

		if gameStarted && !gameOver {
			// Move paddle 1 (W/S)
			if keys[glow.KeyW] {
				paddle1.Y -= paddleSpeed * dt
			}
			if keys[glow.KeyS] {
				paddle1.Y += paddleSpeed * dt
			}

			// Move paddle 2 (Up/Down)
			if keys[glow.KeyUp] {
				paddle2.Y -= paddleSpeed * dt
			}
			if keys[glow.KeyDown] {
				paddle2.Y += paddleSpeed * dt
			}

			// Clamp paddles to screen
			paddle1.Y = clamp(paddle1.Y, 0, float64(screenHeight)-paddle1.Height)
			paddle2.Y = clamp(paddle2.Y, 0, float64(screenHeight)-paddle2.Height)

			// Move ball
			ball.X += ball.VX * dt
			ball.Y += ball.VY * dt

			// Ball collision with top/bottom walls
			if ball.Y <= 0 || ball.Y >= float64(screenHeight)-ball.Size {
				ball.VY = -ball.VY
				ball.Y = clamp(ball.Y, 0, float64(screenHeight)-ball.Size)
			}

			// Ball collision with paddles
			if ballHitsPaddle(ball, paddle1) {
				ball.VX = math.Abs(ball.VX) // Go right
				ball.X = paddle1.X + paddle1.Width + 1
				// Add spin based on where it hits the paddle
				relativeY := (ball.Y + ball.Size/2) - (paddle1.Y + paddle1.Height/2)
				ball.VY += relativeY * 0.1
				// Speed up slightly
				ball.VX *= 1.05
			}

			if ballHitsPaddle(ball, paddle2) {
				ball.VX = -math.Abs(ball.VX) // Go left
				ball.X = paddle2.X - ball.Size - 1
				// Add spin
				relativeY := (ball.Y + ball.Size/2) - (paddle2.Y + paddle2.Height/2)
				ball.VY += relativeY * 0.1
				// Speed up
				ball.VX *= 1.05
			}

			// Scoring
			if ball.X < 0 {
				paddle2.Score++
				if paddle2.Score >= winScore {
					gameOver = true
					winner = 2
				} else {
					resetBall(ball)
				}
			}

			if ball.X > float64(screenWidth) {
				paddle1.Score++
				if paddle1.Score >= winScore {
					gameOver = true
					winner = 1
				} else {
					resetBall(ball)
				}
			}
		}

		// Draw
		canvas := win.Canvas()
		canvas.Clear(glow.RGB(20, 20, 30))

		// Draw center line
		for y := 0; y < screenHeight; y += 20 {
			canvas.DrawRect(screenWidth/2-2, y, 4, 10, glow.RGB(50, 50, 60))
		}

		// Draw paddles
		canvas.DrawRect(int(paddle1.X), int(paddle1.Y), int(paddle1.Width), int(paddle1.Height), glow.RGB(100, 200, 100))
		canvas.DrawRect(int(paddle2.X), int(paddle2.Y), int(paddle2.Width), int(paddle2.Height), glow.RGB(100, 100, 200))

		// Draw ball
		if gameStarted {
			canvas.DrawRect(int(ball.X), int(ball.Y), int(ball.Size), int(ball.Size), glow.White)
		}

		// Draw scores
		drawNumber(canvas, paddle1.Score, screenWidth/4, 50, glow.RGB(100, 200, 100))
		drawNumber(canvas, paddle2.Score, 3*screenWidth/4-20, 50, glow.RGB(100, 100, 200))

		// Draw messages
		if !gameStarted {
			drawCenteredText(canvas, "PRESS SPACE TO START", screenHeight/2)
		} else if gameOver {
			if winner == 1 {
				drawCenteredText(canvas, "PLAYER 1 WINS!", screenHeight/2-20)
			} else {
				drawCenteredText(canvas, "PLAYER 2 WINS!", screenHeight/2-20)
			}
			drawCenteredText(canvas, "PRESS SPACE TO RESTART", screenHeight/2+20)
		}

		win.Present()
		time.Sleep(16 * time.Millisecond)
	}

	fmt.Println("\nGame Over!")
	fmt.Printf("Final Score: Player 1: %d - Player 2: %d\n", paddle1.Score, paddle2.Score)
}

func resetBall(ball *Ball) {
	ball.X = float64(screenWidth) / 2
	ball.Y = float64(screenHeight) / 2

	// Random direction
	angle := (rand.Float64() - 0.5) * math.Pi / 2 // -45 to +45 degrees
	if rand.Intn(2) == 0 {
		ball.VX = ballSpeed * math.Cos(angle)
	} else {
		ball.VX = -ballSpeed * math.Cos(angle)
	}
	ball.VY = ballSpeed * math.Sin(angle)
}

func ballHitsPaddle(ball *Ball, paddle *Paddle) bool {
	return ball.X < paddle.X+paddle.Width &&
		ball.X+ball.Size > paddle.X &&
		ball.Y < paddle.Y+paddle.Height &&
		ball.Y+ball.Size > paddle.Y
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Simple 7-segment style number drawing
func drawNumber(canvas *glow.Canvas, n int, x, y int, color glow.Color) {
	// Segments for digits 0-9
	segments := [][]bool{
		{true, true, true, false, true, true, true},    // 0
		{false, false, true, false, false, true, false}, // 1
		{true, false, true, true, true, false, true},   // 2
		{true, false, true, true, false, true, true},   // 3
		{false, true, true, true, false, true, false},  // 4
		{true, true, false, true, false, true, true},   // 5
		{true, true, false, true, true, true, true},    // 6
		{true, false, true, false, false, true, false}, // 7
		{true, true, true, true, true, true, true},     // 8
		{true, true, true, true, false, true, true},    // 9
	}

	if n < 0 || n > 9 {
		return
	}

	seg := segments[n]
	w, h := 30, 50
	t := 5 // thickness

	// Top
	if seg[0] {
		canvas.DrawRect(x+t, y, w-2*t, t, color)
	}
	// Top-left
	if seg[1] {
		canvas.DrawRect(x, y+t, t, h/2-t, color)
	}
	// Top-right
	if seg[2] {
		canvas.DrawRect(x+w-t, y+t, t, h/2-t, color)
	}
	// Middle
	if seg[3] {
		canvas.DrawRect(x+t, y+h/2-t/2, w-2*t, t, color)
	}
	// Bottom-left
	if seg[4] {
		canvas.DrawRect(x, y+h/2+t/2, t, h/2-t, color)
	}
	// Bottom-right
	if seg[5] {
		canvas.DrawRect(x+w-t, y+h/2+t/2, t, h/2-t, color)
	}
	// Bottom
	if seg[6] {
		canvas.DrawRect(x+t, y+h-t, w-2*t, t, color)
	}
}

func drawCenteredText(canvas *glow.Canvas, text string, y int) {
	// Simple pixel text - just draw a rectangle as placeholder
	w := len(text) * 8
	x := (screenWidth - w) / 2
	canvas.DrawRect(x, y, w, 16, glow.RGB(80, 80, 80))
	// Draw dots for each character
	for i, c := range text {
		if c != ' ' {
			canvas.DrawRect(x+i*8+2, y+4, 4, 8, glow.White)
		}
	}
}
