package x11

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Atom is an X11 atom (interned string identifier)
type Atom uint32

// Common atoms we'll need
var (
	AtomWMProtocols    Atom
	AtomWMDeleteWindow Atom
	AtomWMName         Atom
	AtomString         Atom
	AtomUTF8String     Atom
	AtomNetWMName      Atom
)

// InternAtom converts a string to an atom
func (c *Connection) InternAtom(name string, onlyIfExists bool) (Atom, error) {
	nameBytes := []byte(name)
	nameLen := len(nameBytes)
	padding := (4 - (nameLen % 4)) % 4

	reqLen := 2 + (nameLen+padding)/4
	req := make([]byte, reqLen*4)

	req[0] = OpInternAtom
	if onlyIfExists {
		req[1] = 1
	} else {
		req[1] = 0
	}
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
	binary.LittleEndian.PutUint16(req[4:], uint16(nameLen))
	binary.LittleEndian.PutUint16(req[6:], 0) // Unused
	copy(req[8:], nameBytes)

	if _, err := c.conn.Write(req); err != nil {
		return 0, err
	}

	// Read reply (32 bytes)
	reply := make([]byte, 32)
	if _, err := io.ReadFull(c.conn, reply); err != nil {
		return 0, err
	}

	// Check for error
	if reply[0] == 0 {
		return 0, fmt.Errorf("InternAtom failed for %s", name)
	}

	atom := Atom(binary.LittleEndian.Uint32(reply[8:12]))
	return atom, nil
}

// InitAtoms initializes common atoms
func (c *Connection) InitAtoms() error {
	var err error

	AtomWMProtocols, err = c.InternAtom("WM_PROTOCOLS", false)
	if err != nil {
		return err
	}

	AtomWMDeleteWindow, err = c.InternAtom("WM_DELETE_WINDOW", false)
	if err != nil {
		return err
	}

	AtomWMName, err = c.InternAtom("WM_NAME", false)
	if err != nil {
		return err
	}

	AtomString, err = c.InternAtom("STRING", false)
	if err != nil {
		return err
	}

	AtomUTF8String, err = c.InternAtom("UTF8_STRING", false)
	if err != nil {
		return err
	}

	AtomNetWMName, err = c.InternAtom("_NET_WM_NAME", false)
	if err != nil {
		return err
	}

	return nil
}

// ChangeProperty sets a window property
func (c *Connection) ChangeProperty(window uint32, property, propType Atom,
	format uint8, data []byte) error {

	dataLen := len(data)
	padding := (4 - (dataLen % 4)) % 4

	reqLen := 6 + (dataLen+padding)/4
	req := make([]byte, reqLen*4)

	req[0] = OpChangeProperty
	req[1] = 0 // Mode: Replace
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
	binary.LittleEndian.PutUint32(req[4:], window)
	binary.LittleEndian.PutUint32(req[8:], uint32(property))
	binary.LittleEndian.PutUint32(req[12:], uint32(propType))
	req[16] = format
	// req[17:20] unused
	binary.LittleEndian.PutUint32(req[20:], uint32(dataLen/(int(format)/8)))
	copy(req[24:], data)

	_, err := c.conn.Write(req)
	return err
}

// SetWindowTitle sets the window title
func (c *Connection) SetWindowTitle(window uint32, title string) error {
	titleBytes := []byte(title)

	// Set WM_NAME (classic)
	if err := c.ChangeProperty(window, AtomWMName, AtomString, 8, titleBytes); err != nil {
		return err
	}

	// Set _NET_WM_NAME (modern)
	if err := c.ChangeProperty(window, AtomNetWMName, AtomUTF8String, 8, titleBytes); err != nil {
		return err
	}

	return nil
}

// EnableCloseButton registers for WM_DELETE_WINDOW messages
func (c *Connection) EnableCloseButton(window uint32) error {
	// Get ATOM type
	atomAtom, err := c.InternAtom("ATOM", false)
	if err != nil {
		return err
	}

	// Set WM_PROTOCOLS property to include WM_DELETE_WINDOW
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(AtomWMDeleteWindow))

	return c.ChangeProperty(window, AtomWMProtocols, atomAtom, 32, data)
}

// IsDeleteWindowEvent checks if a ClientMessage is WM_DELETE_WINDOW
func IsDeleteWindowEvent(e ClientMessageEvent) bool {
	if e.Format != 32 {
		return false
	}
	atom := Atom(binary.LittleEndian.Uint32(e.Data[0:4]))
	return atom == AtomWMDeleteWindow
}
