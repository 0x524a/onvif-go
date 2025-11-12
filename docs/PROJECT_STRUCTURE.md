# Project Structure

## Overview

The `onvif-go` project follows the **Standard Go Project Layout** optimized for a library package. This structure provides clear separation between public APIs, private implementation details, executable commands, and supporting resources.

## Directory Layout

```
onvif-go/
├── *.go                    # Public API files (root level)
│   ├── client.go          # Main ONVIF client
│   ├── device.go          # Device service operations
│   ├── media.go           # Media service operations
│   ├── ptz.go             # PTZ service operations
│   ├── imaging.go         # Imaging service operations
│   ├── types.go           # Public type definitions
│   ├── errors.go          # Error types and handling
│   └── doc.go             # Package documentation
│
├── internal/              # Private packages (not importable externally)
│   └── soap/             # SOAP client implementation
│       ├── soap.go       # SOAP envelope building and parsing
│       └── soap_test.go  # SOAP client tests
│
├── discovery/            # Device discovery subpackage (public)
│   ├── discovery.go      # WS-Discovery implementation
│   └── discovery_test.go # Discovery tests
│
├── server/              # ONVIF server implementation (public)
│   ├── server.go        # Main server
│   ├── device.go        # Device service handlers
│   ├── media.go         # Media service handlers
│   ├── ptz.go           # PTZ service handlers
│   ├── imaging.go       # Imaging service handlers
│   └── soap/            # Server SOAP handling
│       └── handler.go   # SOAP request handler
│
├── cmd/                 # Command-line applications
│   ├── onvif-cli/       # Interactive CLI tool
│   ├── onvif-quick/     # Quick test utility
│   ├── onvif-server/    # Virtual camera server
│   ├── onvif-diagnostics/ # Diagnostic tool
│   └── generate-tests/  # Test generation utility
│
├── examples/            # Example applications
│   ├── device-info/     # Get device information
│   ├── discovery/       # Discover cameras
│   ├── ptz-control/     # PTZ operations
│   ├── imaging-settings/ # Imaging configuration
│   ├── complete-demo/   # Full feature demo
│   ├── simplified-endpoint/ # Endpoint format demo
│   └── .../            # Additional examples
│
├── docs/               # Documentation
│   ├── ARCHITECTURE.md # Architecture overview
│   ├── PROJECT_STRUCTURE.md # This file
│   ├── SIMPLIFIED_ENDPOINT.md # Endpoint API docs
│   └── .../           # Additional documentation
│
├── test/              # Additional test utilities
├── testdata/          # Test fixtures and data
├── testing/           # Testing helpers
│
├── .github/           # GitHub workflows and configs
│   └── workflows/
│       └── release.yml # Release automation
│
├── go.mod             # Go module definition
├── go.sum             # Dependency checksums
├── Makefile           # Build automation
├── Dockerfile         # Container image
├── README.md          # Project readme
├── CHANGELOG.md       # Version history
├── LICENSE            # License information
├── CONTRIBUTING.md    # Contribution guidelines
├── QUICKSTART.md      # Quick start guide
└── BUILDING.md        # Build instructions
```

## Design Principles

### 1. Library-First Design

As a **library package**, the main API lives at the root level:

```go
import "github.com/0x524A/onvif-go"

client, err := onvif.NewClient("192.168.1.100")
```

**Benefits:**
- Clean, simple import path
- Follows Go conventions for libraries
- Easy to discover and use
- No unnecessary nesting

### 2. Internal Package for Private Code

The `internal/` directory contains implementation details not intended for external use:

```go
// This import is ONLY available within onvif-go:
import "github.com/0x524A/onvif-go/internal/soap"
```

**Go's internal package restriction:**
- Cannot be imported by external projects
- Enforced by the Go compiler
- Allows refactoring without breaking changes

**What goes in internal/**:
- SOAP client implementation
- Protocol-specific details
- Helper functions not part of public API
- Implementation details that might change

### 3. Subpackages for Additional Features

Public subpackages for optional or specialized functionality:

```go
// Discovery subpackage
import "github.com/0x524A/onvif-go/discovery"

// Server subpackage
import "github.com/0x524A/onvif-go/server"
```

**When to create a subpackage:**
- Logically separate feature set
- Can be used independently
- Different import namespace makes sense
- Clear, single responsibility

### 4. Commands in cmd/

Executable applications in `cmd/` directory:

```
cmd/
├── onvif-cli/       # Main CLI tool
├── onvif-server/    # Virtual camera
└── onvif-quick/     # Quick utility
```

**Naming convention:**
- Directory name = binary name
- Each cmd has its own `main.go`
- Can import the library: `import "github.com/0x524A/onvif-go"`

**Build commands:**
```bash
go build ./cmd/onvif-cli
go build ./cmd/onvif-server
```

### 5. Examples for Documentation

The `examples/` directory demonstrates library usage:

**Structure:**
- Each example is a standalone program
- Clear, focused demonstration
- Can be built and run directly

**Purpose:**
- Supplement documentation
- Show best practices
- Provide starting points for users

### 6. Documentation in docs/

Comprehensive documentation in `docs/` directory:

- `ARCHITECTURE.md` - Design and architecture
- `PROJECT_STRUCTURE.md` - This file
- `SIMPLIFIED_ENDPOINT.md` - Feature documentation
- Additional guides as needed

**Why separate docs/?**
- Keeps root clean
- Organized by topic
- Easy to navigate
- Scalable structure

## Import Patterns

### Public API (Root Package)

```go
// Main client functionality
import "github.com/0x524A/onvif-go"

client, err := onvif.NewClient("192.168.1.100",
    onvif.WithCredentials("admin", "password"),
)
```

### Discovery Subpackage

```go
// Device discovery
import "github.com/0x524A/onvif-go/discovery"

devices, err := discovery.Discover(ctx, 5*time.Second)
```

### Server Subpackage

```go
// Virtual ONVIF server
import "github.com/0x524A/onvif-go/server"

srv := server.NewServer(
    server.WithCredentials("admin", "admin"),
    server.WithAddress(":8080"),
)
```

### Internal Package (Library Use Only)

```go
// Only usable within onvif-go itself
import "github.com/0x524A/onvif-go/internal/soap"

// External projects CANNOT import internal packages
```

## File Organization Best Practices

### Root Package Files

Group by service/functionality:
- `client.go` - Client creation and core functionality
- `device.go` - Device service methods
- `media.go` - Media service methods
- `ptz.go` - PTZ service methods
- `imaging.go` - Imaging service methods
- `types.go` - Type definitions
- `errors.go` - Error types
- `doc.go` - Package documentation

### Test Files

Co-located with source:
- `client_test.go` - Tests for client.go
- `device_test.go` - Tests for device.go
- Mirrors source file structure

### Large Packages

For large packages, consider grouping:
```
server/
├── server.go          # Main server
├── device.go          # Device handlers
├── media.go           # Media handlers
├── ptz.go             # PTZ handlers
├── imaging.go         # Imaging handlers
└── soap/              # SOAP sub-package
    └── handler.go
```

## Comparison with Other Layouts

### ❌ Avoid: pkg/ Directory for Libraries

```
# DON'T DO THIS for libraries:
my-lib/
└── pkg/
    └── mylib/
        └── mylib.go

# Requires: import "github.com/user/my-lib/pkg/mylib"
```

**Why not?**
- Unnecessary nesting
- More complex imports
- Not idiomatic for Go libraries
- `pkg/` is for applications with multiple packages

### ✅ Library Layout (What We Use)

```
onvif-go/
├── *.go              # Public API at root
└── internal/         # Private implementation

# Clean import: import "github.com/user/onvif-go"
```

### 📦 Application Layout (Different Use Case)

For applications (not libraries):
```
my-app/
├── cmd/             # Multiple binaries
├── internal/        # Private app code
├── pkg/             # Exported libraries from this app
└── main.go          # Or in cmd/
```

## Migration Notes

### Recent Changes

**Moved SOAP to internal/:**
- `soap/` → `internal/soap/`
- Updated imports in:
  - `device.go`
  - `media.go`
  - `ptz.go`
  - `imaging.go`
  - `server/soap/handler.go`

**Reason:**
- SOAP client is an implementation detail
- Users should interact through high-level API
- Prevents tight coupling to SOAP specifics
- Allows future protocol changes

### Import Updates

**Old:**
```go
import "github.com/0x524A/onvif-go/soap"
```

**New:**
```go
import "github.com/0x524A/onvif-go/internal/soap"
```

**External users:** No changes needed (they never imported soap directly)

## Benefits of This Structure

### For Library Users

1. **Simple imports**: `import "github.com/0x524A/onvif-go"`
2. **Clear API**: Public vs private clearly separated
3. **Stable interface**: Internal changes don't affect users
4. **Good documentation**: Examples and docs organized

### For Contributors

1. **Clear organization**: Each file has single responsibility
2. **Easy navigation**: Logical directory structure
3. **Safe refactoring**: Internal package allows changes
4. **Standard layout**: Follows Go conventions

### For Maintenance

1. **Backward compatibility**: Internal changes don't break users
2. **Scalability**: Structure supports growth
3. **Testing**: Co-located tests, separate test utilities
4. **Documentation**: Organized in docs/

## Future Considerations

As the project grows:

1. **More subpackages**: Analytics, events, recording services
2. **Additional internal packages**: Caching, connection pooling
3. **Tool improvements**: Enhanced cmd/ utilities
4. **Documentation growth**: More guides in docs/

The current structure supports these additions naturally.

## References

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Blog: Package names](https://go.dev/blog/package-names)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Summary

The onvif-go project structure:
- ✅ Follows Go conventions for libraries
- ✅ Public API at root level
- ✅ Internal package for private code
- ✅ Subpackages for additional features
- ✅ Clear separation of concerns
- ✅ Scalable and maintainable
- ✅ User-friendly imports
