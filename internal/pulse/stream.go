package pulse

import (
	"fmt"
)

// Stream represents a PulseAudio playback stream.
type Stream struct {
	conn    *Connection
	channel uint32 // server-assigned data channel ID
}

// CreatePlaybackStream creates a new playback stream.
func (c *Connection) CreatePlaybackStream(format uint8, channels uint8, rate uint32) (*Stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tag := c.nextTag
	c.nextTag++

	// Build channel map positions
	positions := make([]uint8, channels)
	if channels == 1 {
		positions[0] = ChannelMono
	} else if channels >= 2 {
		positions[0] = ChannelFrontLeft
		positions[1] = ChannelFrontRight
		for i := uint8(2); i < channels; i++ {
			positions[i] = 0
		}
	}

	tb := NewTagBuilder()

	// sample_spec
	tb.AddSampleSpec(format, channels, rate)

	// channel_map
	tb.AddChannelMap(channels, positions)

	// sink_index (PA_INVALID_INDEX = 0xFFFFFFFF means default)
	tb.AddU32(0xFFFFFFFF)

	// sink_name (null = default)
	tb.AddStringNull()

	// Buffer attributes: maxlength, corked, tlength, prebuf, minreq
	tb.AddU32(0xFFFFFFFF) // maxlength (server default)
	tb.AddBool(false)     // corked (start playing immediately)
	tb.AddU32(0xFFFFFFFF) // tlength (server default)
	tb.AddU32(0)          // prebuf (0 = immediate playback, no buffering delay)
	tb.AddU32(0xFFFFFFFF) // minreq (server default)

	// sync_id (0 = none)
	tb.AddU32(0)

	// cvolume
	tb.AddCVolume(channels, 0x10000) // PA_VOLUME_NORM = 0x10000

	// Since protocol >= 12: no_remap, no_remix, fix_format, fix_rate, fix_channels,
	// no_move, variable_rate
	tb.AddBool(false) // no_remap
	tb.AddBool(false) // no_remix
	tb.AddBool(false) // fix_format
	tb.AddBool(false) // fix_rate
	tb.AddBool(false) // fix_channels
	tb.AddBool(false) // no_move
	tb.AddBool(false) // variable_rate

	// Since protocol >= 13: muted, adjust_latency, proplist
	tb.AddBool(false) // muted
	tb.AddBool(true)  // adjust_latency
	tb.AddPropList(map[string]string{
		"media.name": "playback",
	})

	// Since protocol >= 14: volume_set, early_requests
	tb.AddBool(true)  // volume_set
	tb.AddBool(false) // early_requests

	// Since protocol >= 15: muted_set, dont_inhibit_auto_suspend, fail_on_suspend
	tb.AddBool(false) // muted_set
	tb.AddBool(false) // dont_inhibit_auto_suspend
	tb.AddBool(false) // fail_on_suspend

	// Since protocol >= 17: relative_volume
	tb.AddBool(false) // relative_volume

	// Since protocol >= 18: passthrough
	tb.AddBool(false) // passthrough

	// Since protocol >= 21: n_formats, format_info[]
	// Send 1 format matching our sample spec
	tb.AddU8(1)                              // n_formats
	tb.buf = append(tb.buf, TagFormatInfo)   // TAG_FORMAT_INFO
	tb.buf = append(tb.buf, TagU8, 1)        // encoding = PA_ENCODING_PCM (1)
	tb.AddPropList(map[string]string{})      // empty proplist for format info

	frame := BuildCommand(CmdCreatePlaybackStream, tag, tb.Bytes())

	if _, err := c.conn.Write(frame); err != nil {
		return nil, fmt.Errorf("pulse: create_playback_stream write: %w", err)
	}

	// The server may interleave data-request frames before the reply,
	// so we need to drain until we get our control reply.
	replyCmd, _, tp, err := c.drainForReply()
	if err != nil {
		return nil, fmt.Errorf("pulse: create_playback_stream read: %w", err)
	}
	if replyCmd == CmdError {
		code, _ := tp.ReadU32()
		return nil, fmt.Errorf("pulse: create_playback_stream error (code %d)", code)
	}
	if replyCmd != CmdReply {
		return nil, fmt.Errorf("pulse: create_playback_stream unexpected response %d", replyCmd)
	}

	// Parse reply: stream_index, sink_input_index, missing (requested_bytes)
	// then sample_spec, channel_map, buffer_attrs, etc.
	streamIndex, err := tp.ReadU32()
	if err != nil {
		return nil, fmt.Errorf("pulse: parse stream_index: %w", err)
	}
	_ = streamIndex

	sinkInputIndex, err := tp.ReadU32()
	if err != nil {
		return nil, fmt.Errorf("pulse: parse sink_input_index: %w", err)
	}
	_ = sinkInputIndex

	// missing = how many bytes the server wants immediately
	_, err = tp.ReadU32()
	if err != nil {
		return nil, fmt.Errorf("pulse: parse missing: %w", err)
	}

	return &Stream{
		conn:    c,
		channel: streamIndex,
	}, nil
}

// drainForReply reads frames until a control reply is received.
// Must be called with c.mu held.
func (c *Connection) drainForReply() (cmd uint32, tag uint32, tp *TagParser, err error) {
	return c.DrainReplies()
}

// WriteAll writes all PCM data to the stream.
func (s *Stream) WriteAll(data []byte) error {
	return s.conn.WriteData(s.channel, data)
}
