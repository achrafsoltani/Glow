# Extension: Cross-Platform Support

## Overview
Extend Glow to support Windows and macOS in addition to Linux/X11.

## Architecture

```
glow/
├── glow.go           # Public API (unchanged)
├── window.go         # Platform-agnostic Window type
├── internal/
│   ├── x11/          # Linux/X11 (current)
│   ├── win32/        # Windows
│   └── cocoa/        # macOS
```

## Build Tags

Use Go build tags to compile platform-specific code:

```go
// internal/platform_linux.go
//go:build linux

package internal

func NewPlatformWindow(title string, width, height int) (PlatformWindow, error) {
    return x11.NewWindow(title, width, height)
}
```

```go
// internal/platform_windows.go
//go:build windows

package internal

func NewPlatformWindow(title string, width, height int) (PlatformWindow, error) {
    return win32.NewWindow(title, width, height)
}
```

## Platform Interface

```go
// internal/platform.go
package internal

type PlatformWindow interface {
    Present(pixels []byte) error
    PollEvent() Event
    Close()
    Width() int
    Height() int
}
```

## Windows Implementation (Win32)

### Key APIs

```go
// internal/win32/window.go
//go:build windows

package win32

import (
    "syscall"
    "unsafe"
)

var (
    user32   = syscall.NewLazyDLL("user32.dll")
    gdi32    = syscall.NewLazyDLL("gdi32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")

    createWindowEx    = user32.NewProc("CreateWindowExW")
    defWindowProc     = user32.NewProc("DefWindowProcW")
    dispatchMessage   = user32.NewProc("DispatchMessageW")
    getMessage        = user32.NewProc("GetMessageW")
    peekMessage       = user32.NewProc("PeekMessageW")
    postQuitMessage   = user32.NewProc("PostQuitMessage")
    registerClass     = user32.NewProc("RegisterClassExW")
    showWindow        = user32.NewProc("ShowWindow")
    getDC             = user32.NewProc("GetDC")
    releaseDC         = user32.NewProc("ReleaseDC")

    createDIBSection  = gdi32.NewProc("CreateDIBSection")
    bitBlt            = gdi32.NewProc("BitBlt")
)
```

### Window Creation

```go
type Window struct {
    hwnd   uintptr
    hdc    uintptr
    width  int
    height int
    pixels []byte
}

func NewWindow(title string, width, height int) (*Window, error) {
    // 1. Register window class
    className := syscall.StringToUTF16Ptr("GlowWindow")
    // RegisterClassExW(...)

    // 2. Create window
    // CreateWindowExW(...)

    // 3. Show window
    // ShowWindow(hwnd, SW_SHOW)

    // 4. Create DIB for pixel buffer
    // CreateDIBSection(...)

    return &Window{...}, nil
}
```

### Message Loop

```go
func (w *Window) PollEvent() Event {
    var msg MSG

    // Non-blocking check
    ret, _, _ := peekMessage.Call(
        uintptr(unsafe.Pointer(&msg)),
        0, 0, 0,
        PM_REMOVE,
    )

    if ret == 0 {
        return nil
    }

    // Translate and dispatch
    translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
    dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))

    // Convert Windows message to Event
    switch msg.Message {
    case WM_KEYDOWN:
        return &KeyEvent{...}
    case WM_LBUTTONDOWN:
        return &ButtonEvent{...}
    // ...
    }

    return nil
}
```

### Rendering

```go
func (w *Window) Present(pixels []byte) error {
    // Copy to DIB section
    copy(w.dibPixels, pixels)

    // BitBlt to window
    bitBlt.Call(
        w.hdc,
        0, 0,
        uintptr(w.width), uintptr(w.height),
        w.memDC,
        0, 0,
        SRCCOPY,
    )

    return nil
}
```

## macOS Implementation (Cocoa)

### Using cgo with Objective-C

```go
// internal/cocoa/window.go
//go:build darwin

package cocoa

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

void* createWindow(const char* title, int width, int height);
void showWindow(void* window);
void* pollEvent(void* window);
void presentPixels(void* window, void* pixels, int width, int height);
*/
import "C"
```

### Objective-C Bridge

```objc
// internal/cocoa/window.m

#import <Cocoa/Cocoa.h>

@interface GlowWindow : NSWindow
@property (nonatomic) NSBitmapImageRep* bitmap;
@end

void* createWindow(const char* title, int width, int height) {
    NSRect frame = NSMakeRect(0, 0, width, height);
    GlowWindow* window = [[GlowWindow alloc]
        initWithContentRect:frame
        styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
        backing:NSBackingStoreBuffered
        defer:NO];

    [window setTitle:[NSString stringWithUTF8String:title]];
    [window center];

    return (__bridge_retained void*)window;
}

void presentPixels(void* win, void* pixels, int width, int height) {
    GlowWindow* window = (__bridge GlowWindow*)win;

    // Create NSBitmapImageRep from pixels
    // Draw to window's content view
}
```

## Alternative: Pure Go with x/sys

For Windows, you can avoid cgo:

```go
//go:build windows

import "golang.org/x/sys/windows"

// Use windows.NewLazySystemDLL and windows.NewProc
// All Win32 calls through syscall
```

## Cross-Platform Public API

```go
// glow.go - No changes needed!

func NewWindow(title string, width, height int) (*Window, error) {
    platform, err := internal.NewPlatformWindow(title, width, height)
    if err != nil {
        return nil, err
    }

    return &Window{
        platform: platform,
        canvas:   NewCanvas(width, height),
    }, nil
}

func (w *Window) Present() error {
    return w.platform.Present(w.canvas.Pixels())
}
```

## Files to Create

```
internal/
├── platform.go            # Interface definition
├── platform_linux.go      # Linux build
├── platform_windows.go    # Windows build
├── platform_darwin.go     # macOS build
├── x11/                   # Existing code
├── win32/
│   └── window.go          # Windows implementation
└── cocoa/
    ├── window.go          # Go code
    └── window.m           # Objective-C code
```

## Building

```bash
# Linux (default)
go build

# Windows cross-compile
GOOS=windows GOARCH=amd64 go build

# macOS (requires macOS or cross-compiler)
GOOS=darwin GOARCH=amd64 go build
```

## Resources

- Win32 Programming: https://docs.microsoft.com/en-us/windows/win32/
- Cocoa Programming: https://developer.apple.com/documentation/appkit
- go-gl/glfw: https://github.com/go-gl/glfw (reference)
