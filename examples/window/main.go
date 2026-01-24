package main

import (
	"fmt"
	"log"
	"time"

	"github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
	// Connect to X11 server
	conn, err := x11.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Connected to X11")

	// Create a window at position (100, 100) with size 800x600
	windowID, err := conn.CreateWindow(100, 100, 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created window: 0x%x\n", windowID)

	// Map the window (make it visible)
	if err := conn.MapWindow(windowID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Window mapped (visible)")

	// Keep the window open for 5 seconds
	fmt.Println("Window will close in 5 seconds...")
	time.Sleep(5 * time.Second)

	// Clean up
	if err := conn.DestroyWindow(windowID); err != nil {
		log.Printf("Error destroying window: %v", err)
	}
	fmt.Println("Window destroyed")
}
