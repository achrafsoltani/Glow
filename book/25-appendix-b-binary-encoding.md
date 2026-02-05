# Appendix B: Binary Encoding

Reference for binary data manipulation in Go.

## B.1 Byte Order

### Little-Endian (Most Common)

Least significant byte first. Used by x86, AMD64, ARM (usually).

```
Value: 0x12345678

Memory: [78] [56] [34] [12]
         ^lowest address
```

### Big-Endian

Most significant byte first. Used by network protocols, some RISC.

```
Value: 0x12345678

Memory: [12] [34] [56] [78]
         ^lowest address
```

## B.2 Go's encoding/binary Package

### Writing Values

```go
import "encoding/binary"

// Write uint16 (2 bytes)
buf := make([]byte, 2)
binary.LittleEndian.PutUint16(buf, 0x1234)
// buf = [0x34, 0x12]

// Write uint32 (4 bytes)
buf := make([]byte, 4)
binary.LittleEndian.PutUint32(buf, 0x12345678)
// buf = [0x78, 0x56, 0x34, 0x12]

// Write uint64 (8 bytes)
buf := make([]byte, 8)
binary.LittleEndian.PutUint64(buf, 0x123456789ABCDEF0)
```

### Reading Values

```go
// Read uint16
value := binary.LittleEndian.Uint16(buf[0:2])

// Read uint32
value := binary.LittleEndian.Uint32(buf[0:4])

// Read uint64
value := binary.LittleEndian.Uint64(buf[0:8])
```

## B.3 Signed Integers

Go's binary package works with unsigned types. Convert for signed:

```go
// Write int16
buf := make([]byte, 2)
binary.LittleEndian.PutUint16(buf, uint16(int16Value))

// Read int16
value := int16(binary.LittleEndian.Uint16(buf))

// Write int32
binary.LittleEndian.PutUint32(buf, uint32(int32Value))

// Read int32
value := int32(binary.LittleEndian.Uint32(buf))
```

## B.4 Padding and Alignment

X11 requires 4-byte alignment for requests:

```go
func pad4(length int) int {
    return (4 - (length % 4)) % 4
}

// Example: string of length 5 needs 3 bytes padding
name := "Hello"
padding := pad4(len(name))  // 3
totalLen := len(name) + padding  // 8
```

## B.5 Building X11 Requests

Pattern for constructing requests:

```go
func buildRequest(opcode uint8, data1 uint32, data2 uint16) []byte {
    // Calculate length in 4-byte units
    reqLen := 3  // 12 bytes / 4

    req := make([]byte, reqLen*4)

    // Header
    req[0] = opcode                                    // Byte 0: opcode
    req[1] = 0                                         // Byte 1: often unused
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))  // Bytes 2-3: length

    // Data
    binary.LittleEndian.PutUint32(req[4:], data1)     // Bytes 4-7
    binary.LittleEndian.PutUint16(req[8:], data2)     // Bytes 8-9
    // Bytes 10-11: padding (already zero)

    return req
}
```

## B.6 Common Patterns

### Fixed-Size Request

```go
req := make([]byte, 16)  // 4 units * 4 bytes
req[0] = opcode
binary.LittleEndian.PutUint16(req[2:], 4)  // Length = 4 units
// Fill remaining 12 bytes with data
```

### Variable-Size Request

```go
func buildVariableRequest(opcode uint8, fixedData []byte, varData []byte) []byte {
    padding := pad4(len(varData))
    totalLen := 8 + len(varData) + padding  // 8 = header + fixed

    reqLen := totalLen / 4
    req := make([]byte, totalLen)

    req[0] = opcode
    binary.LittleEndian.PutUint16(req[2:], uint16(reqLen))
    copy(req[4:8], fixedData)
    copy(req[8:], varData)

    return req
}
```

### Reading Variable-Length Responses

```go
// Read 32-byte header
header := make([]byte, 32)
io.ReadFull(conn, header)

// Check for additional data
additionalLength := binary.LittleEndian.Uint32(header[4:8])
if additionalLength > 0 {
    additional := make([]byte, additionalLength*4)
    io.ReadFull(conn, additional)
}
```

## B.7 Bit Manipulation

### Setting Bits

```go
// Set bit n
mask |= (1 << n)

// Example: Set event mask bits
eventMask := uint32(0)
eventMask |= (1 << 0)   // KeyPressMask
eventMask |= (1 << 2)   // ButtonPressMask
eventMask |= (1 << 15)  // ExposureMask
```

### Checking Bits

```go
// Check if bit n is set
if value & (1 << n) != 0 {
    // Bit is set
}

// Example: Check modifier keys
if state & ShiftMask != 0 {
    // Shift is held
}
```

### Clearing Bits

```go
// Clear bit n
mask &^= (1 << n)

// Example: Remove a flag
eventMask &^= PointerMotionMask
```

## B.8 Pixel Format Conversion

### RGB to BGRX (X11 format)

```go
func rgbToBGRX(r, g, b uint8) []byte {
    return []byte{b, g, r, 0}
}
```

### RGBA to BGRA (with alpha)

```go
func rgbaToBGRA(r, g, b, a uint8) []byte {
    return []byte{b, g, r, a}
}
```

### Packed 32-bit Color

```go
// Pack RGB into uint32 (for X11 foreground/background)
func packRGB(r, g, b uint8) uint32 {
    return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

// Unpack uint32 to RGB
func unpackRGB(packed uint32) (r, g, b uint8) {
    r = uint8((packed >> 16) & 0xFF)
    g = uint8((packed >> 8) & 0xFF)
    b = uint8(packed & 0xFF)
    return
}
```

## B.9 Unsafe Pointer Tricks

For shared memory and performance-critical code:

```go
import (
    "reflect"
    "unsafe"
)

// Create Go slice from raw pointer
func ptrToSlice(ptr unsafe.Pointer, length int) []byte {
    var slice []byte
    header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
    header.Data = uintptr(ptr)
    header.Len = length
    header.Cap = length
    return slice
}

// Get pointer from slice
func sliceToPtr(slice []byte) unsafe.Pointer {
    return unsafe.Pointer(&slice[0])
}
```

**Warning**: Unsafe code bypasses Go's safety guarantees. Use only when necessary.

## B.10 Debugging Binary Data

### Hex Dump

```go
import "encoding/hex"

func hexDump(data []byte) string {
    return hex.Dump(data)
}

// Output:
// 00000000  55 00 04 00 00 00 01 00  00 00 00 00 ff ff ff 00  |U...............|
```

### Binary Representation

```go
func binaryString(value uint32) string {
    return fmt.Sprintf("%032b", value)
}

// Output: 00000000000000001111111111111111
```

### Field-by-Field Dump

```go
func dumpRequest(req []byte) {
    fmt.Printf("Opcode: %d\n", req[0])
    fmt.Printf("Unused: %d\n", req[1])
    fmt.Printf("Length: %d units (%d bytes)\n",
        binary.LittleEndian.Uint16(req[2:4]),
        binary.LittleEndian.Uint16(req[2:4])*4)

    for i := 4; i < len(req); i += 4 {
        fmt.Printf("Offset %d: 0x%08X\n", i,
            binary.LittleEndian.Uint32(req[i:]))
    }
}
```
