# Extension: MIT-SHM (Shared Memory)

## Overview
Use X11 shared memory for faster rendering (avoid copying pixels over socket).

## Why MIT-SHM?

Current approach:
```
Framebuffer → PutImage → Socket → X Server → Display
              (copies entire buffer every frame)
```

With MIT-SHM:
```
Shared Memory ←→ X Server → Display
              (zero-copy, just signal to display)
```

## Performance Impact

- 800x600 @ 60fps: ~2.7GB/s through socket vs nearly zero with SHM
- Especially important for larger resolutions

## Implementation Overview

### 1. Check for MIT-SHM Extension

```go
import "golang.org/x/sys/unix"

const (
    OpShmQueryVersion = 0  // Offset in SHM extension
    OpShmAttach       = 1
    OpShmDetach       = 2
    OpShmPutImage     = 3
    OpShmCreatePixmap = 4
)

func (c *Connection) hasSHM() bool {
    // Query extension: send QueryExtension request for "MIT-SHM"
    // Parse reply to get extension base opcode
    // Return true if extension exists
}
```

### 2. Create Shared Memory Segment

```go
type SHMBuffer struct {
    shmid   int           // System V SHM ID
    addr    uintptr       // Mapped address
    size    int
    data    []byte        // Slice pointing to shared memory
    shmseg  uint32        // X11 SHM segment ID
}

func CreateSHMBuffer(conn *Connection, width, height int) (*SHMBuffer, error) {
    size := width * height * 4

    // 1. Create shared memory segment
    shmid, _, errno := unix.Syscall(
        unix.SYS_SHMGET,
        unix.IPC_PRIVATE,
        uintptr(size),
        0600|unix.IPC_CREAT,
    )
    if errno != 0 {
        return nil, errno
    }

    // 2. Attach shared memory to our process
    addr, _, errno := unix.Syscall(
        unix.SYS_SHMAT,
        shmid,
        0,
        0,
    )
    if errno != 0 {
        return nil, errno
    }

    // 3. Create byte slice from the memory
    var data []byte
    sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
    sh.Data = addr
    sh.Len = size
    sh.Cap = size

    // 4. Attach to X server
    shmseg := conn.GenerateID()
    // Send ShmAttach request...

    return &SHMBuffer{
        shmid:  int(shmid),
        addr:   addr,
        size:   size,
        data:   data,
        shmseg: shmseg,
    }, nil
}
```

### 3. ShmAttach Request

```go
func (c *Connection) shmAttach(shmseg uint32, shmid int, readOnly bool) error {
    req := make([]byte, 16)
    req[0] = c.shmOpcode + OpShmAttach
    req[1] = 0
    binary.LittleEndian.PutUint16(req[2:], 4)  // Length
    binary.LittleEndian.PutUint32(req[4:], shmseg)
    binary.LittleEndian.PutUint32(req[8:], uint32(shmid))
    if readOnly {
        req[12] = 1
    }

    _, err := c.conn.Write(req)
    return err
}
```

### 4. ShmPutImage Request

```go
func (c *Connection) ShmPutImage(drawable, gc, shmseg uint32,
    width, height uint16, depth uint8) error {

    req := make([]byte, 40)
    req[0] = c.shmOpcode + OpShmPutImage
    req[1] = ImageFormatZPixmap
    binary.LittleEndian.PutUint16(req[2:], 10)  // Length
    binary.LittleEndian.PutUint32(req[4:], drawable)
    binary.LittleEndian.PutUint32(req[8:], gc)
    binary.LittleEndian.PutUint16(req[12:], width)   // Total width
    binary.LittleEndian.PutUint16(req[14:], height)  // Total height
    binary.LittleEndian.PutUint16(req[16:], 0)       // Src X
    binary.LittleEndian.PutUint16(req[18:], 0)       // Src Y
    binary.LittleEndian.PutUint16(req[20:], width)   // Src width
    binary.LittleEndian.PutUint16(req[22:], height)  // Src height
    binary.LittleEndian.PutUint16(req[24:], 0)       // Dst X
    binary.LittleEndian.PutUint16(req[26:], 0)       // Dst Y
    req[28] = depth
    req[29] = ImageFormatZPixmap
    req[30] = 0  // send_event
    binary.LittleEndian.PutUint32(req[32:], shmseg)
    binary.LittleEndian.PutUint32(req[36:], 0)  // Offset

    _, err := c.conn.Write(req)
    return err
}
```

### 5. Cleanup

```go
func (buf *SHMBuffer) Destroy(conn *Connection) {
    // Detach from X server
    // ShmDetach request...

    // Detach from process
    unix.Syscall(unix.SYS_SHMDT, buf.addr, 0, 0)

    // Mark for deletion
    unix.Syscall(unix.SYS_SHMCTL, uintptr(buf.shmid), unix.IPC_RMID, 0)
}
```

### 6. Integration with Framebuffer

```go
type FastFramebuffer struct {
    shmBuffer *SHMBuffer
    width     int
    height    int
}

func NewFastFramebuffer(conn *Connection, width, height int) (*FastFramebuffer, error) {
    if !conn.hasSHM() {
        return nil, errors.New("MIT-SHM not available")
    }

    buf, err := CreateSHMBuffer(conn, width, height)
    if err != nil {
        return nil, err
    }

    return &FastFramebuffer{
        shmBuffer: buf,
        width:     width,
        height:    height,
    }, nil
}

// SetPixel writes directly to shared memory
func (fb *FastFramebuffer) SetPixel(x, y int, r, g, b uint8) {
    if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
        return
    }
    offset := (y*fb.width + x) * 4
    fb.shmBuffer.data[offset] = b
    fb.shmBuffer.data[offset+1] = g
    fb.shmBuffer.data[offset+2] = r
}

// Present is now very fast
func (fb *FastFramebuffer) Present(conn *Connection, window, gc uint32) error {
    return conn.ShmPutImage(window, gc, fb.shmBuffer.shmseg,
        uint16(fb.width), uint16(fb.height), conn.RootDepth)
}
```

## Fallback Strategy

```go
type Framebuffer struct {
    fast   *FastFramebuffer  // nil if SHM unavailable
    normal *x11.Framebuffer  // Fallback
}

func (fb *Framebuffer) Present(conn *Connection, window, gc uint32) error {
    if fb.fast != nil {
        return fb.fast.Present(conn, window, gc)
    }
    return conn.PutImage(window, gc, ...)  // Regular method
}
```

## Files to Create

1. `internal/x11/shm.go` - SHM extension support
2. `internal/x11/shm_linux.go` - Linux syscalls
3. Update `framebuffer.go` - Use SHM when available

## Testing

```bash
# Check if MIT-SHM is available
xdpyinfo | grep MIT-SHM
```

## Resources

- MIT-SHM Extension: https://www.x.org/releases/current/doc/xextproto/shm.html
- System V SHM: https://man7.org/linux/man-pages/man2/shmget.2.html
