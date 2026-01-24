# Step 4: X11 Connection

## Goal
Implement the connection to the X11 server with proper handshake.

## What You'll Learn
- Parsing the DISPLAY variable
- Sending the setup request
- Parsing the setup reply

## 4.1 Connection Structure

Create `internal/x11/conn.go`:

```go
package x11

import (
    "encoding/binary"
    "errors"
    "fmt"
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
    ScreenWidth    uint16
    ScreenHeight   uint16

    // ID generation
    nextID uint32
}
```

## 4.2 Parsing DISPLAY

```go
// Connect establishes a connection to the X11 server
func Connect() (*Connection, error) {
    display := os.Getenv("DISPLAY")
    if display == "" {
        display = ":0"
    }

    // Parse display string (e.g., ":0" or ":0.0" or "localhost:0")
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

    return c, nil
}
```

## 4.3 The Handshake

```go
func (c *Connection) handshake() error {
    // Send connection setup request
    setup := make([]byte, 12)
    setup[0] = 'l'                                   // Little-endian
    setup[1] = 0                                     // Unused
    binary.LittleEndian.PutUint16(setup[2:], 11)    // Protocol major version
    binary.LittleEndian.PutUint16(setup[4:], 0)     // Protocol minor version
    binary.LittleEndian.PutUint16(setup[6:], 0)     // Auth name length
    binary.LittleEndian.PutUint16(setup[8:], 0)     // Auth data length
    binary.LittleEndian.PutUint16(setup[10:], 0)    // Unused

    if _, err := c.conn.Write(setup); err != nil {
        return fmt.Errorf("failed to send setup: %w", err)
    }

    // Read response header
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
    case 2: // Authenticate required
        return errors.New("authentication required")
    default:
        return fmt.Errorf("unknown status: %d", status)
    }
}
```

## 4.4 Parsing Setup Success

```go
func (c *Connection) parseSetupSuccess(header []byte) error {
    // Additional data length is in header[6:8] (in 4-byte units)
    additionalLen := binary.LittleEndian.Uint16(header[6:]) * 4

    // Read the rest of the setup response
    data := make([]byte, additionalLen)
    if _, err := c.conn.Read(data); err != nil {
        return fmt.Errorf("failed to read setup data: %w", err)
    }

    // Parse important fields
    c.ResourceIDBase = binary.LittleEndian.Uint32(data[4:8])
    c.ResourceIDMask = binary.LittleEndian.Uint32(data[8:12])

    // Navigate to screen info
    vendorLen := binary.LittleEndian.Uint16(data[16:18])
    numFormats := data[21]
    numScreens := data[20]

    if numScreens == 0 {
        return errors.New("no screens available")
    }

    // Calculate offset to first screen (skip vendor and formats)
    vendorPadded := (vendorLen + 3) &^ 3  // Pad to 4 bytes
    formatSize := uint16(numFormats) * 8
    screenOffset := 32 + vendorPadded + formatSize

    // Parse first screen
    screen := data[screenOffset:]
    c.RootWindow = binary.LittleEndian.Uint32(screen[0:4])
    c.ScreenWidth = binary.LittleEndian.Uint16(screen[20:22])
    c.ScreenHeight = binary.LittleEndian.Uint16(screen[22:24])
    c.RootDepth = screen[38]
    c.RootVisual = binary.LittleEndian.Uint32(screen[32:36])

    // Initialize ID generator
    c.nextID = c.ResourceIDBase

    return nil
}
```

## 4.5 Resource ID Generation

```go
// GenerateID generates a new unique resource ID
func (c *Connection) GenerateID() uint32 {
    id := c.nextID
    c.nextID++
    return (id & c.ResourceIDMask) | c.ResourceIDBase
}

// Close closes the connection
func (c *Connection) Close() error {
    return c.conn.Close()
}
```

## 4.6 Test It

Create `examples/connect/main.go`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/yourusername/glow/internal/x11"
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
}
```

Run:
```bash
go run examples/connect/main.go
```

## 4.7 Troubleshooting

### "Connection refused"
- X11 server not running
- Wrong DISPLAY variable
- Try: `echo $DISPLAY`

### "Authentication required"
- Need to implement Xauthority reading (see Step 5)

### "Permission denied"
- Socket permissions issue
- Check: `ls -la /tmp/.X11-unix/`

## What We Learned

- X11 uses a binary protocol over Unix sockets
- The setup handshake exchanges capability information
- Resource IDs are generated client-side using server-provided base/mask

## Next Step
Continue to Step 5 to implement authentication.
