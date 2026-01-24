package x11

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// Connection represents a connection to the X11 server
type Connection struct {
	conn net.Conn

	// Setup information from server
	ResourceIDBase uint32
	ResourceIDMask uint32
	RootWindow     uint32
	RootVisual     uint32
	RootDepth      uint8
	BitsPerPixel   uint8 // Bits per pixel for RootDepth
	ScreenWidth    uint16
	ScreenHeight   uint16

	// ID generation
	nextID uint32
}

// Connect establishes a connection to the X11 server
func Connect() (*Connection, error) {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}

	// Parse display string (e.g., ":0" or ":0.0")
	displayNum := "0"
	if idx := strings.Index(display, ":"); idx != -1 {
		rest := display[idx+1:]
		if dotIdx := strings.Index(rest, "."); dotIdx != -1 {
			displayNum = rest[:dotIdx]
		} else {
			displayNum = rest
		}
	}

	// Connect via Unix socket
	socketPath := fmt.Sprintf("/tmp/.X11-unix/X%s", displayNum)
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X11: %w", err)
	}

	c := &Connection{conn: conn}

	if err := c.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	// Initialize atoms for window manager integration
	if err := c.InitAtoms(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to init atoms: %w", err)
	}

	return c, nil
}

// Close closes the connection
func (c *Connection) Close() error {
	return c.conn.Close()
}

// Write writes raw bytes to the X11 connection
func (c *Connection) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

// Reader returns the underlying connection for reading
func (c *Connection) Reader() io.Reader {
	return c.conn
}

func (c *Connection) handshake() error {
	// Read Xauthority for authentication
	var authName, authData []byte

	entries, err := ReadXauthority()
	if err == nil {
		// Try to find auth for display 0
		if auth := FindAuth(entries, "0"); auth != nil {
			authName = []byte(auth.Name)
			authData = auth.Data
		}
	}

	// Calculate padding
	authNamePad := (4 - (len(authName) % 4)) % 4
	authDataPad := (4 - (len(authData) % 4)) % 4

	// Send connection setup request
	// Byte order: 'l' for little-endian, 'B' for big-endian
	setupLen := 12 + len(authName) + authNamePad + len(authData) + authDataPad
	setup := make([]byte, setupLen)
	setup[0] = 'l'                                              // Little-endian
	setup[1] = 0                                                // Unused
	binary.LittleEndian.PutUint16(setup[2:], 11)               // Protocol major version
	binary.LittleEndian.PutUint16(setup[4:], 0)                // Protocol minor version
	binary.LittleEndian.PutUint16(setup[6:], uint16(len(authName)))  // Auth protocol name length
	binary.LittleEndian.PutUint16(setup[8:], uint16(len(authData)))  // Auth data length
	binary.LittleEndian.PutUint16(setup[10:], 0)               // Unused

	// Copy auth name and data
	copy(setup[12:], authName)
	copy(setup[12+len(authName)+authNamePad:], authData)

	if _, err := c.conn.Write(setup); err != nil {
		return fmt.Errorf("failed to send setup: %w", err)
	}

	// Read response header (8 bytes minimum)
	header := make([]byte, 8)
	if _, err := c.conn.Read(header); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	status := header[0]
	switch status {
	case 0: // Failed
		reasonLen := header[1]
		reason := make([]byte, reasonLen)
		c.conn.Read(reason)
		return fmt.Errorf("connection failed: %s", string(reason))
	case 1: // Success
		return c.parseSetupSuccess(header)
	case 2: // Authenticate
		return errors.New("authentication required (not implemented)")
	default:
		return fmt.Errorf("unknown status: %d", status)
	}
}

func (c *Connection) parseSetupSuccess(header []byte) error {
	// Additional data length is in header[6:8] (in 4-byte units)
	additionalLen := binary.LittleEndian.Uint16(header[6:]) * 4

	// Read the rest of the setup response
	data := make([]byte, additionalLen)
	if _, err := c.conn.Read(data); err != nil {
		return fmt.Errorf("failed to read setup data: %w", err)
	}

	// Parse the setup response
	c.ResourceIDBase = binary.LittleEndian.Uint32(data[4:8])
	c.ResourceIDMask = binary.LittleEndian.Uint32(data[8:12])

	// Skip to screen info
	vendorLen := binary.LittleEndian.Uint16(data[16:18])
	numFormats := data[21]
	numScreens := data[20]

	if numScreens == 0 {
		return errors.New("no screens available")
	}

	// Calculate offset to first screen
	// Vendor string is padded to 4-byte boundary
	vendorPadded := (vendorLen + 3) &^ 3
	formatSize := uint16(numFormats) * 8
	screenOffset := 32 + vendorPadded + formatSize

	// Parse first screen
	screen := data[screenOffset:]
	c.RootWindow = binary.LittleEndian.Uint32(screen[0:4])
	c.ScreenWidth = binary.LittleEndian.Uint16(screen[20:22])
	c.ScreenHeight = binary.LittleEndian.Uint16(screen[22:24])
	c.RootDepth = screen[38]
	c.RootVisual = binary.LittleEndian.Uint32(screen[32:36])

	// Parse pixmap formats to find bits-per-pixel for our depth
	// Formats start at offset 32 + vendorPadded
	formatOffset := 32 + int(vendorPadded)
	for i := 0; i < int(numFormats); i++ {
		fmtData := data[formatOffset+i*8:]
		depth := fmtData[0]
		bpp := fmtData[1]
		if depth == c.RootDepth {
			c.BitsPerPixel = bpp
			break
		}
	}

	// Default to 32 bpp if not found
	if c.BitsPerPixel == 0 {
		c.BitsPerPixel = 32
	}

	// Initialize ID generator
	c.nextID = c.ResourceIDBase

	return nil
}

// GenerateID generates a new resource ID
func (c *Connection) GenerateID() uint32 {
	id := c.nextID
	c.nextID++
	return (id & c.ResourceIDMask) | c.ResourceIDBase
}

// Sync sends a GetInputFocus request and waits for the reply
// This ensures all previous requests have been processed
func (c *Connection) Sync() error {
	req := make([]byte, 4)
	req[0] = 43 // GetInputFocus opcode
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], 1) // Length

	if _, err := c.conn.Write(req); err != nil {
		return err
	}

	// Read reply (32 bytes)
	reply := make([]byte, 32)
	if _, err := c.conn.Read(reply); err != nil {
		return err
	}

	// Check if it's an error (first byte = 0)
	if reply[0] == 0 {
		errorCode := reply[1]
		return fmt.Errorf("X11 error: code %d", errorCode)
	}

	return nil
}
