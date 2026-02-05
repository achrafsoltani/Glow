# Extension: Audio System (PulseAudio Native Protocol)

## Overview

Glow includes built-in audio support via the PulseAudio native protocol — the same pure-Go, zero-dependency, Unix-socket approach used for X11 graphics. No CGo, no external libraries, no `libasound2-dev`.

This works on any Linux system running PulseAudio or PipeWire (which exposes a PulseAudio-compatible socket).

## Architecture

```
glow/
├── audio.go                 # Public API (AudioContext, AudioPlayer)
└── internal/pulse/          # PulseAudio native protocol
    ├── protocol.go          # Constants, TagBuilder, TagParser, framing
    ├── auth.go              # Cookie reading
    ├── conn.go              # Socket connection + handshake
    └── stream.go            # Playback stream + PCM writing
```

The design mirrors `internal/x11/`:

| X11 Graphics | PulseAudio Audio |
|---|---|
| Unix socket `/tmp/.X11-unix/X0` | Unix socket `$XDG_RUNTIME_DIR/pulse/native` |
| Little-endian binary protocol | Big-endian tagged protocol |
| MIT-MAGIC-COOKIE-1 auth | 256-byte cookie auth |
| Opcodes (CreateWindow, PutImage, ...) | Commands (AUTH, CREATE_PLAYBACK_STREAM, ...) |

## Public API

```go
// Create an audio context (connects to PulseAudio)
ctx, err := glow.NewAudioContext(44100, 1, 2) // sampleRate, channels, bitDepthBytes
if err != nil {
    log.Fatal(err)
}
defer ctx.Close()

// Play PCM data from an io.Reader (fire-and-forget)
player := ctx.NewPlayer(bytes.NewReader(pcmData))
player.Play() // non-blocking, runs in a goroutine
```

### Types

```go
type AudioContext struct { ... }

func NewAudioContext(sampleRate, channels, bitDepth int) (*AudioContext, error)
func (ctx *AudioContext) NewPlayer(r io.Reader) *AudioPlayer
func (ctx *AudioContext) Close()

type AudioPlayer struct { ... }

func (p *AudioPlayer) Play()
```

### Parameters

- **sampleRate**: Sample rate in Hz (e.g. 44100, 22050)
- **channels**: 1 for mono, 2 for stereo
- **bitDepth**: Bytes per sample — 1 (U8), 2 (S16LE), 3 (S24LE), or 4 (S32LE)

## PulseAudio Native Protocol

### Transport

PulseAudio uses a Unix domain socket, typically at:

1. `$PULSE_SERVER` (if set, may have `unix:` prefix)
2. `$XDG_RUNTIME_DIR/pulse/native`
3. `/run/user/<uid>/pulse/native`

PipeWire provides the same socket at the same path.

### Authentication

A 256-byte cookie, searched in order:

1. `$PULSE_COOKIE`
2. `~/.config/pulse/cookie`
3. `~/.pulse-cookie`

PipeWire accepts a zero-filled cookie (anonymous), so we fall back to that.

### Frame Format

Every message uses a 20-byte **descriptor** header:

```
+--------+--------+-----------+-----------+-------+
| Length  | Channel | Offset Hi | Offset Lo | Flags |
| 4 bytes | 4 bytes | 4 bytes   | 4 bytes   | 4 bytes|
+--------+--------+-----------+-----------+-------+
```

All fields are **big-endian** (unlike X11's little-endian).

- **Channel** = `0xFFFFFFFF` for control messages, or a stream channel ID for data
- **Length** = size of the payload following the descriptor

### Tagged Values

The payload uses a tagged format where each value is prefixed by a single-byte type tag:

| Tag | Code | Payload |
|---|---|---|
| `TAG_U32` | `'L'` | 4-byte big-endian uint32 |
| `TAG_S64` | `'R'` | 8-byte big-endian int64 |
| `TAG_U8` | `'B'` | 1-byte value |
| `TAG_STRING` | `'t'` | null-terminated UTF-8 |
| `TAG_STRING_NULL` | `'N'` | (no payload — represents null) |
| `TAG_BOOL_TRUE` | `'1'` | (no payload) |
| `TAG_BOOL_FALSE` | `'0'` | (no payload) |
| `TAG_ARBITRARY` | `'x'` | 4-byte length + raw bytes |
| `TAG_SAMPLE_SPEC` | `'a'` | format(1) + channels(1) + rate(4) |
| `TAG_CHANNEL_MAP` | `'m'` | channels(1) + position bytes |
| `TAG_CVOLUME` | `'v'` | channels(1) + 4 bytes per channel |
| `TAG_PROPLIST` | `'P'` | key-value pairs, terminated by `TAG_STRING_NULL` |
| `TAG_FORMAT_INFO` | `'f'` | encoding(U8) + proplist |

### Command Sequence

Control messages always start with `TAG_U32(command)` + `TAG_U32(tag)`:

**1. AUTH (command=8)**

```
→ TAG_U32(protocol_version=35) + TAG_ARBITRARY(cookie, 256 bytes)
← TAG_U32(server_version)
```

**2. SET_CLIENT_NAME (command=9)**

```
→ TAG_PROPLIST({"application.name": "glow"})
← TAG_U32(client_index)
```

**3. CREATE_PLAYBACK_STREAM (command=3)**

```
→ TAG_SAMPLE_SPEC(format, channels, rate)
  TAG_CHANNEL_MAP(...)
  TAG_U32(sink_index=0xFFFFFFFF)  // default sink
  TAG_STRING_NULL                  // sink_name=null
  TAG_U32(maxlength=-1)           // server decides
  TAG_BOOL(corked=false)          // start playing immediately
  TAG_U32(tlength=-1)
  TAG_U32(prebuf=0)               // immediate playback, no buffering delay
  TAG_U32(minreq=-1)
  TAG_U32(sync_id=0)
  TAG_CVOLUME(channels, 0x10000)  // PA_VOLUME_NORM
  ... (version-conditional boolean flags)
  TAG_PROPLIST({"media.name": "playback"})
  ... (more version-conditional fields)
← TAG_U32(stream_index)           // = channel ID for data
  TAG_U32(sink_input_index)
  TAG_U32(missing)                // requested_bytes
  ...
```

Replies use command ID 2 (`PA_COMMAND_REPLY`) and errors use 0 (`PA_COMMAND_ERROR`).
The server also sends asynchronous notifications (e.g. `PA_COMMAND_REQUEST=61`,
`PA_COMMAND_STARTED=86`) which must be skipped when waiting for a reply.

**4. Data frames** (raw PCM)

After creating a stream, send PCM data using the stream's channel ID:

```
Descriptor: length=<pcm_size>, channel=<stream_index>
Payload: raw PCM bytes (no tags)
```

The server drains the stream naturally. No explicit close is needed for fire-and-forget playback.

## Implementation Details

### TagBuilder / TagParser

The `TagBuilder` accumulates tagged values into a `[]byte`, while `TagParser` reads them back:

```go
// Building a command payload
tb := pulse.NewTagBuilder()
tb.AddU32(35)                    // protocol version
tb.AddArbitrary(cookie)          // 256-byte cookie
payload := tb.Bytes()

// Parsing a reply
tp := pulse.NewTagParser(replyData)
version, _ := tp.ReadU32()
```

### Connection + Mutex

The `Connection` struct serialises concurrent access with a mutex, since multiple `Play()` goroutines may create streams simultaneously:

```go
type Connection struct {
    conn    net.Conn
    mu      sync.Mutex
    nextTag uint32
    ...
}
```

### Fire-and-Forget Playback

Each `AudioPlayer.Play()` call:

1. Spawns a goroutine
2. Reads all data from the `io.Reader`
3. Creates a new playback stream (its own channel)
4. Writes PCM data in 64KB chunks
5. The server drains the buffer; no explicit close needed

This allows overlapping sounds — multiple streams play concurrently.

## Usage Example: Procedural Sound Effects

```go
package main

import (
    "bytes"
    "math"

    "github.com/AchrafSoltani/glow"
)

const sampleRate = 44100

func generateBeep(freq float64, duration float64) []byte {
    samples := int(float64(sampleRate) * duration)
    buf := make([]byte, samples*2) // 16-bit mono

    for i := 0; i < samples; i++ {
        t := float64(i) / float64(sampleRate)
        progress := float64(i) / float64(samples)

        val := math.Sin(2 * math.Pi * freq * t)
        env := 1.0 - progress // linear fade
        sample := int16(val * env * 8000)

        buf[i*2] = byte(sample)
        buf[i*2+1] = byte(sample >> 8)
    }
    return buf
}

func main() {
    ctx, err := glow.NewAudioContext(sampleRate, 1, 2)
    if err != nil {
        panic(err)
    }
    defer ctx.Close()

    beep := generateBeep(440.0, 0.3)
    player := ctx.NewPlayer(bytes.NewReader(beep))
    player.Play()

    // Keep alive while audio plays
    time.Sleep(time.Second)
}
```

## Migration from oto/v2

| oto/v2 | Glow Audio |
|---|---|
| `oto.NewContext(rate, ch, depth)` returns `(ctx, ready, err)` | `glow.NewAudioContext(rate, ch, depth)` returns `(ctx, err)` |
| `<-ready` (async init) | (synchronous — no channel wait) |
| `ctx.NewPlayer(reader)` | `ctx.NewPlayer(reader)` |
| `player.Play()` | `player.Play()` |
| Requires `libasound2-dev` build dep | No build dependencies |
| Uses CGo via purego | Pure Go, zero dependencies |

The API is intentionally compatible. To migrate:

1. Replace `"github.com/hajimehoshi/oto/v2"` with `"github.com/AchrafSoltani/glow"`
2. Change `oto.Context` → `glow.AudioContext`
3. Change `oto.NewContext(rate, ch, depth)` → `glow.NewAudioContext(rate, ch, depth)`
4. Remove the `<-ready` channel wait
5. Remove `libasound2-dev` from build dependencies
6. Run `go mod tidy`

## Files

| File | Package | Purpose |
|---|---|---|
| `audio.go` | `glow` | Public API (AudioContext, AudioPlayer) |
| `internal/pulse/protocol.go` | `pulse` | Constants, TagBuilder, TagParser, descriptor framing |
| `internal/pulse/auth.go` | `pulse` | Cookie reading |
| `internal/pulse/conn.go` | `pulse` | Socket connection, AUTH, SET_CLIENT_NAME |
| `internal/pulse/stream.go` | `pulse` | CREATE_PLAYBACK_STREAM, PCM data writing |

## Compatibility

- **PulseAudio**: Full support
- **PipeWire**: Full support (via PulseAudio compatibility socket)
- **ALSA-only**: Not supported (no PulseAudio socket available)
- **Wayland**: Works (audio is independent of display server)
