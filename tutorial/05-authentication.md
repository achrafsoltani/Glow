# Step 5: X11 Authentication

## Goal
Implement Xauthority file reading for X11 authentication.

## What You'll Learn
- Xauthority file format
- MIT-MAGIC-COOKIE-1 authentication
- Reading binary files in Go

## 5.1 Why Authentication?

X11 servers often require authentication to prevent unauthorized access. The most common method is MIT-MAGIC-COOKIE-1, which uses a shared secret stored in `~/.Xauthority`.

## 5.2 Xauthority File Format

The file contains multiple entries:

```
┌─────────┬─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│ Family  │ Address Length  │    Address      │ Number Length   │    Number       │
│ 2 bytes │    2 bytes      │   variable      │    2 bytes      │   variable      │
├─────────┴─────────────────┴─────────────────┴─────────────────┴─────────────────┤
│ Name Length  │      Name       │ Data Length  │      Data                        │
│   2 bytes    │    variable     │   2 bytes    │    variable                      │
└──────────────┴─────────────────┴──────────────┴──────────────────────────────────┘
```

All multi-byte integers are **big-endian** (unlike X11 protocol which is little-endian).

## 5.3 Implementation

Create `internal/x11/auth.go`:

```go
package x11

import (
    "encoding/binary"
    "io"
    "os"
    "path/filepath"
)

// AuthEntry represents an Xauthority entry
type AuthEntry struct {
    Family  uint16
    Address string
    Display string
    Name    string
    Data    []byte
}

// ReadXauthority reads the Xauthority file
func ReadXauthority() ([]AuthEntry, error) {
    // Find Xauthority file
    xauthPath := os.Getenv("XAUTHORITY")
    if xauthPath == "" {
        home, err := os.UserHomeDir()
        if err != nil {
            return nil, err
        }
        xauthPath = filepath.Join(home, ".Xauthority")
    }

    file, err := os.Open(xauthPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var entries []AuthEntry

    for {
        entry, err := readAuthEntry(file)
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        entries = append(entries, entry)
    }

    return entries, nil
}

func readAuthEntry(r io.Reader) (AuthEntry, error) {
    var entry AuthEntry

    // Family (2 bytes, big-endian!)
    if err := binary.Read(r, binary.BigEndian, &entry.Family); err != nil {
        return entry, err
    }

    // Address
    addr, err := readCountedString(r)
    if err != nil {
        return entry, err
    }
    entry.Address = string(addr)

    // Display number
    display, err := readCountedString(r)
    if err != nil {
        return entry, err
    }
    entry.Display = string(display)

    // Auth name (e.g., "MIT-MAGIC-COOKIE-1")
    name, err := readCountedString(r)
    if err != nil {
        return entry, err
    }
    entry.Name = string(name)

    // Auth data (the cookie)
    data, err := readCountedString(r)
    if err != nil {
        return entry, err
    }
    entry.Data = data

    return entry, nil
}

func readCountedString(r io.Reader) ([]byte, error) {
    var length uint16
    if err := binary.Read(r, binary.BigEndian, &length); err != nil {
        return nil, err
    }

    data := make([]byte, length)
    if _, err := io.ReadFull(r, data); err != nil {
        return nil, err
    }

    return data, nil
}
```

## 5.4 Finding the Right Entry

```go
// FindAuth finds authentication for a display
func FindAuth(entries []AuthEntry, displayNum string) *AuthEntry {
    const (
        FamilyLocal     = 256
        FamilyWild      = 65535
        FamilyLocalHost = 252
    )

    hostname, _ := os.Hostname()

    for i := range entries {
        e := &entries[i]

        // Check display number
        if e.Display != displayNum && e.Display != "" {
            continue
        }

        // Check family/address
        switch e.Family {
        case FamilyLocal:
            if e.Address == hostname || e.Address == "" {
                return e
            }
        case FamilyWild:
            return e
        case FamilyLocalHost:
            return e
        default:
            if e.Address == hostname || e.Address == "localhost" || e.Address == "" {
                return e
            }
        }
    }

    return nil
}
```

## 5.5 Update Connection Handshake

Update `conn.go` to use authentication:

```go
func (c *Connection) handshake() error {
    // Read Xauthority for authentication
    var authName, authData []byte

    entries, err := ReadXauthority()
    if err == nil {
        if auth := FindAuth(entries, "0"); auth != nil {
            authName = []byte(auth.Name)
            authData = auth.Data
        }
    }

    // Calculate padding (must be multiple of 4)
    authNamePad := (4 - (len(authName) % 4)) % 4
    authDataPad := (4 - (len(authData) % 4)) % 4

    // Build setup request
    setupLen := 12 + len(authName) + authNamePad + len(authData) + authDataPad
    setup := make([]byte, setupLen)

    setup[0] = 'l'  // Little-endian
    binary.LittleEndian.PutUint16(setup[2:], 11)     // Major version
    binary.LittleEndian.PutUint16(setup[4:], 0)      // Minor version
    binary.LittleEndian.PutUint16(setup[6:], uint16(len(authName)))
    binary.LittleEndian.PutUint16(setup[8:], uint16(len(authData)))

    // Copy auth name and data (with padding)
    copy(setup[12:], authName)
    copy(setup[12+len(authName)+authNamePad:], authData)

    if _, err := c.conn.Write(setup); err != nil {
        return fmt.Errorf("failed to send setup: %w", err)
    }

    // ... rest of handshake
}
```

## 5.6 Understanding the Cookie

The MIT-MAGIC-COOKIE-1 is simply a random 16-byte value:

```bash
# View your cookie
xauth list
# Output: hostname/unix:0  MIT-MAGIC-COOKIE-1  a1b2c3d4e5f6...

# The cookie is sent verbatim - server compares it
```

## 5.7 Testing

After implementing authentication, the connection should work:

```bash
go run examples/connect/main.go
# Should connect without "authentication required" error
```

## Common Issues

### "Authorization required, but no authorization protocol specified"
- Xauthority file not found
- Wrong display number in FindAuth
- Cookie doesn't match

### Debug Tips

```go
// Add debug output
entries, _ := ReadXauthority()
for _, e := range entries {
    fmt.Printf("Family=%d Address=%s Display=%s Name=%s\n",
        e.Family, e.Address, e.Display, e.Name)
}
```

## Security Note

The Xauthority file contains sensitive data. Never log or expose the cookie bytes in production code.

## Next Step
Continue to Step 6 to implement window creation.
