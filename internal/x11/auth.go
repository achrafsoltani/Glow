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

	// Family (2 bytes, big-endian)
	if err := binary.Read(r, binary.BigEndian, &entry.Family); err != nil {
		return entry, err
	}

	// Address
	addr, err := readString(r)
	if err != nil {
		return entry, err
	}
	entry.Address = string(addr)

	// Display number
	display, err := readString(r)
	if err != nil {
		return entry, err
	}
	entry.Display = string(display)

	// Auth name
	name, err := readString(r)
	if err != nil {
		return entry, err
	}
	entry.Name = string(name)

	// Auth data
	data, err := readString(r)
	if err != nil {
		return entry, err
	}
	entry.Data = data

	return entry, nil
}

func readString(r io.Reader) ([]byte, error) {
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

// FindAuth finds authentication for a display
func FindAuth(entries []AuthEntry, displayNum string) *AuthEntry {
	// Family values
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
			// For other families, check if address matches
			if e.Address == hostname || e.Address == "localhost" || e.Address == "" {
				return e
			}
		}
	}

	return nil
}
