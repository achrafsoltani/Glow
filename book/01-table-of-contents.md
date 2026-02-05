# Table of Contents

## Part I: Foundation

### Chapter 1: Introduction
- 1.1 What We're Building
- 1.2 Why Pure Go?
- 1.3 Architecture Overview
- 1.4 Setting Up the Project

### Chapter 2: Go for Systems Programming
- 2.1 Binary Data and Byte Slices
- 2.2 The encoding/binary Package
- 2.3 Endianness Explained
- 2.4 Working with Buffers
- 2.5 Network Sockets in Go

## Part II: The X11 Protocol

### Chapter 3: Understanding X11
- 3.1 A Brief History
- 3.2 The Client-Server Model
- 3.3 Protocol Structure
- 3.4 Requests, Replies, Events, and Errors
- 3.5 Resource IDs

### Chapter 4: Connecting to the X Server
- 4.1 Finding the Display
- 4.2 Unix Domain Sockets
- 4.3 The Connection Handshake
- 4.4 Parsing the Setup Response
- 4.5 Extracting Screen Information

### Chapter 5: Authentication
- 5.1 Why Authentication?
- 5.2 The Xauthority File
- 5.3 MIT-MAGIC-COOKIE-1
- 5.4 Parsing Xauthority Entries
- 5.5 Sending Credentials

## Part III: Window Management

### Chapter 6: Creating Windows
- 6.1 The CreateWindow Request
- 6.2 Window Attributes and Masks
- 6.3 Mapping and Unmapping
- 6.4 Window Hierarchy
- 6.5 Destroying Windows

### Chapter 7: Window Properties
- 7.1 Understanding Atoms
- 7.2 The InternAtom Request
- 7.3 Setting the Window Title
- 7.4 The Close Button Protocol
- 7.5 EWMH and ICCCM Basics

## Part IV: Event Handling

### Chapter 8: The Event System
- 8.1 Event Types Overview
- 8.2 Event Masks
- 8.3 Reading Events from the Socket
- 8.4 Parsing Event Data
- 8.5 Keyboard Events
- 8.6 Mouse Events
- 8.7 Window Events

### Chapter 9: Non-Blocking Events
- 9.1 The Problem with Blocking
- 9.2 Goroutines and Channels
- 9.3 Building an Event Queue
- 9.4 Polling vs Waiting
- 9.5 Thread Safety Considerations

## Part V: Graphics

### Chapter 10: The Graphics Context
- 10.1 What is a GC?
- 10.2 Creating a Graphics Context
- 10.3 GC Attributes
- 10.4 Foreground and Background Colors
- 10.5 Freeing Resources

### Chapter 11: Rendering with PutImage
- 11.1 Image Formats in X11
- 11.2 ZPixmap Explained
- 11.3 The PutImage Request
- 11.4 Pixel Byte Order (BGRA)
- 11.5 The Request Size Limit
- 11.6 Splitting Large Images

### Chapter 12: The Framebuffer
- 12.1 Software Rendering Basics
- 12.2 Designing the Framebuffer
- 12.3 Setting Pixels
- 12.4 Clearing the Screen
- 12.5 Coordinate Systems

### Chapter 13: Drawing Primitives
- 13.1 Lines with Bresenham's Algorithm
- 13.2 Rectangles (Filled and Outline)
- 13.3 Circles with the Midpoint Algorithm
- 13.4 Triangles
- 13.5 Clipping

## Part VI: API Design

### Chapter 14: The Public Interface
- 14.1 Design Principles
- 14.2 The Window Type
- 14.3 The Canvas Type
- 14.4 Color Handling
- 14.5 Error Handling

### Chapter 15: Event Abstraction
- 15.1 Defining Event Types
- 15.2 Key Codes and Mapping
- 15.3 Mouse Buttons
- 15.4 The PollEvent Pattern
- 15.5 Quit Events

## Part VII: Projects

### Chapter 16: Pong
- 16.1 Game Architecture
- 16.2 The Game Loop
- 16.3 Paddle and Ball Physics
- 16.4 Collision Detection
- 16.5 Scoring and Display
- 16.6 Input Handling

### Chapter 17: Paint
- 17.1 Application Design
- 17.2 Tool System
- 17.3 Brush and Eraser
- 17.4 Shape Tools
- 17.5 Color Palette
- 17.6 UI Elements

### Chapter 18: Particle System
- 18.1 Particle Basics
- 18.2 Physics Simulation
- 18.3 Emitter Patterns
- 18.4 Visual Effects
- 18.5 Performance Optimization

## Part VIII: Extensions

### Chapter 19: Sprites and Images
- 19.1 Image File Formats
- 19.2 Loading PNG Files
- 19.3 Sprite Rendering
- 19.4 Transparency and Alpha Blending

### Chapter 20: Text Rendering
- 20.1 Bitmap Fonts
- 20.2 TrueType Fonts
- 20.3 Font Rasterization
- 20.4 Text Layout

### Chapter 21: Audio — PulseAudio Native Protocol
- 21.1 Audio on Linux
- 21.2 PulseAudio Protocol Overview
- 21.3 Implementation
- 21.4 Public API
- 21.5 Comparing X11 and PulseAudio
- 21.6 Procedural Sound Effects
- 21.7 Summary

### Chapter 22: What's Next
- 22.1 What You've Built
- 22.2 Skills Acquired
- 22.3 Possible Extensions
- 22.4 Performance Path
- 22.5 Learning Resources
- 22.6 Similar Projects
- 22.7 Community and Contribution
- 22.8 Final Thoughts

### Chapter 23: Gamepad Support (Planned)
- 22.1 Linux Joystick API
- 22.2 Reading Input Events
- 22.3 Controller Mapping
- 22.4 Dead Zones and Calibration

### Chapter 23: Performance Optimization
- 23.1 Profiling Go Code
- 23.2 MIT-SHM Extension
- 23.3 Shared Memory Rendering
- 23.4 Dirty Rectangles

### Chapter 24: Cross-Platform
- 24.1 Build Tags in Go
- 24.2 Platform Abstraction
- 24.3 Windows with Win32
- 24.4 macOS with Cocoa

## Appendices

### Appendix A: X11 Protocol Reference
- Request Opcodes
- Event Types
- Error Codes
- Atom Names

### Appendix B: Key Code Tables
- X11 Key Codes
- Common Keyboard Layouts

### Appendix C: Troubleshooting
- Connection Errors
- Authentication Failures
- Rendering Issues
- Performance Problems

### Appendix D: Further Reading
- Books
- Specifications
- Open Source Projects
