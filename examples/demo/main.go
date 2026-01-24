package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/AchrafSoltani/glow"
)

func main() {
	// Create a window
	win, err := glow.NewWindow("Glow Demo", 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	defer win.Close()

	fmt.Println("=== Glow Demo ===")
	fmt.Println("WASD or Arrow keys to move")
	fmt.Println("Click to place circles")
	fmt.Println("ESC or Q to quit")

	// Player state
	playerX := float64(win.Width()) / 2
	playerY := float64(win.Height()) / 2
	playerSpeed := 5.0

	// Track key states
	keys := make(map[glow.Key]bool)

	// Circles placed by clicking
	type Circle struct {
		X, Y   int
		Color  glow.Color
		Radius int
	}
	var circles []Circle

	// Animation
	frame := 0
	running := true

	for running {
		// Handle events
		for {
			event := win.PollEvent()
			if event == nil {
				break
			}

			switch event.Type {
			case glow.EventKeyDown:
				keys[event.Key] = true
				if event.Key == glow.KeyEscape || event.Key == glow.KeyQ {
					running = false
				}

			case glow.EventKeyUp:
				keys[event.Key] = false

			case glow.EventMouseButtonDown:
				if event.Button == glow.MouseLeft {
					// Add a circle where clicked
					circles = append(circles, Circle{
						X:      event.X,
						Y:      event.Y,
						Color:  glow.RGB(uint8(event.X%256), uint8(event.Y%256), 150),
						Radius: 20,
					})
				}
			}
		}

		// Move player based on keys held
		if keys[glow.KeyW] || keys[glow.KeyUp] {
			playerY -= playerSpeed
		}
		if keys[glow.KeyS] || keys[glow.KeyDown] {
			playerY += playerSpeed
		}
		if keys[glow.KeyA] || keys[glow.KeyLeft] {
			playerX -= playerSpeed
		}
		if keys[glow.KeyD] || keys[glow.KeyRight] {
			playerX += playerSpeed
		}

		// Keep player in bounds
		playerX = clamp(playerX, 20, float64(win.Width())-20)
		playerY = clamp(playerY, 20, float64(win.Height())-20)

		// Draw
		canvas := win.Canvas()

		// Clear with animated background
		t := float64(frame) * 0.01
		bgR := uint8(20 + 10*math.Sin(t))
		bgG := uint8(20 + 10*math.Sin(t+1))
		bgB := uint8(40 + 10*math.Sin(t+2))
		canvas.Clear(glow.RGB(bgR, bgG, bgB))

		// Draw placed circles
		for _, c := range circles {
			canvas.FillCircle(c.X, c.Y, c.Radius, c.Color)
			canvas.DrawCircle(c.X, c.Y, c.Radius, glow.White)
		}

		// Draw player with pulsing effect
		pulse := 5 * math.Sin(float64(frame)*0.1)
		canvas.FillCircle(int(playerX), int(playerY), int(20+pulse), glow.Green)
		canvas.DrawCircle(int(playerX), int(playerY), int(25+pulse), glow.Yellow)

		// Draw instructions
		canvas.DrawRect(10, 10, 200, 80, glow.RGB(0, 0, 0))
		canvas.DrawRectOutline(10, 10, 200, 80, glow.Gray)

		// Draw frame counter in corner
		for i := 0; i < 5; i++ {
			x := 15 + (frame/10+i)%180
			canvas.SetPixel(x, 50, glow.White)
		}

		// Present to screen
		if err := win.Present(); err != nil {
			log.Printf("Present error: %v", err)
		}

		frame++
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	fmt.Printf("\nExited after %d frames\n", frame)
	fmt.Printf("Placed %d circles\n", len(circles))
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
