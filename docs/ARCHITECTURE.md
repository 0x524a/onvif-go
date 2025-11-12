# go-onvif Architecture & Design

## Overview

go-onvif is a modern, performant Go library for communicating with ONVIF-compliant IP cameras and devices. It provides a clean, type-safe API with comprehensive support for device management, media streaming, PTZ control, and imaging settings.

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
│  - onvif.Client: Main entry point                           │
│  - Context-aware operations                                  │
│  - Connection pooling                                        │
│  - Credential management                                     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                       Service Layer                          │
│  - Device Service  (device.go)                              │
│  - Media Service   (media.go)                               │
│  - PTZ Service     (ptz.go)                                 │
│  - Imaging Service (imaging.go)                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Transport Layer                         │
│  - SOAP Client (soap/soap.go)                               │
│  - WS-Security Authentication                                │
│  - XML Marshaling/Unmarshaling                              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Network Layer                           │
│  - HTTP Client with connection pooling                       │
│  - TLS support                                               │
│  - Timeout management                                        │
└─────────────────────────────────────────────────────────────┘
```

### Discovery Component

```
┌─────────────────────────────────────────────────────────────┐
│                    WS-Discovery Service                      │
│  - Multicast UDP probe                                       │
│  - Device enumeration                                        │
│  - Service endpoint discovery                                │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Context-First Design

All network operations accept `context.Context` as the first parameter, enabling:
- Request cancellation
- Timeout control
- Request tracing
- Graceful shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

info, err := client.GetDeviceInformation(ctx)
```

### 2. Functional Options Pattern

Client configuration uses functional options for flexibility:

```go
client, err := onvif.NewClient(
    endpoint,
    onvif.WithCredentials(username, password),
    onvif.WithTimeout(30*time.Second),
    onvif.WithHTTPClient(customClient),
)
```

### 3. Type Safety

Strong typing throughout the API with comprehensive struct definitions:
- Clear data structures for all ONVIF types
- Type-safe service methods
- Compile-time error detection

### 4. Error Handling

Multiple error handling strategies:
- Sentinel errors for common cases (`ErrServiceNotSupported`, `ErrAuthenticationFailed`)
- Typed `ONVIFError` for SOAP faults
- Wrapped errors with context

```go
if err := client.ContinuousMove(ctx, profileToken, velocity, nil); err != nil {
    if errors.Is(err, onvif.ErrServiceNotSupported) {
        // Handle missing PTZ support
    } else if onvif.IsONVIFError(err) {
        // Handle SOAP fault
    }
}
```

### 5. Concurrency Safety

Thread-safe operations with proper locking:
- Mutex-protected credential management
- Safe concurrent API calls
- Connection pool management

### 6. Performance Optimization

Multiple performance optimizations:
- HTTP connection pooling
- Reusable HTTP client
- Efficient XML marshaling
- Minimal allocations in hot paths

## Service Implementations

### Device Service

Provides device management functionality:
- Device information retrieval
- Capability discovery
- System operations (reboot, date/time)
- Service endpoint enumeration

### Media Service

Handles media profiles and streaming:
- Profile management
- Stream URI generation (RTSP/HTTP)
- Snapshot URI retrieval
- Encoder configuration

### PTZ Service

Controls pan-tilt-zoom operations:
- Continuous movement
- Absolute positioning
- Relative positioning
- Preset management
- Status monitoring

### Imaging Service

Manages image settings:
- Brightness, contrast, saturation
- Exposure control
- Focus management
- White balance
- Wide dynamic range (WDR)

## Security

### WS-Security Implementation

Authentication uses WS-Security UsernameToken with password digest:

1. Generate random nonce (16 bytes)
2. Get current UTC timestamp
3. Calculate digest: `Base64(SHA1(nonce + created + password))`
4. Include in SOAP header

```xml
<Security>
  <UsernameToken>
    <Username>admin</Username>
    <Password Type="...#PasswordDigest">digest</Password>
    <Nonce EncodingType="...#Base64Binary">nonce</Nonce>
    <Created>2024-01-01T12:00:00Z</Created>
  </UsernameToken>
</Security>
```

### Transport Security

- Supports HTTP and HTTPS
- Configurable TLS settings via custom HTTP client
- Certificate validation control

## Discovery Protocol

WS-Discovery implementation:

1. Send multicast probe to `239.255.255.250:3702`
2. Listen for probe matches
3. Parse device information from responses
4. Extract service endpoints (XAddrs)
5. Deduplicate devices by endpoint reference

## SOAP Message Flow

```
Client Request
     ↓
Build SOAP Envelope
     ↓
Add WS-Security Header (if authenticated)
     ↓
Marshal to XML
     ↓
HTTP POST
     ↓
Receive Response
     ↓
Parse SOAP Envelope
     ↓
Check for Fault
     ↓
Unmarshal Response Data
     ↓
Return to Caller
```

## Testing Strategy

### Unit Tests
- Client initialization and configuration
- Error handling
- Type validation
- Option application

### Integration Tests (with mock servers)
- SOAP message formatting
- Response parsing
- Error handling

### Real Device Tests
- Full service workflows
- PTZ operations
- Media streaming
- Discovery

## Performance Characteristics

### Benchmarks (typical)
- Client creation: ~100 µs
- SOAP call: ~10-50 ms (network dependent)
- Discovery: ~1-5 seconds
- Memory usage: ~1-5 MB per client

### Scalability
- Supports hundreds of concurrent clients
- Connection pooling reduces overhead
- Minimal memory footprint per device

## Future Enhancements

### Planned Features
- Event service (event subscription, pull-point)
- Analytics service (rule engine, motion detection)
- Recording service (recording management)
- Replay service (playback control)
- Advanced security (X.509 certificates)

### Optimizations
- Response caching for static data
- Batch operations support
- Streaming data handling
- WebSocket support for events

## Best Practices

### Client Lifecycle
```go
// Create client once
client, err := onvif.NewClient(endpoint, options...)
if err != nil {
    return err
}

// Initialize to discover services
if err := client.Initialize(ctx); err != nil {
    return err
}

// Reuse client for multiple operations
// ...

// No explicit cleanup needed (HTTP client manages connections)
```

### Error Handling
```go
info, err := client.GetDeviceInformation(ctx)
if err != nil {
    // Check for specific errors
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
    return fmt.Errorf("failed to get device info: %w", err)
}
```

### Resource Management
```go
// Use contexts with timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Operations automatically respect context cancellation
result, err := client.Operation(ctx, ...)
```

## Dependencies

Minimal external dependencies:
- `golang.org/x/net`: HTTP/2 support and IDNA
- `golang.org/x/text`: Character encoding
- Go standard library: Everything else

## Compliance

- **ONVIF Core Specification**: ✓
- **ONVIF Profile S** (Streaming): ✓
- **ONVIF Profile T** (Advanced Streaming): Partial
- **ONVIF Profile G** (Recording): Planned
- **WS-Security**: ✓ (UsernameToken)
- **WS-Discovery**: ✓

## Conclusion

go-onvif provides a modern, performant, and easy-to-use Go library for ONVIF camera integration. Its architecture prioritizes:
- Developer experience (simple, intuitive API)
- Type safety (compile-time error detection)
- Performance (connection pooling, efficient operations)
- Reliability (comprehensive error handling)
- Standards compliance (ONVIF specifications)
