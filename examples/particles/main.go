// Particles - Physics-based particle system built with Glow
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/AchrafSoltani/glow"
)

const (
	screenWidth  = 1000
	screenHeight = 700
	maxParticles = 5000
	gravity      = 0.2
)

type Particle struct {
	X, Y     float64
	VX, VY   float64
	Life     float64
	MaxLife  float64
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
	particles    []Particle
	emitterType  EmitterType
	emitterX     float64
	emitterY     float64
	emitRate     int
	frame        int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	win, err := glow.NewWindow("Glow Particles", screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer win.Close()

	fmt.Println("=== GLOW PARTICLES ===")
	fmt.Println("Click: Spawn explosion")
	fmt.Println("1: Fountain")
	fmt.Println("2: Explosion")
	fmt.Println("3: Fire")
	fmt.Println("4: Snow")
	fmt.Println("5: Spiral")
	fmt.Println("Space: Toggle continuous emission")
	fmt.Println("ESC: Quit")

	ps := &ParticleSystem{
		particles:   make([]Particle, maxParticles),
		emitterType: EmitterFountain,
		emitterX:    screenWidth / 2,
		emitterY:    screenHeight - 50,
		emitRate:    10,
	}

	// Initialize particle pool
	for i := range ps.particles {
		ps.particles[i].Active = false
	}

	keys := make(map[glow.Key]bool)
	continuousEmit := true
	mouseX := screenWidth / 2
	_ = keys // Used for future expansion

	running := true
	for running {
		// Handle events
		for {
			event := win.PollEvent()
			if event == nil {
				break
			}

			switch event.Type {
			case glow.EventQuit:
				running = false

			case glow.EventKeyDown:
				keys[event.Key] = true
				switch event.Key {
				case glow.KeyEscape:
					running = false
				case glow.KeySpace:
					continuousEmit = !continuousEmit
					fmt.Printf("Continuous emission: %v\n", continuousEmit)
				case glow.Key1:
					ps.emitterType = EmitterFountain
					ps.emitterY = screenHeight - 50
					fmt.Println("Emitter: Fountain")
				case glow.Key2:
					ps.emitterType = EmitterExplosion
					fmt.Println("Emitter: Explosion")
				case glow.Key3:
					ps.emitterType = EmitterFire
					ps.emitterY = screenHeight - 50
					fmt.Println("Emitter: Fire")
				case glow.Key4:
					ps.emitterType = EmitterSnow
					ps.emitterY = 10
					fmt.Println("Emitter: Snow")
				case glow.Key5:
					ps.emitterType = EmitterSpiral
					ps.emitterY = screenHeight / 2
					fmt.Println("Emitter: Spiral")
				}

			case glow.EventKeyUp:
				keys[event.Key] = false

			case glow.EventMouseMotion:
				mouseX = event.X

			case glow.EventMouseButtonDown:
				if event.Button == glow.MouseLeft {
					// Spawn explosion at click
					spawnExplosion(ps, float64(event.X), float64(event.Y), 100)
				}
			}
		}

		// Update emitter position based on mode
		if ps.emitterType == EmitterSpiral {
			t := float64(ps.frame) * 0.05
			ps.emitterX = float64(screenWidth)/2 + 150*math.Cos(t)
			ps.emitterY = float64(screenHeight)/2 + 150*math.Sin(t)
		} else if ps.emitterType != EmitterSnow {
			ps.emitterX = float64(mouseX)
		}

		// Emit particles
		if continuousEmit {
			switch ps.emitterType {
			case EmitterFountain:
				emitFountain(ps, ps.emitRate)
			case EmitterExplosion:
				if ps.frame%30 == 0 {
					spawnExplosion(ps, ps.emitterX, ps.emitterY, 50)
				}
			case EmitterFire:
				emitFire(ps, ps.emitRate*2)
			case EmitterSnow:
				emitSnow(ps, ps.emitRate/2)
			case EmitterSpiral:
				emitSpiral(ps, ps.emitRate)
			}
		}

		// Update particles
		updateParticles(ps)

		// Draw
		canvas := win.Canvas()
		canvas.Clear(glow.RGB(10, 10, 20))

		// Draw particles
		drawParticles(canvas, ps)

		// Draw emitter indicator
		if continuousEmit && ps.emitterType != EmitterSnow {
			canvas.DrawCircle(int(ps.emitterX), int(ps.emitterY), 5, glow.RGB(100, 100, 100))
		}

		// Draw stats
		activeCount := 0
		for i := range ps.particles {
			if ps.particles[i].Active {
				activeCount++
			}
		}
		drawStats(canvas, activeCount, ps.emitterType)

		win.Present()
		ps.frame++
		time.Sleep(16 * time.Millisecond)
	}
}

func findInactiveParticle(ps *ParticleSystem) *Particle {
	for i := range ps.particles {
		if !ps.particles[i].Active {
			return &ps.particles[i]
		}
	}
	return nil
}

func emitFountain(ps *ParticleSystem, count int) {
	for i := 0; i < count; i++ {
		p := findInactiveParticle(ps)
		if p == nil {
			return
		}

		angle := -math.Pi/2 + (rand.Float64()-0.5)*0.5
		speed := 5 + rand.Float64()*5

		p.X = ps.emitterX + (rand.Float64()-0.5)*10
		p.Y = ps.emitterY
		p.VX = speed * math.Cos(angle)
		p.VY = speed * math.Sin(angle)
		p.Life = 100 + rand.Float64()*50
		p.MaxLife = p.Life
		p.Size = 2 + rand.Float64()*3
		// Blue to cyan gradient
		p.R = uint8(50 + rand.Intn(50))
		p.G = uint8(150 + rand.Intn(100))
		p.B = uint8(200 + rand.Intn(55))
		p.Active = true
	}
}

func spawnExplosion(ps *ParticleSystem, x, y float64, count int) {
	for i := 0; i < count; i++ {
		p := findInactiveParticle(ps)
		if p == nil {
			return
		}

		angle := rand.Float64() * 2 * math.Pi
		speed := 2 + rand.Float64()*8

		p.X = x
		p.Y = y
		p.VX = speed * math.Cos(angle)
		p.VY = speed * math.Sin(angle)
		p.Life = 50 + rand.Float64()*50
		p.MaxLife = p.Life
		p.Size = 2 + rand.Float64()*4
		// Orange/yellow explosion
		p.R = uint8(200 + rand.Intn(55))
		p.G = uint8(100 + rand.Intn(100))
		p.B = uint8(rand.Intn(50))
		p.Active = true
	}
}

func emitFire(ps *ParticleSystem, count int) {
	for i := 0; i < count; i++ {
		p := findInactiveParticle(ps)
		if p == nil {
			return
		}

		p.X = ps.emitterX + (rand.Float64()-0.5)*40
		p.Y = ps.emitterY
		p.VX = (rand.Float64() - 0.5) * 2
		p.VY = -2 - rand.Float64()*3
		p.Life = 40 + rand.Float64()*30
		p.MaxLife = p.Life
		p.Size = 3 + rand.Float64()*5
		// Fire colors (red/orange/yellow)
		p.R = uint8(200 + rand.Intn(55))
		p.G = uint8(50 + rand.Intn(150))
		p.B = 0
		p.Active = true
	}
}

func emitSnow(ps *ParticleSystem, count int) {
	for i := 0; i < count; i++ {
		p := findInactiveParticle(ps)
		if p == nil {
			return
		}

		p.X = rand.Float64() * screenWidth
		p.Y = 0
		p.VX = (rand.Float64() - 0.5) * 1
		p.VY = 1 + rand.Float64()*2
		p.Life = 300 + rand.Float64()*100
		p.MaxLife = p.Life
		p.Size = 1 + rand.Float64()*3
		// White/light blue
		v := uint8(200 + rand.Intn(55))
		p.R = v
		p.G = v
		p.B = uint8(min(255, int(v)+20))
		p.Active = true
	}
}

func emitSpiral(ps *ParticleSystem, count int) {
	for i := 0; i < count; i++ {
		p := findInactiveParticle(ps)
		if p == nil {
			return
		}

		p.X = ps.emitterX
		p.Y = ps.emitterY
		// Emit outward from spiral center
		angle := rand.Float64() * 2 * math.Pi
		speed := 1 + rand.Float64()*2
		p.VX = speed * math.Cos(angle)
		p.VY = speed * math.Sin(angle)
		p.Life = 80 + rand.Float64()*40
		p.MaxLife = p.Life
		p.Size = 2 + rand.Float64()*2
		// Rainbow colors based on angle
		p.R = uint8(128 + 127*math.Sin(angle))
		p.G = uint8(128 + 127*math.Sin(angle+2))
		p.B = uint8(128 + 127*math.Sin(angle+4))
		p.Active = true
	}
}

func updateParticles(ps *ParticleSystem) {
	for i := range ps.particles {
		p := &ps.particles[i]
		if !p.Active {
			continue
		}

		// Apply physics
		p.X += p.VX
		p.Y += p.VY

		// Gravity (except for snow which falls slowly)
		if ps.emitterType != EmitterSnow {
			p.VY += gravity
		} else {
			// Snow sways
			p.VX += (rand.Float64() - 0.5) * 0.2
			p.VX *= 0.99 // Damping
		}

		// Age
		p.Life--

		// Die if too old or off screen
		if p.Life <= 0 || p.Y > screenHeight+10 || p.X < -10 || p.X > screenWidth+10 {
			p.Active = false
		}
	}
}

func drawParticles(canvas *glow.Canvas, ps *ParticleSystem) {
	for i := range ps.particles {
		p := &ps.particles[i]
		if !p.Active {
			continue
		}

		// Fade based on life
		lifeFactor := p.Life / p.MaxLife
		r := uint8(float64(p.R) * lifeFactor)
		g := uint8(float64(p.G) * lifeFactor)
		b := uint8(float64(p.B) * lifeFactor)

		// Size can shrink over time for some effects
		size := int(p.Size * lifeFactor)
		if size < 1 {
			size = 1
		}

		if size <= 2 {
			canvas.SetPixel(int(p.X), int(p.Y), glow.RGB(r, g, b))
			if size == 2 {
				canvas.SetPixel(int(p.X)+1, int(p.Y), glow.RGB(r, g, b))
				canvas.SetPixel(int(p.X), int(p.Y)+1, glow.RGB(r, g, b))
				canvas.SetPixel(int(p.X)+1, int(p.Y)+1, glow.RGB(r, g, b))
			}
		} else {
			canvas.FillCircle(int(p.X), int(p.Y), size, glow.RGB(r, g, b))
		}
	}
}

func drawStats(canvas *glow.Canvas, count int, emitter EmitterType) {
	// Background
	canvas.DrawRect(10, 10, 180, 50, glow.RGB(0, 0, 0))
	canvas.DrawRectOutline(10, 10, 180, 50, glow.RGB(50, 50, 50))

	// Particle count indicator
	barWidth := min(count/30, 160)
	canvas.DrawRect(20, 20, barWidth, 10, glow.RGB(100, 200, 100))

	// Emitter type indicator
	emitterNames := []string{"FOUNTAIN", "EXPLOSION", "FIRE", "SNOW", "SPIRAL"}
	name := emitterNames[emitter]
	x := 20
	for _, c := range name {
		if c != ' ' {
			canvas.DrawRect(x, 40, 4, 8, glow.White)
		}
		x += 6
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
