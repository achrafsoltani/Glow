# Step 3: Understanding the X11 Protocol

## Goal
Understand the X11 protocol structure before implementing it.

## What You'll Learn
- X11 client-server architecture
- Protocol message format
- Request and reply structure

## 3.1 X11 Architecture

X11 is a **client-server** protocol (with confusing naming):

```
┌─────────────────┐         ┌─────────────────┐
│   Your App      │         │   X11 Server    │
│   (X Client)    │◄───────►│   (Display)     │
└─────────────────┘  Socket └─────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                ┌───▼───┐          ┌────▼────┐
                │Monitor│          │Keyboard │
                └───────┘          │ Mouse   │
                                   └─────────┘
```

**Confusing part**: Your application is the "client" and the display server is the "server", even though your app shows windows.

## 3.2 Connection Flow

```
1. Client connects to server (Unix socket or TCP)
2. Client sends setup request (byte order, protocol version)
3. Server replies with setup info (screen size, root window, etc.)
4. Client sends requests (create window, draw, etc.)
5. Server sends events (key press, mouse, etc.)
6. Server sends replies to queries
```

## 3.3 The DISPLAY Variable

The `DISPLAY` environment variable tells clients where to connect:

```
DISPLAY=:0          → /tmp/.X11-unix/X0 (local, display 0)
DISPLAY=:1          → /tmp/.X11-unix/X1 (local, display 1)
DISPLAY=host:0      → TCP to host, port 6000
DISPLAY=host:1      → TCP to host, port 6001
```

## 3.4 Request Format

Every X11 request follows this structure:

```
┌────────┬────────┬─────────────────┬─────────────────────────┐
│ Opcode │  Data  │ Request Length  │    Request Data         │
│ 1 byte │ 1 byte │    2 bytes      │  (length-1)*4 bytes     │
└────────┴────────┴─────────────────┴─────────────────────────┘
```

- **Opcode**: Which operation (1=CreateWindow, 8=MapWindow, etc.)
- **Data**: Opcode-specific (often unused, set to 0)
- **Length**: Total request size in 4-byte units
- **Request Data**: Parameters for the operation

### Example: MapWindow Request

```
Opcode: 8 (MapWindow)
Length: 2 (8 bytes total)

┌────┬────┬────────┬─────────────────────────┐
│ 08 │ 00 │ 02 00  │ [window ID - 4 bytes]   │
└────┴────┴────────┴─────────────────────────┘
```

## 3.5 Event Format

All events are exactly 32 bytes:

```
┌────────┬───────────────────────────────────────────────────────┐
│  Type  │                    Event Data                         │
│ 1 byte │                     31 bytes                          │
└────────┴───────────────────────────────────────────────────────┘
```

### Common Event Types

| Type | Name | Description |
|------|------|-------------|
| 2 | KeyPress | Key was pressed |
| 3 | KeyRelease | Key was released |
| 4 | ButtonPress | Mouse button pressed |
| 5 | ButtonRelease | Mouse button released |
| 6 | MotionNotify | Mouse moved |
| 12 | Expose | Window needs redraw |
| 22 | ConfigureNotify | Window resized/moved |
| 33 | ClientMessage | Message from WM |

## 3.6 Connection Setup

### Setup Request (from client)

```
┌─────────────┬────────┬─────────────┬─────────────┬──────────┬──────────┬─────────┐
│ Byte Order  │ Unused │ Major Ver.  │ Minor Ver.  │ Auth Len │ Data Len │ Unused  │
│  1 byte     │ 1 byte │  2 bytes    │  2 bytes    │ 2 bytes  │ 2 bytes  │ 2 bytes │
└─────────────┴────────┴─────────────┴─────────────┴──────────┴──────────┴─────────┘
│                                Auth Name (padded to 4 bytes)                      │
│                                Auth Data (padded to 4 bytes)                      │
```

- Byte order: 'l' (0x6C) for little-endian, 'B' (0x42) for big-endian
- Protocol version: 11.0 (major=11, minor=0)

### Setup Reply (from server)

```
┌─────────┬─────────┬─────────────┬─────────────┬────────────────┐
│ Status  │  Unused │ Major Ver.  │ Minor Ver.  │ Additional Len │
│ 1 byte  │  1 byte │  2 bytes    │  2 bytes    │    2 bytes     │
└─────────┴─────────┴─────────────┴─────────────┴────────────────┘
│                     Additional Data...                          │
│  (Contains screen info, root window, visual info, etc.)         │
```

Status codes:
- 0 = Failed
- 1 = Success
- 2 = Authenticate (need to provide auth)

## 3.7 Important Concepts

### Resource IDs
Everything in X11 has an ID (windows, pixmaps, graphics contexts). The server tells you a base and mask:

```go
// From setup reply
resourceIDBase := 0x04000000
resourceIDMask := 0x001FFFFF

// Generate new IDs
nextID := resourceIDBase
newWindowID := (nextID & resourceIDMask) | resourceIDBase
nextID++
```

### Atoms
Strings are expensive to send repeatedly, so X11 uses "atoms" - numeric IDs for strings:

```
"WM_DELETE_WINDOW" → Atom 123
"WM_NAME"          → Atom 39
"_NET_WM_NAME"     → Atom 456
```

You request an atom once with InternAtom, then use the number.

### Event Masks
Windows only receive events they ask for:

```go
eventMask := KeyPressMask | KeyReleaseMask | ExposureMask
```

## 3.8 Common Opcodes

| Opcode | Name | Description |
|--------|------|-------------|
| 1 | CreateWindow | Create a new window |
| 4 | DestroyWindow | Destroy a window |
| 8 | MapWindow | Make window visible |
| 10 | UnmapWindow | Hide window |
| 16 | InternAtom | Get atom for string |
| 18 | ChangeProperty | Set window property |
| 55 | CreateGC | Create graphics context |
| 60 | FreeGC | Free graphics context |
| 72 | PutImage | Draw pixel data |

## 3.9 Protocol Constants File

Create `internal/x11/protocol.go`:

```go
package x11

// Request opcodes
const (
    OpCreateWindow   = 1
    OpDestroyWindow  = 4
    OpMapWindow      = 8
    OpUnmapWindow    = 10
    OpInternAtom     = 16
    OpChangeProperty = 18
    OpCreateGC       = 55
    OpFreeGC         = 60
    OpPutImage       = 72
)

// Event types
const (
    EventKeyPress    = 2
    EventKeyRelease  = 3
    EventButtonPress = 4
    // ... etc
)

// Event masks
const (
    KeyPressMask    = 1 << 0
    KeyReleaseMask  = 1 << 1
    // ... etc
)
```

## Resources

- X11 Protocol Reference: https://www.x.org/releases/current/doc/xproto/x11protocol.html
- XCB Documentation: https://xcb.freedesktop.org/

## Next Step
Continue to Step 4 to implement the X11 connection.
