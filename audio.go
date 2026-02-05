package glow

import (
	"io"
	"log"

	"github.com/AchrafSoltani/glow/internal/pulse"
)

// AudioContext manages a connection to the PulseAudio server.
type AudioContext struct {
	conn       *pulse.Connection
	sampleRate uint32
	channels   uint8
	format     uint8
}

// NewAudioContext creates a new audio context connected to PulseAudio.
// sampleRate is in Hz (e.g. 44100), channels is 1 for mono or 2 for stereo,
// and bitDepth is the number of bytes per sample (2 for 16-bit).
func NewAudioContext(sampleRate, channels, bitDepth int) (*AudioContext, error) {
	conn, err := pulse.Connect()
	if err != nil {
		return nil, err
	}

	// Map bitDepth to PA sample format
	var format uint8
	switch bitDepth {
	case 1:
		format = pulse.SampleU8
	case 2:
		format = pulse.SampleS16LE
	case 3:
		format = pulse.SampleS24LE
	case 4:
		format = pulse.SampleS32LE
	default:
		format = pulse.SampleS16LE
	}

	return &AudioContext{
		conn:       conn,
		sampleRate: uint32(sampleRate),
		channels:   uint8(channels),
		format:     format,
	}, nil
}

// NewPlayer creates a new audio player that reads PCM data from r.
func (ctx *AudioContext) NewPlayer(r io.Reader) *AudioPlayer {
	return &AudioPlayer{
		ctx:    ctx,
		reader: r,
	}
}

// Close closes the audio context and its PulseAudio connection.
func (ctx *AudioContext) Close() {
	if ctx.conn != nil {
		ctx.conn.Close()
	}
}

// AudioPlayer plays PCM audio data from an io.Reader.
type AudioPlayer struct {
	ctx    *AudioContext
	reader io.Reader
}

// Play starts playback in a goroutine. It reads all data from the reader,
// creates a PulseAudio playback stream, and writes the PCM data.
// This is fire-and-forget — the stream drains naturally.
func (p *AudioPlayer) Play() {
	go func() {
		data, err := io.ReadAll(p.reader)
		if err != nil {
			log.Printf("glow audio: read error: %v", err)
			return
		}
		if len(data) == 0 {
			return
		}

		stream, err := p.ctx.conn.CreatePlaybackStream(
			p.ctx.format,
			p.ctx.channels,
			p.ctx.sampleRate,
		)
		if err != nil {
			log.Printf("glow audio: create stream error: %v", err)
			return
		}

		if err := stream.WriteAll(data); err != nil {
			log.Printf("glow audio: write error: %v", err)
		}
	}()
}
