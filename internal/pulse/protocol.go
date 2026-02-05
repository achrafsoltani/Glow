package pulse

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// PulseAudio native protocol command IDs
// These must match the enum in pulsecore/native-common.h exactly.
const (
	CmdError                = 0
	CmdTimeout              = 1
	CmdReply                = 2
	CmdCreatePlaybackStream = 3
	CmdDeletePlaybackStream = 4
	CmdCreateRecordStream   = 5
	CmdDeleteRecordStream   = 6
	CmdExit                 = 7
	CmdAuth                 = 8
	CmdSetClientName        = 9
	CmdDrainPlaybackStream  = 12
	CmdRequest              = 61
)

// Sample formats
const (
	SampleU8        = 0
	SampleALaw      = 1
	SampleULaw      = 2
	SampleS16LE     = 3
	SampleS16BE     = 4
	SampleFloat32LE = 5
	SampleFloat32BE = 6
	SampleS32LE     = 7
	SampleS32BE     = 8
	SampleS24LE     = 9
	SampleS24BE     = 10
	SampleS2432LE   = 11
	SampleS2432BE   = 12
)

// Channel positions
const (
	ChannelMono      = 0
	ChannelFrontLeft = 1
	ChannelFrontRight = 2
)

// Tag types used in the PulseAudio tagged protocol
const (
	TagStringNull = 'N'
	TagU32        = 'L'
	TagS64        = 'R'
	TagSampleSpec = 'a'
	TagArbitrary  = 'x'
	TagBoolTrue   = '1'
	TagBoolFalse  = '0'
	TagU8         = 'B'
	TagString     = 't'
	TagChannelMap = 'm'
	TagCVolume    = 'v'
	TagPropList   = 'P'
	TagFormatInfo = 'f'
)

// Protocol version we advertise (35 is widely supported)
const ProtocolVersion = 35

// ControlChannel is the channel ID used for control messages
const ControlChannel = 0xFFFFFFFF

// DescriptorSize is the size of a PA frame descriptor
const DescriptorSize = 20

// Errors
var (
	ErrServerError = errors.New("pulse: server returned error")
	ErrProtocol    = errors.New("pulse: protocol error")
)

// TagBuilder accumulates tagged values into a byte slice.
type TagBuilder struct {
	buf []byte
}

// NewTagBuilder creates a new TagBuilder.
func NewTagBuilder() *TagBuilder {
	return &TagBuilder{}
}

// AddU32 appends a TAG_U32 value.
func (tb *TagBuilder) AddU32(v uint32) {
	tb.buf = append(tb.buf, TagU32)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	tb.buf = append(tb.buf, b...)
}

// AddS64 appends a TAG_S64 value.
func (tb *TagBuilder) AddS64(v int64) {
	tb.buf = append(tb.buf, TagS64)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	tb.buf = append(tb.buf, b...)
}

// AddString appends a TAG_STRING value (null-terminated).
func (tb *TagBuilder) AddString(s string) {
	tb.buf = append(tb.buf, TagString)
	tb.buf = append(tb.buf, []byte(s)...)
	tb.buf = append(tb.buf, 0)
}

// AddStringNull appends a TAG_STRING_NULL (null/empty string marker).
func (tb *TagBuilder) AddStringNull() {
	tb.buf = append(tb.buf, TagStringNull)
}

// AddBool appends a boolean tag.
func (tb *TagBuilder) AddBool(v bool) {
	if v {
		tb.buf = append(tb.buf, TagBoolTrue)
	} else {
		tb.buf = append(tb.buf, TagBoolFalse)
	}
}

// AddU8 appends a TAG_U8 value.
func (tb *TagBuilder) AddU8(v uint8) {
	tb.buf = append(tb.buf, TagU8)
	tb.buf = append(tb.buf, v)
}

// AddArbitrary appends a TAG_ARBITRARY (length-prefixed raw bytes).
func (tb *TagBuilder) AddArbitrary(data []byte) {
	tb.buf = append(tb.buf, TagArbitrary)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(len(data)))
	tb.buf = append(tb.buf, b...)
	tb.buf = append(tb.buf, data...)
}

// AddSampleSpec appends a TAG_SAMPLE_SPEC (format, channels, rate).
func (tb *TagBuilder) AddSampleSpec(format uint8, channels uint8, rate uint32) {
	tb.buf = append(tb.buf, TagSampleSpec)
	tb.buf = append(tb.buf, format)
	tb.buf = append(tb.buf, channels)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, rate)
	tb.buf = append(tb.buf, b...)
}

// AddChannelMap appends a TAG_CHANNEL_MAP.
func (tb *TagBuilder) AddChannelMap(channels uint8, positions []uint8) {
	tb.buf = append(tb.buf, TagChannelMap)
	tb.buf = append(tb.buf, channels)
	tb.buf = append(tb.buf, positions...)
}

// AddCVolume appends a TAG_CVOLUME (per-channel volumes).
func (tb *TagBuilder) AddCVolume(channels uint8, volume uint32) {
	tb.buf = append(tb.buf, TagCVolume)
	tb.buf = append(tb.buf, channels)
	for i := uint8(0); i < channels; i++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, volume)
		tb.buf = append(tb.buf, b...)
	}
}

// AddPropList appends a TAG_PROPLIST with key-value pairs.
func (tb *TagBuilder) AddPropList(props map[string]string) {
	tb.buf = append(tb.buf, TagPropList)
	for k, v := range props {
		// Key as string tag
		tb.buf = append(tb.buf, TagString)
		tb.buf = append(tb.buf, []byte(k)...)
		tb.buf = append(tb.buf, 0)

		// Length of value (including null terminator)
		vBytes := append([]byte(v), 0)
		tb.buf = append(tb.buf, TagU32)
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(len(vBytes)))
		tb.buf = append(tb.buf, b...)

		// Value as arbitrary
		tb.buf = append(tb.buf, TagArbitrary)
		binary.BigEndian.PutUint32(b, uint32(len(vBytes)))
		tb.buf = append(tb.buf, b...)
		tb.buf = append(tb.buf, vBytes...)
	}
	// Terminate with TAG_STRING_NULL
	tb.buf = append(tb.buf, TagStringNull)
}

// Bytes returns the accumulated tag data.
func (tb *TagBuilder) Bytes() []byte {
	return tb.buf
}

// TagParser reads tagged values from server replies.
type TagParser struct {
	data []byte
	pos  int
}

// NewTagParser creates a parser for the given data.
func NewTagParser(data []byte) *TagParser {
	return &TagParser{data: data}
}

// Remaining returns how many bytes are left.
func (tp *TagParser) Remaining() int {
	return len(tp.data) - tp.pos
}

// ReadU32 reads a TAG_U32.
func (tp *TagParser) ReadU32() (uint32, error) {
	if tp.pos >= len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading U32 tag byte")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagU32 {
		return 0, fmt.Errorf("pulse: expected TAG_U32 (0x%02x), got 0x%02x", TagU32, tag)
	}
	if tp.pos+4 > len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading U32 value")
	}
	v := binary.BigEndian.Uint32(tp.data[tp.pos:])
	tp.pos += 4
	return v, nil
}

// ReadS64 reads a TAG_S64.
func (tp *TagParser) ReadS64() (int64, error) {
	if tp.pos >= len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading S64 tag byte")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagS64 {
		return 0, fmt.Errorf("pulse: expected TAG_S64 (0x%02x), got 0x%02x", TagS64, tag)
	}
	if tp.pos+8 > len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading S64 value")
	}
	v := int64(binary.BigEndian.Uint64(tp.data[tp.pos:]))
	tp.pos += 8
	return v, nil
}

// ReadBool reads a boolean tag.
func (tp *TagParser) ReadBool() (bool, error) {
	if tp.pos >= len(tp.data) {
		return false, fmt.Errorf("pulse: unexpected end of data reading bool")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	switch tag {
	case TagBoolTrue:
		return true, nil
	case TagBoolFalse:
		return false, nil
	default:
		return false, fmt.Errorf("pulse: expected bool tag, got 0x%02x", tag)
	}
}

// ReadString reads a TAG_STRING.
func (tp *TagParser) ReadString() (string, error) {
	if tp.pos >= len(tp.data) {
		return "", fmt.Errorf("pulse: unexpected end of data reading string tag byte")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag == TagStringNull {
		return "", nil
	}
	if tag != TagString {
		return "", fmt.Errorf("pulse: expected TAG_STRING (0x%02x), got 0x%02x", TagString, tag)
	}
	// Find null terminator
	start := tp.pos
	for tp.pos < len(tp.data) && tp.data[tp.pos] != 0 {
		tp.pos++
	}
	if tp.pos >= len(tp.data) {
		return "", fmt.Errorf("pulse: string not null-terminated")
	}
	s := string(tp.data[start:tp.pos])
	tp.pos++ // skip null byte
	return s, nil
}

// ReadArbitrary reads a TAG_ARBITRARY.
func (tp *TagParser) ReadArbitrary() ([]byte, error) {
	if tp.pos >= len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading arbitrary tag byte")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagArbitrary {
		return nil, fmt.Errorf("pulse: expected TAG_ARBITRARY, got 0x%02x", tag)
	}
	if tp.pos+4 > len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading arbitrary length")
	}
	length := binary.BigEndian.Uint32(tp.data[tp.pos:])
	tp.pos += 4
	if tp.pos+int(length) > len(tp.data) {
		return nil, fmt.Errorf("pulse: arbitrary data truncated")
	}
	data := make([]byte, length)
	copy(data, tp.data[tp.pos:tp.pos+int(length)])
	tp.pos += int(length)
	return data, nil
}

// ReadSampleSpec reads a TAG_SAMPLE_SPEC.
func (tp *TagParser) ReadSampleSpec() (format uint8, channels uint8, rate uint32, err error) {
	if tp.pos >= len(tp.data) {
		return 0, 0, 0, fmt.Errorf("pulse: unexpected end of data reading sample spec tag")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagSampleSpec {
		return 0, 0, 0, fmt.Errorf("pulse: expected TAG_SAMPLE_SPEC, got 0x%02x", tag)
	}
	if tp.pos+6 > len(tp.data) {
		return 0, 0, 0, fmt.Errorf("pulse: unexpected end of data reading sample spec")
	}
	format = tp.data[tp.pos]
	channels = tp.data[tp.pos+1]
	rate = binary.BigEndian.Uint32(tp.data[tp.pos+2:])
	tp.pos += 6
	return
}

// ReadChannelMap reads a TAG_CHANNEL_MAP.
func (tp *TagParser) ReadChannelMap() ([]uint8, error) {
	if tp.pos >= len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading channel map tag")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagChannelMap {
		return nil, fmt.Errorf("pulse: expected TAG_CHANNEL_MAP, got 0x%02x", tag)
	}
	if tp.pos >= len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading channel map count")
	}
	count := tp.data[tp.pos]
	tp.pos++
	if tp.pos+int(count) > len(tp.data) {
		return nil, fmt.Errorf("pulse: channel map data truncated")
	}
	positions := make([]uint8, count)
	copy(positions, tp.data[tp.pos:tp.pos+int(count)])
	tp.pos += int(count)
	return positions, nil
}

// ReadCVolume reads a TAG_CVOLUME.
func (tp *TagParser) ReadCVolume() ([]uint32, error) {
	if tp.pos >= len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading cvolume tag")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagCVolume {
		return nil, fmt.Errorf("pulse: expected TAG_CVOLUME, got 0x%02x", tag)
	}
	if tp.pos >= len(tp.data) {
		return nil, fmt.Errorf("pulse: unexpected end of data reading cvolume count")
	}
	count := tp.data[tp.pos]
	tp.pos++
	vols := make([]uint32, count)
	for i := uint8(0); i < count; i++ {
		if tp.pos+4 > len(tp.data) {
			return nil, fmt.Errorf("pulse: cvolume data truncated")
		}
		vols[i] = binary.BigEndian.Uint32(tp.data[tp.pos:])
		tp.pos += 4
	}
	return vols, nil
}

// ReadU8 reads a TAG_U8.
func (tp *TagParser) ReadU8() (uint8, error) {
	if tp.pos >= len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading U8 tag byte")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagU8 {
		return 0, fmt.Errorf("pulse: expected TAG_U8, got 0x%02x", tag)
	}
	if tp.pos >= len(tp.data) {
		return 0, fmt.Errorf("pulse: unexpected end of data reading U8 value")
	}
	v := tp.data[tp.pos]
	tp.pos++
	return v, nil
}

// ReadFormatInfo reads a TAG_FORMAT_INFO.
func (tp *TagParser) ReadFormatInfo() error {
	if tp.pos >= len(tp.data) {
		return fmt.Errorf("pulse: unexpected end of data reading format info tag")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagFormatInfo {
		return fmt.Errorf("pulse: expected TAG_FORMAT_INFO, got 0x%02x", tag)
	}
	// encoding
	if tp.pos >= len(tp.data) {
		return fmt.Errorf("pulse: unexpected end of data reading format info U8")
	}
	tp.pos++ // skip encoding U8 tag
	tp.pos++ // skip encoding value
	// proplist follows — skip it
	return tp.SkipPropList()
}

// SkipPropList reads and discards a proplist.
func (tp *TagParser) SkipPropList() error {
	if tp.pos >= len(tp.data) {
		return fmt.Errorf("pulse: unexpected end of data reading proplist tag")
	}
	tag := tp.data[tp.pos]
	tp.pos++
	if tag != TagPropList {
		return fmt.Errorf("pulse: expected TAG_PROPLIST, got 0x%02x", tag)
	}
	for {
		if tp.pos >= len(tp.data) {
			return fmt.Errorf("pulse: proplist not terminated")
		}
		if tp.data[tp.pos] == TagStringNull {
			tp.pos++
			return nil
		}
		// Read key string
		if _, err := tp.ReadString(); err != nil {
			return err
		}
		// Read length U32
		if _, err := tp.ReadU32(); err != nil {
			return err
		}
		// Read arbitrary value
		if _, err := tp.ReadArbitrary(); err != nil {
			return err
		}
	}
}

// Skip advances past the next tagged value without interpreting it.
func (tp *TagParser) Skip() error {
	if tp.pos >= len(tp.data) {
		return fmt.Errorf("pulse: nothing to skip")
	}
	tag := tp.data[tp.pos]
	switch tag {
	case TagU32:
		tp.pos += 5 // tag + 4 bytes
	case TagS64:
		tp.pos += 9 // tag + 8 bytes
	case TagU8:
		tp.pos += 2 // tag + 1 byte
	case TagBoolTrue, TagBoolFalse:
		tp.pos += 1
	case TagStringNull:
		tp.pos += 1
	case TagString:
		tp.pos++ // skip tag
		for tp.pos < len(tp.data) && tp.data[tp.pos] != 0 {
			tp.pos++
		}
		if tp.pos < len(tp.data) {
			tp.pos++ // skip null
		}
	case TagArbitrary:
		tp.pos++ // skip tag
		if tp.pos+4 > len(tp.data) {
			return fmt.Errorf("pulse: truncated arbitrary in skip")
		}
		length := binary.BigEndian.Uint32(tp.data[tp.pos:])
		tp.pos += 4 + int(length)
	case TagSampleSpec:
		tp.pos += 7 // tag + 1 + 1 + 4
	case TagChannelMap:
		tp.pos++ // tag
		if tp.pos >= len(tp.data) {
			return fmt.Errorf("pulse: truncated channel map in skip")
		}
		count := tp.data[tp.pos]
		tp.pos += 1 + int(count)
	case TagCVolume:
		tp.pos++ // tag
		if tp.pos >= len(tp.data) {
			return fmt.Errorf("pulse: truncated cvolume in skip")
		}
		count := tp.data[tp.pos]
		tp.pos += 1 + int(count)*4
	case TagPropList:
		return tp.SkipPropList()
	case TagFormatInfo:
		return tp.ReadFormatInfo()
	default:
		return fmt.Errorf("pulse: unknown tag 0x%02x at position %d, cannot skip", tag, tp.pos)
	}
	return nil
}

// BuildDescriptor creates a 20-byte PA frame descriptor.
// Format: length(4) + channel(4) + offset_hi(4) + offset_lo(4) + flags(4)
func BuildDescriptor(length uint32, channel uint32) []byte {
	desc := make([]byte, DescriptorSize)
	binary.BigEndian.PutUint32(desc[0:], length)
	binary.BigEndian.PutUint32(desc[4:], channel)
	// offset_hi, offset_lo, flags all zero
	return desc
}

// BuildCommand creates a complete PA control frame: descriptor + command tag + tag seq + payload.
func BuildCommand(command uint32, tag uint32, payload []byte) []byte {
	// The payload is: TAG_U32(command) + TAG_U32(tag) + additional payload tags
	tb := NewTagBuilder()
	tb.AddU32(command)
	tb.AddU32(tag)
	cmdPayload := append(tb.Bytes(), payload...)

	desc := BuildDescriptor(uint32(len(cmdPayload)), ControlChannel)
	return append(desc, cmdPayload...)
}
