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

	windowID, err := conn.CreateWindow(100, 100, 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.DestroyWindow(windowID)

	if err := conn.MapWindow(windowID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Glow Event Demo ===")
	fmt.Println("Click, move mouse, press keys in the window")
	fmt.Println("Press 'q' (keycode 24) or 'Escape' (keycode 9) to quit")
	fmt.Println()

	// Event loop
	for {
		event, err := conn.NextEvent()
		if err != nil {
			log.Printf("Error reading event: %v", err)
			break
		}

		switch e := event.(type) {
		case x11.KeyEvent:
			action := "pressed"
			if e.EventType == x11.EventKeyRelease {
				action = "released"
			}
			fmt.Printf("Key %s: keycode=%d at (%d, %d)\n",
				action, e.Keycode, e.X, e.Y)

			// Quit on 'q' (24) or Escape (9)
			if e.EventType == x11.EventKeyPress {
				if e.Keycode == 24 || e.Keycode == 9 {
					fmt.Println("\nQuitting...")
					return
				}
			}

		case x11.ButtonEvent:
			action := "pressed"
			if e.EventType == x11.EventButtonRelease {
				action = "released"
			}
			buttonName := ""
			switch e.Button {
			case 1:
				buttonName = "Left"
			case 2:
				buttonName = "Middle"
			case 3:
				buttonName = "Right"
			case 4:
				buttonName = "WheelUp"
			case 5:
				buttonName = "WheelDown"
			default:
				buttonName = fmt.Sprintf("Button%d", e.Button)
			}
			fmt.Printf("Mouse %s %s at (%d, %d)\n",
				buttonName, action, e.X, e.Y)

		case x11.MotionEvent:
			// Uncomment to see mouse motion (very verbose!)
			// fmt.Printf("Mouse moved to (%d, %d)\n", e.X, e.Y)

		case x11.ExposeEvent:
			fmt.Printf("Window exposed: region (%d,%d) size %dx%d\n",
				e.X, e.Y, e.Width, e.Height)

		case x11.ConfigureEvent:
			fmt.Printf("Window configured: pos (%d,%d) size %dx%d\n",
				e.X, e.Y, e.Width, e.Height)

		case x11.UnknownEvent:
			// Silently ignore unknown events
		}
	}
}
