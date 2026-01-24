package main

import (
	"fmt"
	"log"

	"github.com/AchrafSoltani/glow/internal/x11"
)

func main() {
	conn, err := x11.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to X11!")
	fmt.Printf("Screen size: %dx%d\n", conn.ScreenWidth, conn.ScreenHeight)
	fmt.Printf("Root window: 0x%x\n", conn.RootWindow)
	fmt.Printf("Root depth: %d bits\n", conn.RootDepth)
	fmt.Printf("Root visual: 0x%x\n", conn.RootVisual)
	fmt.Printf("Resource ID base: 0x%x\n", conn.ResourceIDBase)
	fmt.Printf("Resource ID mask: 0x%x\n", conn.ResourceIDMask)
}
