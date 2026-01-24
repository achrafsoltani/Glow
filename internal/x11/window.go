package x11

import (
	"encoding/binary"
)

// CreateWindow creates a new window and returns its ID
func (c *Connection) CreateWindow(x, y int16, width, height uint16) (uint32, error) {
	windowID := c.GenerateID()

	// We want to receive these events
	eventMask := uint32(
		ExposureMask |
		KeyPressMask |
		KeyReleaseMask |
		ButtonPressMask |
		ButtonReleaseMask |
		PointerMotionMask |
		StructureNotifyMask,
	)

	// We're setting: background pixel (black) and event mask
	valueMask := uint32(CWBackPixel | CWEventMask)
	valueCount := 2

	// Request length in 4-byte units: header (8 words) + values
	reqLen := 8 + valueCount
	req := make([]byte, reqLen*4)

	// Build the CreateWindow request
	req[0] = OpCreateWindow                                  // Opcode
	req[1] = c.RootDepth                                     // Depth (copy from root)
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))   // Request length
	binary.LittleEndian.PutUint32(req[4:], windowID)         // New window ID
	binary.LittleEndian.PutUint32(req[8:], c.RootWindow)     // Parent window
	binary.LittleEndian.PutUint16(req[12:], uint16(x))       // X position
	binary.LittleEndian.PutUint16(req[14:], uint16(y))       // Y position
	binary.LittleEndian.PutUint16(req[16:], width)           // Width
	binary.LittleEndian.PutUint16(req[18:], height)          // Height
	binary.LittleEndian.PutUint16(req[20:], 0)               // Border width
	binary.LittleEndian.PutUint16(req[22:], WindowClassInputOutput) // Window class
	binary.LittleEndian.PutUint32(req[24:], c.RootVisual)    // Visual ID
	binary.LittleEndian.PutUint32(req[28:], valueMask)       // Value mask

	// Values are written in order of the bits in valueMask
	binary.LittleEndian.PutUint32(req[32:], 0x00000000) // CWBackPixel: black
	binary.LittleEndian.PutUint32(req[36:], eventMask) // CWEventMask

	if _, err := c.conn.Write(req); err != nil {
		return 0, err
	}

	return windowID, nil
}

// MapWindow makes a window visible on screen
func (c *Connection) MapWindow(windowID uint32) error {
	req := make([]byte, 8)
	req[0] = OpMapWindow
	req[1] = 0 // Unused
	binary.LittleEndian.PutUint16(req[2:], 2) // Request length: 2 words
	binary.LittleEndian.PutUint32(req[4:], windowID)

	_, err := c.conn.Write(req)
	return err
}

// UnmapWindow hides a window
func (c *Connection) UnmapWindow(windowID uint32) error {
	req := make([]byte, 8)
	req[0] = OpUnmapWindow
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], 2)
	binary.LittleEndian.PutUint32(req[4:], windowID)

	_, err := c.conn.Write(req)
	return err
}

// DestroyWindow destroys a window and frees its resources
func (c *Connection) DestroyWindow(windowID uint32) error {
	req := make([]byte, 8)
	req[0] = OpDestroyWindow
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], 2)
	binary.LittleEndian.PutUint32(req[4:], windowID)

	_, err := c.conn.Write(req)
	return err
}
