package pulse

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// Connection represents a connection to the PulseAudio server.
type Connection struct {
	conn          net.Conn
	mu            sync.Mutex
	nextTag       uint32
	serverVersion uint32
}

// Connect connects to the PulseAudio server and performs the handshake.
func Connect() (*Connection, error) {
	socketPath := findSocket()
	if socketPath == "" {
		return nil, fmt.Errorf("pulse: could not find PulseAudio socket")
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("pulse: dial %s: %w", socketPath, err)
	}

	c := &Connection{
		conn: conn,
	}

	if err := c.auth(); err != nil {
		conn.Close()
		return nil, err
	}

	if err := c.setClientName(); err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

// Close closes the connection.
func (c *Connection) Close() error {
	return c.conn.Close()
}

// ServerVersion returns the server's protocol version.
func (c *Connection) ServerVersion() uint32 {
	return c.serverVersion
}

// findSocket locates the PulseAudio Unix socket.
func findSocket() string {
	// Try $PULSE_SERVER
	if server := os.Getenv("PULSE_SERVER"); server != "" {
		// Handle unix: prefix
		if len(server) > 5 && server[:5] == "unix:" {
			return server[5:]
		}
		// If it starts with / treat as path
		if server[0] == '/' {
			return server
		}
	}

	// Try $XDG_RUNTIME_DIR/pulse/native
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		path := filepath.Join(runtimeDir, "pulse", "native")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Try /run/user/<uid>/pulse/native
	uid := strconv.Itoa(os.Getuid())
	path := filepath.Join("/run", "user", uid, "pulse", "native")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// auth performs the AUTH command handshake.
func (c *Connection) auth() error {
	cookie := ReadCookie()

	tb := NewTagBuilder()
	tb.AddU32(ProtocolVersion) // protocol version
	tb.AddArbitrary(cookie)    // auth cookie

	tag := c.nextTag
	c.nextTag++
	frame := BuildCommand(CmdAuth, tag, tb.Bytes())

	if _, err := c.conn.Write(frame); err != nil {
		return fmt.Errorf("pulse: auth write: %w", err)
	}

	replyCmd, _, tp, err := c.readReply()
	if err != nil {
		return fmt.Errorf("pulse: auth read: %w", err)
	}
	if replyCmd == CmdError {
		code, _ := tp.ReadU32()
		return fmt.Errorf("pulse: auth rejected (error code %d)", code)
	}
	if replyCmd != CmdReply {
		return fmt.Errorf("pulse: auth unexpected response %d", replyCmd)
	}

	// Parse server protocol version
	serverVersion, err := tp.ReadU32()
	if err != nil {
		return fmt.Errorf("pulse: auth parse version: %w", err)
	}
	c.serverVersion = serverVersion

	return nil
}

// setClientName sends SET_CLIENT_NAME to identify ourselves.
func (c *Connection) setClientName() error {
	tb := NewTagBuilder()
	tb.AddPropList(map[string]string{
		"application.name": "glow",
	})

	tag := c.nextTag
	c.nextTag++
	frame := BuildCommand(CmdSetClientName, tag, tb.Bytes())

	if _, err := c.conn.Write(frame); err != nil {
		return fmt.Errorf("pulse: set_client_name write: %w", err)
	}

	replyCmd, _, _, err := c.readReply()
	if err != nil {
		return fmt.Errorf("pulse: set_client_name read: %w", err)
	}
	if replyCmd == CmdError {
		return fmt.Errorf("pulse: set_client_name rejected")
	}
	if replyCmd != CmdReply {
		return fmt.Errorf("pulse: set_client_name unexpected response %d", replyCmd)
	}

	return nil
}

// SendCommand sends a command and returns the parsed reply.
// The caller must hold no lock — this method acquires c.mu.
func (c *Connection) SendCommand(command uint32, payload []byte) (uint32, uint32, *TagParser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tag := c.nextTag
	c.nextTag++
	frame := BuildCommand(command, tag, payload)

	if _, err := c.conn.Write(frame); err != nil {
		return 0, 0, nil, fmt.Errorf("pulse: write command %d: %w", command, err)
	}

	return c.readReply()
}

// WriteData writes raw PCM data on a stream channel.
func (c *Connection) WriteData(channel uint32, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Send data in chunks to avoid overly large writes.
	// PA accepts data frames up to 64KB typically, but let's use
	// a generous chunk size. The server tells us how much it wants
	// via requested_bytes, but for fire-and-forget we just send all.
	const maxChunk = 65536

	for len(data) > 0 {
		chunk := data
		if len(chunk) > maxChunk {
			chunk = data[:maxChunk]
		}
		data = data[len(chunk):]

		desc := BuildDescriptor(uint32(len(chunk)), channel)
		if _, err := c.conn.Write(desc); err != nil {
			return fmt.Errorf("pulse: write data descriptor: %w", err)
		}
		if _, err := c.conn.Write(chunk); err != nil {
			return fmt.Errorf("pulse: write data payload: %w", err)
		}
	}

	return nil
}

// readReply reads a single PA frame from the connection.
// Returns the command, tag, and a TagParser for the remaining payload.
func (c *Connection) readReply() (cmd uint32, tag uint32, tp *TagParser, err error) {
	// Read descriptor
	desc := make([]byte, DescriptorSize)
	if _, err = io.ReadFull(c.conn, desc); err != nil {
		return 0, 0, nil, fmt.Errorf("pulse: read descriptor: %w", err)
	}

	length := binary.BigEndian.Uint32(desc[0:4])
	channel := binary.BigEndian.Uint32(desc[4:8])

	if length == 0 {
		return 0, 0, NewTagParser(nil), nil
	}

	payload := make([]byte, length)
	if _, err = io.ReadFull(c.conn, payload); err != nil {
		return 0, 0, nil, fmt.Errorf("pulse: read payload (%d bytes): %w", length, err)
	}

	// Non-control channel — data frame, skip
	if channel != ControlChannel {
		return 0, 0, NewTagParser(nil), nil
	}

	tp = NewTagParser(payload)

	// Parse command and tag
	cmd, err = tp.ReadU32()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("pulse: parse command: %w", err)
	}
	tag, err = tp.ReadU32()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("pulse: parse tag: %w", err)
	}

	return cmd, tag, tp, nil
}

// DrainReplies reads and discards incoming frames until a REPLY or
// ERROR control message arrives. This handles interleaved async
// notifications (STARTED, REQUEST, SUBSCRIBE_EVENT, etc.) that the
// server sends before or alongside the actual reply.
func (c *Connection) DrainReplies() (cmd uint32, tag uint32, tp *TagParser, err error) {
	for {
		// Read descriptor
		desc := make([]byte, DescriptorSize)
		if _, err = io.ReadFull(c.conn, desc); err != nil {
			return 0, 0, nil, fmt.Errorf("pulse: drain read descriptor: %w", err)
		}

		length := binary.BigEndian.Uint32(desc[0:4])
		channel := binary.BigEndian.Uint32(desc[4:8])

		if length == 0 {
			continue
		}

		payload := make([]byte, length)
		if _, err = io.ReadFull(c.conn, payload); err != nil {
			return 0, 0, nil, fmt.Errorf("pulse: drain read payload: %w", err)
		}

		// Skip non-control frames (data frames on stream channels)
		if channel != ControlChannel {
			continue
		}

		tp = NewTagParser(payload)
		cmd, err = tp.ReadU32()
		if err != nil {
			return 0, 0, nil, err
		}
		tag, err = tp.ReadU32()
		if err != nil {
			return 0, 0, nil, err
		}

		// Only return on REPLY or ERROR — skip async notifications
		if cmd == CmdReply || cmd == CmdError {
			return cmd, tag, tp, nil
		}
		// Otherwise discard (STARTED, REQUEST, SUBSCRIBE_EVENT, etc.)
	}
}
