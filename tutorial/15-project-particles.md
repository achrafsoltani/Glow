# Step 15: Project - Particle System

## Goal
Build a physics-based particle system with multiple emitter types.

## Particle Structure

```go
const maxParticles = 5000

type Particle struct {
    X, Y     float64  // Position
    VX, VY   float64  // Velocity
    Life     float64  // Remaining life
    MaxLife  float64  // Initial life (for fading)
    Size     float64
    R, G, B  uint8
    Active   bool
}

type EmitterType int

const (
    EmitterFountain EmitterType = iota
    EmitterExplosion
    EmitterFire
    EmitterSnow
    EmitterSpiral
)

type ParticleSystem struct {
    particles   []Particle
    emitterType EmitterType
    emitterX    float64
    emitterY    float64
}
```

## Particle Pool

Use a fixed array and reuse inactive particles:

```go
func NewParticleSystem() *ParticleSystem {
    return &ParticleSystem{
        particles: make([]Particle, maxParticles),
    }
}

func (ps *ParticleSystem) findInactive() *Particle {
    for i := range ps.particles {
        if !ps.particles[i].Active {
            return &ps.particles[i]
        }
    }
    return nil  // All particles in use
}
```

## Emitter Types

### Fountain
```go
func (ps *ParticleSystem) emitFountain(count int) {
    for i := 0; i < count; i++ {
        p := ps.findInactive()
        if p == nil { return }

        angle := -math.Pi/2 + (rand.Float64()-0.5)*0.5  // Upward spread
        speed := 5 + rand.Float64()*5

        p.X = ps.emitterX
        p.Y = ps.emitterY
        p.VX = speed * math.Cos(angle)
        p.VY = speed * math.Sin(angle)
        p.Life = 100 + rand.Float64()*50
        p.MaxLife = p.Life
        p.Size = 2 + rand.Float64()*3
        p.R, p.G, p.B = 50, 150, 255  // Blue
        p.Active = true
    }
}
```

### Explosion
```go
func (ps *ParticleSystem) emitExplosion(x, y float64, count int) {
    for i := 0; i < count; i++ {
        p := ps.findInactive()
        if p == nil { return }

        angle := rand.Float64() * 2 * math.Pi  // All directions
        speed := 2 + rand.Float64()*8

        p.X, p.Y = x, y
        p.VX = speed * math.Cos(angle)
        p.VY = speed * math.Sin(angle)
        p.Life = 50 + rand.Float64()*50
        p.MaxLife = p.Life
        p.Size = 2 + rand.Float64()*4
        p.R, p.G, p.B = 255, uint8(100+rand.Intn(100)), 0  // Orange
        p.Active = true
    }
}
```

### Fire
```go
func (ps *ParticleSystem) emitFire(count int) {
    for i := 0; i < count; i++ {
        p := ps.findInactive()
        if p == nil { return }

        p.X = ps.emitterX + (rand.Float64()-0.5)*40
        p.Y = ps.emitterY
        p.VX = (rand.Float64() - 0.5) * 2
        p.VY = -2 - rand.Float64()*3  // Upward
        p.Life = 40 + rand.Float64()*30
        p.MaxLife = p.Life
        p.Size = 3 + rand.Float64()*5
        p.R = 255
        p.G = uint8(50 + rand.Intn(150))
        p.B = 0
        p.Active = true
    }
}
```

## Physics Update

```go
const gravity = 0.2

func (ps *ParticleSystem) update() {
    for i := range ps.particles {
        p := &ps.particles[i]
        if !p.Active { continue }

        // Movement
        p.X += p.VX
        p.Y += p.VY

        // Gravity (except snow)
        if ps.emitterType != EmitterSnow {
            p.VY += gravity
        } else {
            // Snow sways
            p.VX += (rand.Float64() - 0.5) * 0.2
            p.VX *= 0.99
        }

        // Age
        p.Life--

        // Death
        if p.Life <= 0 || p.Y > screenHeight+10 {
            p.Active = false
        }
    }
}
```

## Rendering with Fading

```go
func (ps *ParticleSystem) draw(canvas *glow.Canvas) {
    for i := range ps.particles {
        p := &ps.particles[i]
        if !p.Active { continue }

        // Fade based on remaining life
        factor := p.Life / p.MaxLife
        r := uint8(float64(p.R) * factor)
        g := uint8(float64(p.G) * factor)
        b := uint8(float64(p.B) * factor)

        // Size can also shrink
        size := int(p.Size * factor)
        if size < 1 { size = 1 }

        if size <= 2 {
            canvas.SetPixel(int(p.X), int(p.Y), glow.RGB(r, g, b))
        } else {
            canvas.FillCircle(int(p.X), int(p.Y), size, glow.RGB(r, g, b))
        }
    }
}
```

## Main Loop

```go
func main() {
    win, _ := glow.NewWindow("Particles", 1000, 700)
    defer win.Close()

    ps := NewParticleSystem()
    ps.emitterX = 500
    ps.emitterY = 650

    for running {
        // Events
        for event := win.PollEvent(); event != nil; event = win.PollEvent() {
            if event.Type == glow.EventMouseButtonDown {
                ps.emitExplosion(float64(event.X), float64(event.Y), 100)
            }
        }

        // Emit continuously
        ps.emitFountain(10)

        // Update
        ps.update()

        // Draw
        canvas := win.Canvas()
        canvas.Clear(glow.RGB(10, 10, 20))
        ps.draw(canvas)
        win.Present()

        time.Sleep(16 * time.Millisecond)
    }
}
```

## Learning Outcomes

- Object pooling
- Physics simulation
- Visual effects (fading, color)
- Performance with many objects

## Run It

```bash
go run examples/particles/main.go
```

Controls:
- Click: Spawn explosion
- 1-5: Change emitter type
- Space: Toggle emission
- ESC: Quit
