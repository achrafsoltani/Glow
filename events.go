package glow

import (
	"github.com/AchrafSoltani/glow/internal/x11"
)

// EventType identifies the type of event
type EventType int

const (
	EventNone EventType = iota
	EventQuit
	EventKeyDown
	EventKeyUp
	EventMouseButtonDown
	EventMouseButtonUp
	EventMouseMotion
	EventWindowResize
	EventWindowExpose
)

// Event represents an input or window event
type Event struct {
	Type   EventType
	Key    Key
	Button MouseButton
	X, Y   int
	Width  int
	Height int
}

// Key represents a keyboard key (X11 keycode)
type Key uint8

// Common key codes (X11 keycodes - may vary by keyboard layout)
const (
	KeyUnknown Key = 0
	KeyEscape  Key = 9
	KeyF1      Key = 67
	KeyF2      Key = 68
	KeyF3      Key = 69
	KeyF4      Key = 70
	KeyF5      Key = 71
	KeyF6      Key = 72
	KeyF7      Key = 73
	KeyF8      Key = 74
	KeyF9      Key = 75
	KeyF10     Key = 76
	KeyF11     Key = 95
	KeyF12     Key = 96

	Key1 Key = 10
	Key2 Key = 11
	Key3 Key = 12
	Key4 Key = 13
	Key5 Key = 14
	Key6 Key = 15
	Key7 Key = 16
	Key8 Key = 17
	Key9 Key = 18
	Key0 Key = 19

	KeyQ Key = 24
	KeyW Key = 25
	KeyE Key = 26
	KeyR Key = 27
	KeyT Key = 28
	KeyY Key = 29
	KeyU Key = 30
	KeyI Key = 31
	KeyO Key = 32
	KeyP Key = 33

	KeyA Key = 38
	KeyS Key = 39
	KeyD Key = 40
	KeyF Key = 41
	KeyG Key = 42
	KeyH Key = 43
	KeyJ Key = 44
	KeyK Key = 45
	KeyL Key = 46

	KeyZ Key = 52
	KeyX Key = 53
	KeyC Key = 54
	KeyV Key = 55
	KeyB Key = 56
	KeyN Key = 57
	KeyM Key = 58

	KeySpace     Key = 65
	KeyBackspace Key = 22
	KeyTab       Key = 23
	KeyEnter     Key = 36
	KeyShiftL    Key = 50
	KeyShiftR    Key = 62
	KeyCtrlL     Key = 37
	KeyCtrlR     Key = 105
	KeyAltL      Key = 64
	KeyAltR      Key = 108

	KeyLeft  Key = 113
	KeyUp    Key = 111
	KeyRight Key = 114
	KeyDown  Key = 116
)

// MouseButton represents a mouse button
type MouseButton uint8

const (
	MouseNone       MouseButton = 0
	MouseLeft       MouseButton = 1
	MouseMiddle     MouseButton = 2
	MouseRight      MouseButton = 3
	MouseWheelUp    MouseButton = 4
	MouseWheelDown  MouseButton = 5
)

// PollEvent returns the next event, or nil if none available
// This is non-blocking - returns immediately
func (w *Window) PollEvent() *Event {
	select {
	case e := <-w.eventChan:
		// Update window dimensions if resize event
		if e.Type == EventWindowResize {
			w.width = e.Width
			w.height = e.Height
		}
		return &e
	default:
		return nil
	}
}

// WaitEvent blocks until an event is available
func (w *Window) WaitEvent() *Event {
	e := <-w.eventChan
	// Update window dimensions if resize event
	if e.Type == EventWindowResize {
		w.width = e.Width
		w.height = e.Height
	}
	return &e
}

// pollEvents runs in a goroutine, reading X11 events and sending to channel
func (w *Window) pollEvents() {
	for {
		select {
		case <-w.quitChan:
			return
		default:
			xEvent, err := w.conn.NextEvent()
			if err != nil {
				continue
			}

			if event := w.convertEvent(xEvent); event != nil {
				select {
				case w.eventChan <- *event:
				case <-w.quitChan:
					return
				default:
					// Channel full, drop event
				}
			}
		}
	}
}

func (w *Window) convertEvent(xEvent x11.Event) *Event {
	if xEvent == nil {
		return nil
	}

	switch e := xEvent.(type) {
	case x11.KeyEvent:
		evType := EventKeyDown
		if e.EventType == x11.EventKeyRelease {
			evType = EventKeyUp
		}
		return &Event{
			Type: evType,
			Key:  Key(e.Keycode),
			X:    int(e.X),
			Y:    int(e.Y),
		}

	case x11.ButtonEvent:
		evType := EventMouseButtonDown
		if e.EventType == x11.EventButtonRelease {
			evType = EventMouseButtonUp
		}
		return &Event{
			Type:   evType,
			Button: MouseButton(e.Button),
			X:      int(e.X),
			Y:      int(e.Y),
		}

	case x11.MotionEvent:
		return &Event{
			Type: EventMouseMotion,
			X:    int(e.X),
			Y:    int(e.Y),
		}

	case x11.ExposeEvent:
		return &Event{
			Type:   EventWindowExpose,
			Width:  int(e.Width),
			Height: int(e.Height),
		}

	case x11.ConfigureEvent:
		return &Event{
			Type:   EventWindowResize,
			X:      int(e.X),
			Y:      int(e.Y),
			Width:  int(e.Width),
			Height: int(e.Height),
		}

	case x11.ClientMessageEvent:
		// Check for window close button
		if x11.IsDeleteWindowEvent(e) {
			return &Event{
				Type: EventQuit,
			}
		}
	}

	return nil
}
