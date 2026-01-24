package x11

import (
	"encoding/binary"
)

// Graphics Context value masks
const (
	GCFunction          = 1 << 0
	GCForeground        = 1 << 2
	GCBackground        = 1 << 3
	GCLineWidth         = 1 << 4
	GCGraphicsExposures = 1 << 16
)

// CreateGC creates a graphics context for drawing
func (c *Connection) CreateGC(drawable uint32) (uint32, error) {
	gcID := c.GenerateID()

	// Set foreground (white), background (black), and disable
	// graphics exposures (we don't want Expose events from drawing)
	valueMask := uint32(GCForeground | GCBackground | GCGraphicsExposures)
	valueCount := 3

	reqLen := 4 + valueCount
	req := make([]byte, reqLen*4)

	req[0] = OpCreateGC
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
	binary.LittleEndian.PutUint32(req[4:], gcID)
	binary.LittleEndian.PutUint32(req[8:], drawable)
	binary.LittleEndian.PutUint32(req[12:], valueMask)
	binary.LittleEndian.PutUint32(req[16:], 0xFFFFFF) // Foreground: white
	binary.LittleEndian.PutUint32(req[20:], 0x000000) // Background: black
	binary.LittleEndian.PutUint32(req[24:], 0)        // GraphicsExposures: off

	if _, err := c.conn.Write(req); err != nil {
		return 0, err
	}

	return gcID, nil
}

// FreeGC frees a graphics context
func (c *Connection) FreeGC(gcID uint32) error {
	req := make([]byte, 8)
	req[0] = OpFreeGC
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], 2)
	binary.LittleEndian.PutUint32(req[4:], gcID)

	_, err := c.conn.Write(req)
	return err
}

// PutImage sends pixel data to a drawable (window or pixmap)
// This is the core function for software rendering
func (c *Connection) PutImage(drawable, gc uint32, width, height uint16,
	dstX, dstY int16, depth uint8, data []byte) error {

	// Data must be padded to 4-byte boundary
	dataLen := len(data)
	padding := (4 - (dataLen % 4)) % 4

	// Request length in 4-byte units
	reqLen := 6 + (dataLen+padding)/4
	req := make([]byte, reqLen*4)

	req[0] = OpPutImage
	req[1] = ImageFormatZPixmap // ZPixmap = raw pixel data
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
	binary.LittleEndian.PutUint32(req[4:], drawable)
	binary.LittleEndian.PutUint32(req[8:], gc)
	binary.LittleEndian.PutUint16(req[12:], width)
	binary.LittleEndian.PutUint16(req[14:], height)
	binary.LittleEndian.PutUint16(req[16:], uint16(dstX))
	binary.LittleEndian.PutUint16(req[18:], uint16(dstY))
	req[20] = 0     // Left pad (unused for ZPixmap)
	req[21] = depth // Bits per pixel
	binary.LittleEndian.PutUint16(req[22:], 0) // Unused

	// Copy pixel data
	copy(req[24:], data)

	_, err := c.conn.Write(req)
	return err
}

// Rectangle for fill operations
type Rectangle struct {
	X, Y          int16
	Width, Height uint16
}

// FillRectangles fills rectangles with the GC's foreground color
func (c *Connection) FillRectangles(drawable, gc uint32, rects []Rectangle) error {
	reqLen := 3 + len(rects)*2
	req := make([]byte, reqLen*4)

	req[0] = OpPolyFillRect
	req[1] = 0
	binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
	binary.LittleEndian.PutUint32(req[4:], drawable)
	binary.LittleEndian.PutUint32(req[8:], gc)

	offset := 12
	for _, r := range rects {
		binary.LittleEndian.PutUint16(req[offset:], uint16(r.X))
		binary.LittleEndian.PutUint16(req[offset+2:], uint16(r.Y))
		binary.LittleEndian.PutUint16(req[offset+4:], r.Width)
		binary.LittleEndian.PutUint16(req[offset+6:], r.Height)
		offset += 8
	}

	_, err := c.conn.Write(req)
	return err
}
