# Step 9: Non-Blocking Events with Goroutines

## Goal
Implement non-blocking event polling using Go's concurrency features.

## What You'll Learn
- Goroutines for background tasks
- Channels for communication
- Select for non-blocking operations

## 9.1 The Problem

`NextEvent()` blocks until an event arrives. This is bad for games:

```go
// BAD: Blocks entire game loop waiting for input
for {
    event := conn.NextEvent()  // Blocks here!
    handleEvent(event)
    updateGame()
    render()
}
```

## 9.2 Solution: Event Channel

Use a goroutine to read events into a channel:

```go
type Window struct {
    conn      *Connection
    windowID  uint32
    eventChan chan Event
    quitChan  chan struct{}
}

func NewWindow() *Window {
    w := &Window{
        eventChan: make(chan Event, 256),  // Buffered channel
        quitChan:  make(chan struct{}),
    }

    // Start event reading goroutine
    go w.pollEvents()

    return w
}
```

## 9.3 Event Polling Goroutine

```go
func (w *Window) pollEvents() {
    for {
        select {
        case <-w.quitChan:
            return  // Stop when window closes
        default:
            event, err := w.conn.NextEvent()
            if err != nil {
                continue
            }

            if event != nil {
                select {
                case w.eventChan <- event:
                    // Event sent
                default:
                    // Channel full, drop event
                }
            }
        }
    }
}
```

## 9.4 Non-Blocking Poll

```go
// PollEvent returns next event or nil if none available
func (w *Window) PollEvent() Event {
    select {
    case e := <-w.eventChan:
        return e
    default:
        return nil  // No event waiting
    }
}

// WaitEvent blocks until an event is available
func (w *Window) WaitEvent() Event {
    return <-w.eventChan
}
```

## 9.5 Clean Shutdown

```go
func (w *Window) Close() {
    close(w.quitChan)  // Signal goroutine to stop
    // ... cleanup
}
```

## 9.6 Game Loop Pattern

```go
func main() {
    win := NewWindow("Game", 800, 600)
    defer win.Close()

    running := true
    for running {
        // Process all pending events
        for {
            event := win.PollEvent()
            if event == nil {
                break  // No more events
            }

            switch e := event.(type) {
            case *QuitEvent:
                running = false
            case *KeyEvent:
                handleKey(e)
            }
        }

        // Update game state
        update()

        // Render
        render()

        // Frame timing
        time.Sleep(16 * time.Millisecond)
    }
}
```

## 9.7 Tracking Key States

For smooth movement, track which keys are held:

```go
type Input struct {
    keys map[Key]bool
    mu   sync.RWMutex
}

func (i *Input) SetKey(key Key, pressed bool) {
    i.mu.Lock()
    i.keys[key] = pressed
    i.mu.Unlock()
}

func (i *Input) IsKeyDown(key Key) bool {
    i.mu.RLock()
    defer i.mu.RUnlock()
    return i.keys[key]
}

// In event handling:
case KeyEvent:
    if e.EventType == EventKeyPress {
        input.SetKey(Key(e.Keycode), true)
    } else {
        input.SetKey(Key(e.Keycode), false)
    }

// In update:
if input.IsKeyDown(KeyW) {
    player.Y -= speed
}
```

## 9.8 Event Conversion

Convert X11 events to your public API types:

```go
func (w *Window) convertEvent(xEvent x11.Event) *Event {
    switch e := xEvent.(type) {
    case x11.KeyEvent:
        evType := EventKeyDown
        if e.EventType == x11.EventKeyRelease {
            evType = EventKeyUp
        }
        return &Event{
            Type: evType,
            Key:  Key(e.Keycode),
            X:    int(e.X),
            Y:    int(e.Y),
        }

    case x11.ButtonEvent:
        evType := EventMouseButtonDown
        if e.EventType == x11.EventButtonRelease {
            evType = EventMouseButtonUp
        }
        return &Event{
            Type:   evType,
            Button: MouseButton(e.Button),
            X:      int(e.X),
            Y:      int(e.Y),
        }

    // ... other events
    }
    return nil
}
```

## 9.9 Thread Safety Considerations

The X11 connection itself is not thread-safe. Keep all X11 operations in one goroutine:

```go
// GOOD: All X11 calls in main goroutine
go pollEvents()  // Only reads from connection

// BAD: Multiple goroutines writing to X11
go func() { conn.MapWindow(id) }()
go func() { conn.PutImage(...) }()  // Race condition!
```

## Benefits of This Approach

1. **Responsive UI**: Events processed without blocking rendering
2. **Smooth Animation**: Consistent frame timing
3. **Input Buffering**: Events queued if processing is slow
4. **Clean Separation**: Event reading separate from handling

## Next Step
Continue to Step 10 to implement graphics rendering.
