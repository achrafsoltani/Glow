package pulse

import (
	"os"
	"path/filepath"
)

const cookieSize = 256

// ReadCookie reads the PulseAudio authentication cookie.
// Search order: $PULSE_COOKIE → ~/.config/pulse/cookie → ~/.pulse-cookie
// Returns 256 zero bytes as fallback (PipeWire accepts anonymous connections).
func ReadCookie() []byte {
	// Try $PULSE_COOKIE environment variable
	if path := os.Getenv("PULSE_COOKIE"); path != "" {
		if data, err := os.ReadFile(path); err == nil && len(data) >= cookieSize {
			return data[:cookieSize]
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return make([]byte, cookieSize)
	}

	// Try ~/.config/pulse/cookie
	path := filepath.Join(home, ".config", "pulse", "cookie")
	if data, err := os.ReadFile(path); err == nil && len(data) >= cookieSize {
		return data[:cookieSize]
	}

	// Try ~/.pulse-cookie (legacy)
	path = filepath.Join(home, ".pulse-cookie")
	if data, err := os.ReadFile(path); err == nil && len(data) >= cookieSize {
		return data[:cookieSize]
	}

	// Fallback: zero cookie (PipeWire accepts this)
	return make([]byte, cookieSize)
}
