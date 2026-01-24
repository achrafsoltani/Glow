# Step 7: Window Properties (Title & Close Button)

## Goal
Add window title and handle the close button properly.

## What You'll Learn
- X11 atoms
- Window properties
- Window manager integration

## 7.1 What Are Atoms?

Atoms are X11's way of efficiently representing strings. Instead of sending "WM_NAME" as a string every time, you convert it to a number once and use that.

```
"WM_NAME"          → Atom 39
"WM_DELETE_WINDOW" → Atom 428
"_NET_WM_NAME"     → Atom 322
```

## 7.2 InternAtom Request

Create `internal/x11/atoms.go`:

```go
package x11

import (
    "encoding/binary"
    "io"
)

// Atom is an X11 atom (interned string)
type Atom uint32

// Common atoms
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
    }
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint16(req[4:], uint16(nameLen))
    copy(req[8:], nameBytes)

    if _, err := c.conn.Write(req); err != nil {
        return 0, err
    }

    // Read reply (32 bytes)
    reply := make([]byte, 32)
    if _, err := io.ReadFull(c.conn, reply); err != nil {
        return 0, err
    }

    atom := Atom(binary.LittleEndian.Uint32(reply[8:12]))
    return atom, nil
}
```

## 7.3 Initialize Common Atoms

```go
// InitAtoms initializes commonly used atoms
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
```

Call this in `Connect()` after the handshake.

## 7.4 ChangeProperty Request

```go
const OpChangeProperty = 18

// ChangeProperty sets a window property
func (c *Connection) ChangeProperty(window uint32, property, propType Atom,
    format uint8, data []byte) error {

    dataLen := len(data)
    padding := (4 - (dataLen % 4)) % 4

    reqLen := 6 + (dataLen+padding)/4
    req := make([]byte, reqLen*4)

    req[0] = OpChangeProperty
    req[1] = 0  // Mode: Replace
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    binary.LittleEndian.PutUint32(req[4:], window)
    binary.LittleEndian.PutUint32(req[8:], uint32(property))
    binary.LittleEndian.PutUint32(req[12:], uint32(propType))
    req[16] = format  // 8, 16, or 32 bits per element
    binary.LittleEndian.PutUint32(req[20:], uint32(dataLen/(int(format)/8)))
    copy(req[24:], data)

    _, err := c.conn.Write(req)
    return err
}
```

## 7.5 Set Window Title

```go
// SetWindowTitle sets the window title
func (c *Connection) SetWindowTitle(window uint32, title string) error {
    titleBytes := []byte(title)

    // Set WM_NAME (classic X11)
    if err := c.ChangeProperty(window, AtomWMName, AtomString, 8, titleBytes); err != nil {
        return err
    }

    // Set _NET_WM_NAME (modern/freedesktop)
    if err := c.ChangeProperty(window, AtomNetWMName, AtomUTF8String, 8, titleBytes); err != nil {
        return err
    }

    return nil
}
```

## 7.6 Enable Close Button

The close button sends a `ClientMessage` event instead of killing your app:

```go
// EnableCloseButton registers for WM_DELETE_WINDOW
func (c *Connection) EnableCloseButton(window uint32) error {
    atomAtom, err := c.InternAtom("ATOM", false)
    if err != nil {
        return err
    }

    // Set WM_PROTOCOLS to include WM_DELETE_WINDOW
    data := make([]byte, 4)
    binary.LittleEndian.PutUint32(data, uint32(AtomWMDeleteWindow))

    return c.ChangeProperty(window, AtomWMProtocols, atomAtom, 32, data)
}

// IsDeleteWindowEvent checks if event is close button
func IsDeleteWindowEvent(e ClientMessageEvent) bool {
    if e.Format != 32 {
        return false
    }
    atom := Atom(binary.LittleEndian.Uint32(e.Data[0:4]))
    return atom == AtomWMDeleteWindow
}
```

## 7.7 Usage Example

```go
windowID, _ := conn.CreateWindow(100, 100, 800, 600)

// Set title
conn.SetWindowTitle(windowID, "My Application")

// Enable close button
conn.EnableCloseButton(windowID)

conn.MapWindow(windowID)

// In event loop:
case x11.ClientMessageEvent:
    if x11.IsDeleteWindowEvent(e) {
        running = false  // User clicked close button
    }
```

## 7.8 Other Useful Properties

```go
// Set window class (for window managers)
// Format: "instance\0class\0"
func (c *Connection) SetWindowClass(window uint32, instance, class string) error {
    wmClass, _ := c.InternAtom("WM_CLASS", false)
    data := []byte(instance + "\x00" + class + "\x00")
    return c.ChangeProperty(window, wmClass, AtomString, 8, data)
}

// Set window to stay on top
func (c *Connection) SetAlwaysOnTop(window uint32) error {
    // This requires _NET_WM_STATE and _NET_WM_STATE_ABOVE atoms
    // More complex - involves sending client messages to root window
}
```

## Window Manager Communication

The window manager (GNOME, KDE, i3, etc.) and your app communicate via:

1. **Properties**: Static data stored on windows
2. **Client Messages**: Dynamic commands/notifications

```
Your App ←→ X11 Server ←→ Window Manager
                           (decorates windows,
                            handles close/minimize)
```

## Next Step
Continue to Step 8 to implement event handling.
