# Step 1: Project Setup

## Goal
Set up the Go module and directory structure for the Glow library.

## What You'll Learn
- Go module initialization
- Project organization best practices
- Internal packages in Go

## Steps

### 1.1 Create the Project Directory

```bash
mkdir -p ~/Projects/glow
cd ~/Projects/glow
```

### 1.2 Initialize Go Module

```bash
go mod init github.com/yourusername/glow
```

This creates `go.mod`:
```go
module github.com/yourusername/glow

go 1.21
```

### 1.3 Create Directory Structure

```bash
mkdir -p internal/x11
mkdir -p examples
```

**Why `internal/`?**
Go's `internal` directory is special - packages inside cannot be imported by external code. This lets us hide implementation details while exposing a clean public API.

### 1.4 Understanding the Structure

```
glow/
├── go.mod              # Module definition
├── glow.go             # Public API (will create later)
├── internal/           # Private implementation
│   └── x11/            # X11-specific code
│       ├── conn.go     # Connection handling
│       ├── window.go   # Window management
│       └── ...
└── examples/           # Example programs
    └── ...
```

## Key Concepts

### Go Modules
- `go.mod` defines your module path and dependencies
- Module path is typically your repository URL
- Use `go mod tidy` to clean up dependencies

### Package Organization
- One package per directory
- Package name matches directory name (usually)
- `internal/` packages are private to your module

### Import Paths
```go
// External code can import:
import "github.com/yourusername/glow"

// But NOT:
import "github.com/yourusername/glow/internal/x11"  // Error!
```

## Verification

Your directory should look like:
```
$ tree
.
├── go.mod
├── internal
│   └── x11
└── examples

3 directories, 1 file
```

## Next Step
Continue to Step 2 to learn the Go fundamentals needed for graphics programming.
