package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
	conn, err := x11.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	width, height := uint16(800), uint16(600)

	windowID, err := conn.CreateWindow(100, 100, width, height)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.DestroyWindow(windowID)

	gcID, err := conn.CreateGC(windowID)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.FreeGC(gcID)

	if err := conn.MapWindow(windowID); err != nil {
		log.Fatal(err)
	}

	// Create framebuffer for software rendering
	fb := x11.NewFramebuffer(int(width), int(height))

	fmt.Println("=== Glow Drawing Demo ===")
	fmt.Println("Watch the animation!")
	fmt.Println("Press Ctrl+C to quit")

	// Animation loop
	frame := 0
	startTime := time.Now()

	for {
		// Clear to dark blue
		fb.Clear(20, 20, 40)

		// Animated bouncing circle
		t := float64(frame) * 0.03
		cx := int(width)/2 + int(150*math.Cos(t))
		cy := int(height)/2 + int(100*math.Sin(t*1.5))
		fb.FillCircle(cx, cy, 40, 255, 100, 100)

		// Rotating lines from center
		centerX, centerY := int(width)/2, int(height)/2
		for i := 0; i < 8; i++ {
			angle := t + float64(i)*math.Pi/4
			x2 := centerX + int(200*math.Cos(angle))
			y2 := centerY + int(200*math.Sin(angle))
			// Rainbow colors
			r := uint8(128 + 127*math.Sin(angle))
			g := uint8(128 + 127*math.Sin(angle+2))
			b := uint8(128 + 127*math.Sin(angle+4))
			fb.DrawLine(centerX, centerY, x2, y2, r, g, b)
		}

		// Static shapes
		fb.DrawRect(50, 50, 80, 60, 0, 200, 0)           // Green rectangle
		fb.DrawRectOutline(50, 50, 80, 60, 255, 255, 0) // Yellow outline
		fb.DrawCircle(700, 100, 60, 100, 100, 255)      // Blue circle outline
		fb.DrawTriangle(700, 500, 750, 400, 650, 400, 255, 200, 0) // Orange triangle

		// Moving vertical bars
		for i := 0; i < 5; i++ {
			x := (frame*2 + i*160) % int(width)
			fb.DrawRect(x, 0, 3, int(height), 50, 50, 60)
		}

		// Pulsing circle
		pulseRadius := 30 + int(20*math.Sin(t*3))
		fb.FillCircle(100, int(height)-100, pulseRadius, 200, 50, 200)

		// Send framebuffer to window
		err := conn.PutImage(windowID, gcID, width, height, 0, 0,
			conn.RootDepth, fb.Pixels)
		if err != nil {
			log.Printf("PutImage error: %v", err)
		}

		frame++

		// Show FPS every second
		if frame%60 == 0 {
			elapsed := time.Since(startTime).Seconds()
			fps := float64(frame) / elapsed
			fmt.Printf("\rFrame %d, FPS: %.1f  ", frame, fps)
		}

		// Cap at ~60 FPS
		time.Sleep(16 * time.Millisecond)
	}
}
