package x11

import (
	"encoding/binary"
	"io"
)

// Event is the interface for all X11 events
type Event interface {
	Type() int
}

// KeyEvent represents a key press or release
type KeyEvent struct {
	EventType int
	Keycode   uint8
	State     uint16 // Modifier state (shift, ctrl, etc.)
	X, Y      int16  // Position relative to window
	RootX     int16  // Position relative to root window
	RootY     int16
}

func (e KeyEvent) Type() int { return e.EventType }

// ButtonEvent represents a mouse button press or release
type ButtonEvent struct {
	EventType int
	Button    uint8  // 1=left, 2=middle, 3=right, 4=wheel up, 5=wheel down
	State     uint16 // Modifier state
	X, Y      int16
	RootX     int16
	RootY     int16
}

func (e ButtonEvent) Type() int { return e.EventType }

// MotionEvent represents mouse movement
type MotionEvent struct {
	X, Y  int16
	RootX int16
	RootY int16
	State uint16 // Which buttons/modifiers are held
}

func (e MotionEvent) Type() int { return EventMotionNotify }

// ExposeEvent means part of the window needs redrawing
type ExposeEvent struct {
	Window uint32
	X, Y   uint16
	Width  uint16
	Height uint16
	Count  uint16 // Number of Expose events to follow
}

func (e ExposeEvent) Type() int { return EventExpose }

// ConfigureEvent means the window was resized or moved
type ConfigureEvent struct {
	Window uint32
	X, Y   int16
	Width  uint16
	Height uint16
}

func (e ConfigureEvent) Type() int { return EventConfigureNotify }

// ClientMessageEvent is used for window manager communication
type ClientMessageEvent struct {
	Window    uint32
	Format    uint8
	MessageType uint32
	Data      [20]byte
}

func (e ClientMessageEvent) Type() int { return EventClientMessage }

// UnknownEvent for events we don't handle yet
type UnknownEvent struct {
	EventType int
	Data      [32]byte
}

func (e UnknownEvent) Type() int { return e.EventType }

// NextEvent blocks until an event is received, then returns it
func (c *Connection) NextEvent() (Event, error) {
	// All X11 events are exactly 32 bytes
	buf := make([]byte, 32)
	_, err := io.ReadFull(c.conn, buf)
	if err != nil {
		return nil, err
	}

	// Event type is in first byte (high bit is "sent by SendEvent")
	eventType := int(buf[0] & 0x7F)

	switch eventType {
	case EventKeyPress, EventKeyRelease:
		return KeyEvent{
			EventType: eventType,
			Keycode:   buf[1],
			State:     binary.LittleEndian.Uint16(buf[28:30]),
			X:         int16(binary.LittleEndian.Uint16(buf[24:26])),
			Y:         int16(binary.LittleEndian.Uint16(buf[26:28])),
			RootX:     int16(binary.LittleEndian.Uint16(buf[20:22])),
			RootY:     int16(binary.LittleEndian.Uint16(buf[22:24])),
		}, nil

	case EventButtonPress, EventButtonRelease:
		return ButtonEvent{
			EventType: eventType,
			Button:    buf[1],
			State:     binary.LittleEndian.Uint16(buf[28:30]),
			X:         int16(binary.LittleEndian.Uint16(buf[24:26])),
			Y:         int16(binary.LittleEndian.Uint16(buf[26:28])),
			RootX:     int16(binary.LittleEndian.Uint16(buf[20:22])),
			RootY:     int16(binary.LittleEndian.Uint16(buf[22:24])),
		}, nil

	case EventMotionNotify:
		return MotionEvent{
			X:     int16(binary.LittleEndian.Uint16(buf[24:26])),
			Y:     int16(binary.LittleEndian.Uint16(buf[26:28])),
			RootX: int16(binary.LittleEndian.Uint16(buf[20:22])),
			RootY: int16(binary.LittleEndian.Uint16(buf[22:24])),
			State: binary.LittleEndian.Uint16(buf[28:30]),
		}, nil

	case EventExpose:
		return ExposeEvent{
			Window: binary.LittleEndian.Uint32(buf[4:8]),
			X:      binary.LittleEndian.Uint16(buf[8:10]),
			Y:      binary.LittleEndian.Uint16(buf[10:12]),
			Width:  binary.LittleEndian.Uint16(buf[12:14]),
			Height: binary.LittleEndian.Uint16(buf[14:16]),
			Count:  binary.LittleEndian.Uint16(buf[16:18]),
		}, nil

	case EventConfigureNotify:
		return ConfigureEvent{
			Window: binary.LittleEndian.Uint32(buf[4:8]),
			X:      int16(binary.LittleEndian.Uint16(buf[16:18])),
			Y:      int16(binary.LittleEndian.Uint16(buf[18:20])),
			Width:  binary.LittleEndian.Uint16(buf[20:22]),
			Height: binary.LittleEndian.Uint16(buf[22:24]),
		}, nil

	case EventClientMessage:
		e := ClientMessageEvent{
			Window:      binary.LittleEndian.Uint32(buf[4:8]),
			Format:      buf[1],
			MessageType: binary.LittleEndian.Uint32(buf[8:12]),
		}
		copy(e.Data[:], buf[12:32])
		return e, nil

	default:
		e := UnknownEvent{EventType: eventType}
		copy(e.Data[:], buf)
		return e, nil
	}
}
