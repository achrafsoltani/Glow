# Extension: Audio System

## Overview
Add sound effects and music playback.

## Approaches

1. **ALSA** - Linux native audio (complex)
2. **PulseAudio** - Modern Linux audio (via socket)
3. **oto** - Cross-platform Go library (recommended)
4. **beep** - Simple Go audio library

## Recommended: Using oto Library

```bash
go get github.com/hajimehoshi/oto/v2
```

### Basic Audio System

```go
import (
    "github.com/hajimehoshi/oto/v2"
)

type AudioSystem struct {
    context *oto.Context
    ready   bool
}

func NewAudioSystem() (*AudioSystem, error) {
    // Sample rate: 44100 Hz
    // Channels: 2 (stereo)
    // Bit depth: 16
    ctx, ready, err := oto.NewContext(44100, 2, 2)
    if err != nil {
        return nil, err
    }

    <-ready  // Wait for audio to be ready

    return &AudioSystem{
        context: ctx,
        ready:   true,
    }, nil
}
```

### Loading WAV Files

```go
type Sound struct {
    data       []byte
    sampleRate int
    channels   int
}

func LoadWAV(filename string) (*Sound, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // WAV header parsing
    // RIFF header
    if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
        return nil, errors.New("not a WAV file")
    }

    // fmt chunk
    channels := int(binary.LittleEndian.Uint16(data[22:24]))
    sampleRate := int(binary.LittleEndian.Uint32(data[24:28]))

    // Find data chunk
    offset := 12
    for offset < len(data)-8 {
        chunkID := string(data[offset : offset+4])
        chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))

        if chunkID == "data" {
            return &Sound{
                data:       data[offset+8 : offset+8+chunkSize],
                sampleRate: sampleRate,
                channels:   channels,
            }, nil
        }

        offset += 8 + chunkSize
    }

    return nil, errors.New("no data chunk found")
}
```

### Playing Sounds

```go
func (a *AudioSystem) Play(sound *Sound) {
    player := a.context.NewPlayer(bytes.NewReader(sound.data))
    player.Play()
    // Note: player runs in background
}

// With volume control
type SoundPlayer struct {
    player oto.Player
    volume float64
}

func (a *AudioSystem) PlayWithVolume(sound *Sound, volume float64) *SoundPlayer {
    // Create volume-adjusted data
    adjusted := adjustVolume(sound.data, volume)
    player := a.context.NewPlayer(bytes.NewReader(adjusted))
    player.Play()
    return &SoundPlayer{player: player, volume: volume}
}

func adjustVolume(data []byte, volume float64) []byte {
    result := make([]byte, len(data))
    for i := 0; i < len(data); i += 2 {
        sample := int16(binary.LittleEndian.Uint16(data[i:]))
        sample = int16(float64(sample) * volume)
        binary.LittleEndian.PutUint16(result[i:], uint16(sample))
    }
    return result
}
```

### Background Music

```go
type MusicPlayer struct {
    audio    *AudioSystem
    playing  bool
    loop     bool
    current  *Sound
    stopChan chan struct{}
}

func (m *MusicPlayer) Play(sound *Sound, loop bool) {
    m.Stop()  // Stop current

    m.current = sound
    m.loop = loop
    m.playing = true
    m.stopChan = make(chan struct{})

    go func() {
        for m.playing {
            player := m.audio.context.NewPlayer(bytes.NewReader(sound.data))
            player.Play()

            // Wait for completion or stop
            for player.IsPlaying() {
                select {
                case <-m.stopChan:
                    player.Close()
                    return
                default:
                    time.Sleep(10 * time.Millisecond)
                }
            }

            if !m.loop {
                break
            }
        }
    }()
}

func (m *MusicPlayer) Stop() {
    if m.playing {
        m.playing = false
        close(m.stopChan)
    }
}
```

### Usage Example

```go
func main() {
    audio, _ := glow.NewAudioSystem()

    // Load sounds
    jumpSound, _ := glow.LoadWAV("jump.wav")
    coinSound, _ := glow.LoadWAV("coin.wav")
    music, _ := glow.LoadWAV("music.wav")

    // Background music
    audio.PlayMusic(music, true)  // Loop

    // In game loop:
    if playerJumped {
        audio.Play(jumpSound)
    }
    if collectedCoin {
        audio.Play(coinSound)
    }
}
```

## Alternative: Pure ALSA (Advanced)

For direct ALSA access without external libraries:

```go
import "golang.org/x/sys/unix"

func openALSA() (*os.File, error) {
    // Open ALSA device
    return os.OpenFile("/dev/snd/pcmC0D0p", os.O_WRONLY, 0)
}

// Requires ioctl calls to configure
// Very complex - use oto instead
```

## Files to Create

1. `audio.go` - AudioSystem, Sound types
2. `wav.go` - WAV loading
3. `music.go` - Background music player

## Dependencies

```bash
# For oto on Linux
sudo apt install libasound2-dev
```

## Resources

- oto library: https://github.com/hajimehoshi/oto
- WAV format: https://en.wikipedia.org/wiki/WAV
- beep library: https://github.com/faiface/beep
