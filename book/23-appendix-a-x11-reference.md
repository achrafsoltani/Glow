# Appendix A: X11 Protocol Reference

Quick reference for X11 opcodes, events, and data structures used in this book.

## A.1 Request Opcodes

| Opcode | Name | Description |
|--------|------|-------------|
| 1 | CreateWindow | Create a new window |
| 2 | ChangeWindowAttributes | Modify window attributes |
| 4 | DestroyWindow | Destroy a window |
| 8 | MapWindow | Make window visible |
| 10 | UnmapWindow | Hide window |
| 12 | ConfigureWindow | Change window geometry |
| 16 | InternAtom | Get atom for string |
| 17 | GetAtomName | Get string for atom |
| 18 | ChangeProperty | Set window property |
| 19 | DeleteProperty | Remove window property |
| 20 | GetProperty | Read window property |
| 43 | GetInputFocus | Get current focus (useful for sync) |
| 55 | CreateGC | Create graphics context |
| 56 | ChangeGC | Modify GC attributes |
| 60 | FreeGC | Destroy graphics context |
| 72 | PutImage | Send pixel data |
| 73 | GetImage | Read pixel data |
| 98 | QueryExtension | Check extension availability |

## A.2 Event Types

| Code | Name | Triggered By |
|------|------|--------------|
| 2 | KeyPress | Key pressed |
| 3 | KeyRelease | Key released |
| 4 | ButtonPress | Mouse button pressed |
| 5 | ButtonRelease | Mouse button released |
| 6 | MotionNotify | Mouse moved |
| 7 | EnterNotify | Mouse entered window |
| 8 | LeaveNotify | Mouse left window |
| 9 | FocusIn | Window gained focus |
| 10 | FocusOut | Window lost focus |
| 12 | Expose | Window needs redraw |
| 22 | ConfigureNotify | Window resized/moved |
| 33 | ClientMessage | Inter-client message |

## A.3 Event Masks

```go
const (
    KeyPressMask        = 1 << 0
    KeyReleaseMask      = 1 << 1
    ButtonPressMask     = 1 << 2
    ButtonReleaseMask   = 1 << 3
    EnterWindowMask     = 1 << 4
    LeaveWindowMask     = 1 << 5
    PointerMotionMask   = 1 << 6
    ButtonMotionMask    = 1 << 13
    ExposureMask        = 1 << 15
    StructureNotifyMask = 1 << 17
    FocusChangeMask     = 1 << 21
)
```

## A.4 Window Attributes

```go
const (
    CWBackPixmap       = 1 << 0
    CWBackPixel        = 1 << 1
    CWBorderPixmap     = 1 << 2
    CWBorderPixel      = 1 << 3
    CWBitGravity       = 1 << 4
    CWWinGravity       = 1 << 5
    CWBackingStore     = 1 << 6
    CWBackingPlanes    = 1 << 7
    CWBackingPixel     = 1 << 8
    CWOverrideRedirect = 1 << 9
    CWSaveUnder        = 1 << 10
    CWEventMask        = 1 << 11
    CWDontPropagate    = 1 << 12
    CWColormap         = 1 << 13
    CWCursor           = 1 << 14
)
```

## A.5 GC Attributes

```go
const (
    GCFunction          = 1 << 0
    GCPlaneMask         = 1 << 1
    GCForeground        = 1 << 2
    GCBackground        = 1 << 3
    GCLineWidth         = 1 << 4
    GCLineStyle         = 1 << 5
    GCCapStyle          = 1 << 6
    GCJoinStyle         = 1 << 7
    GCFillStyle         = 1 << 8
    GCFillRule          = 1 << 9
    GCTile              = 1 << 10
    GCStipple           = 1 << 11
    GCTileStipXOrigin   = 1 << 12
    GCTileStipYOrigin   = 1 << 13
    GCFont              = 1 << 14
    GCSubwindowMode     = 1 << 15
    GCGraphicsExposures = 1 << 16
    GCClipXOrigin       = 1 << 17
    GCClipYOrigin       = 1 << 18
    GCClipMask          = 1 << 19
    GCDashOffset        = 1 << 20
    GCDashList          = 1 << 21
    GCArcMode           = 1 << 22
)
```

## A.6 Error Codes

| Code | Name | Description |
|------|------|-------------|
| 1 | BadRequest | Invalid request code |
| 2 | BadValue | Integer parameter out of range |
| 3 | BadWindow | Window ID not valid |
| 4 | BadPixmap | Pixmap ID not valid |
| 5 | BadAtom | Atom ID not valid |
| 6 | BadCursor | Cursor ID not valid |
| 7 | BadFont | Font ID not valid |
| 8 | BadMatch | Parameter mismatch |
| 9 | BadDrawable | Drawable (window or pixmap) not valid |
| 10 | BadAccess | Access denied |
| 11 | BadAlloc | Server out of resources |
| 12 | BadColor | Colormap entry not valid |
| 13 | BadGC | GC ID not valid |
| 14 | BadIDChoice | Resource ID already in use |
| 15 | BadName | Named color/font not found |
| 16 | BadLength | Request length incorrect |
| 17 | BadImplementation | Server implementation error |

## A.7 Modifier Masks

```go
const (
    ShiftMask   = 1 << 0
    LockMask    = 1 << 1  // Caps Lock
    ControlMask = 1 << 2
    Mod1Mask    = 1 << 3  // Alt
    Mod2Mask    = 1 << 4  // Num Lock
    Mod3Mask    = 1 << 5
    Mod4Mask    = 1 << 6  // Super/Windows
    Mod5Mask    = 1 << 7
    Button1Mask = 1 << 8
    Button2Mask = 1 << 9
    Button3Mask = 1 << 10
    Button4Mask = 1 << 11
    Button5Mask = 1 << 12
)
```

## A.8 Image Formats

| Code | Name | Description |
|------|------|-------------|
| 0 | XYBitmap | 1-bit per pixel |
| 1 | XYPixmap | Planar format |
| 2 | ZPixmap | Packed pixels (most common) |

## A.9 Standard Atoms

| Atom | Name | Purpose |
|------|------|---------|
| 1 | PRIMARY | Selection |
| 2 | SECONDARY | Selection |
| 3 | ARC | Shape |
| 4 | ATOM | Type |
| ... | ... | ... |
| 31 | WM_NAME | Window title (Latin-1) |
| 32 | WM_ICON_NAME | Icon title |
| 33 | WM_CLIENT_MACHINE | Host name |

Non-standard atoms (created via InternAtom):
- `WM_PROTOCOLS` - Protocol negotiation
- `WM_DELETE_WINDOW` - Close button handling
- `_NET_WM_NAME` - Window title (UTF-8)
- `UTF8_STRING` - UTF-8 string type

## A.10 Connection Setup Response

```
Byte 0:     Success (1) / Failure (0)
Byte 1:     Unused (if success) / Reason length (if failure)
Bytes 2-3:  Protocol major version
Bytes 4-5:  Protocol minor version
Bytes 6-7:  Additional data length (in 4-byte units)

Additional data (success):
  Bytes 0-3:   Release number
  Bytes 4-7:   Resource ID base
  Bytes 8-11:  Resource ID mask
  Bytes 12-15: Motion buffer size
  Bytes 16-17: Vendor length
  Bytes 18-19: Maximum request length
  Byte 20:     Number of screens
  Byte 21:     Number of formats
  Byte 22:     Image byte order (0=LSB, 1=MSB)
  Byte 23:     Bitmap bit order
  Byte 24:     Bitmap scanline unit
  Byte 25:     Bitmap scanline pad
  Byte 26:     Min keycode
  Byte 27:     Max keycode
  Bytes 28-31: Unused
  Followed by: Vendor string, formats, screens
```

## A.11 Common Key Codes (X11/Linux)

```go
const (
    KeyEscape    = 9
    Key1         = 10
    Key2         = 11
    // ... 3-9
    Key0         = 19
    KeyMinus     = 20
    KeyEqual     = 21
    KeyBackspace = 22
    KeyTab       = 23
    KeyQ         = 24
    KeyW         = 25
    KeyE         = 26
    KeyR         = 27
    KeyT         = 28
    KeyY         = 29
    KeyU         = 30
    KeyI         = 31
    KeyO         = 32
    KeyP         = 33
    KeyEnter     = 36
    KeyA         = 38
    KeyS         = 39
    KeyD         = 40
    KeyF         = 41
    KeyG         = 42
    KeyH         = 43
    KeyJ         = 44
    KeyK         = 45
    KeyL         = 46
    KeyZ         = 52
    KeyX         = 53
    KeyC         = 54
    KeyV         = 55
    KeyB         = 56
    KeyN         = 57
    KeyM         = 58
    KeySpace     = 65
    KeyF1        = 67
    KeyF2        = 68
    // ... F3-F10
    KeyF11       = 95
    KeyF12       = 96
    KeyUp        = 111
    KeyLeft      = 113
    KeyRight     = 114
    KeyDown      = 116
)
```

Note: Key codes can vary by keyboard layout and X server configuration. Use XKB for reliable key symbol translation.
