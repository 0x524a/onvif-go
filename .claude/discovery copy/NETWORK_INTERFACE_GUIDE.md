# Network Interface Discovery Guide

This guide explains how to use the network interface selection feature for ONVIF device discovery.

## Overview

When you have multiple network interfaces on your system, you may need to specify which interface to use for sending multicast discovery messages to find your cameras. This is especially important when:

- You have multiple network cards (Ethernet, WiFi, Virtual Adapters)
- Cameras are on a specific network segment
- The auto-detected interface doesn't reach your cameras
- You want to isolate discovery traffic to a specific network

## Features

✅ **Specify by Interface Name** - Use interface name (e.g., "eth0", "wlan0")  
✅ **Specify by IP Address** - Use any IP assigned to the interface  
✅ **List Available Interfaces** - See all interfaces with their configurations  
✅ **Backward Compatible** - Existing code continues to work unchanged  
✅ **Helpful Error Messages** - Lists available interfaces when one isn't found  

## Basic Usage

### 1. List Available Network Interfaces

```go
package main

import (
    "fmt"
    "log"
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    interfaces, err := discovery.ListNetworkInterfaces()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Available Network Interfaces:")
    for _, iface := range interfaces {
        fmt.Printf("  %s - Up: %v, Multicast: %v\n", iface.Name, iface.Up, iface.Multicast)
        for _, addr := range iface.Addresses {
            fmt.Printf("    IP: %s\n", addr)
        }
    }
}
```

**Output Example:**
```
Available Network Interfaces:
  lo - Up: true, Multicast: true
    IP: 127.0.0.1
    IP: ::1
  eth0 - Up: true, Multicast: true
    IP: 192.168.1.100
    IP: 169.254.1.1
  wlan0 - Up: true, Multicast: true
    IP: 192.168.88.50
  docker0 - Up: true, Multicast: true
    IP: 172.17.0.1
```

### 2. Discover Cameras on Specific Interface (by name)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    opts := &discovery.DiscoverOptions{
        NetworkInterface: "eth0",  // Discover on Ethernet
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices on eth0:\n", len(devices))
    for _, device := range devices {
        fmt.Printf("  - %s\n", device.GetDeviceEndpoint())
    }
}
```

### 3. Discover Cameras Using IP Address

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    opts := &discovery.DiscoverOptions{
        NetworkInterface: "192.168.1.100",  // Use interface with this IP
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices:\n", len(devices))
    for _, device := range devices {
        fmt.Printf("  - %s\n", device.GetDeviceEndpoint())
    }
}
```

### 4. Backward Compatible - No Changes Required

Existing code continues to work without modification:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // This still works exactly as before
    devices, err := discovery.Discover(ctx, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices\n", len(devices))
}
```

## API Reference

### DiscoverOptions

```go
type DiscoverOptions struct {
    // NetworkInterface specifies the network interface to use for multicast.
    // If empty, the system will choose the default interface.
    // Examples: "eth0", "wlan0", "192.168.1.100"
    NetworkInterface string
}
```

### Functions

#### `Discover(ctx context.Context, timeout time.Duration) ([]*Device, error)`

Discovers ONVIF devices using the default network interface (backward compatible).

**Parameters:**
- `ctx`: Context for cancellation and timeout
- `timeout`: How long to listen for responses

**Returns:**
- `[]*Device`: Discovered devices
- `error`: Any error that occurred

#### `DiscoverWithOptions(ctx context.Context, timeout time.Duration, opts *DiscoverOptions) ([]*Device, error)`

Discovers ONVIF devices with custom options including network interface selection.

**Parameters:**
- `ctx`: Context for cancellation and timeout
- `timeout`: How long to listen for responses
- `opts`: Discovery options (including NetworkInterface)

**Returns:**
- `[]*Device`: Discovered devices
- `error`: Any error that occurred

#### `ListNetworkInterfaces() ([]NetworkInterface, error)`

Lists all available network interfaces with their details.

**Returns:**
- `[]NetworkInterface`: All network interfaces
- `error`: Any error that occurred

### NetworkInterface

```go
type NetworkInterface struct {
    // Name of the interface (e.g., "eth0", "wlan0")
    Name string
    
    // IP addresses assigned to this interface
    Addresses []string
    
    // Up indicates if the interface is up
    Up bool
    
    // Multicast indicates if the interface supports multicast
    Multicast bool
}
```

## Common Scenarios

### Scenario 1: Multiple Ethernet and WiFi Interfaces

You have both Ethernet (eth0) and WiFi (wlan0), cameras are on Ethernet:

```go
// List to see what's available
interfaces, _ := discovery.ListNetworkInterfaces()
for _, i := range interfaces {
    log.Printf("%s: %v", i.Name, i.Addresses)
}

// Discover on Ethernet only
opts := &discovery.DiscoverOptions{
    NetworkInterface: "eth0",
}
devices, _ := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
```

### Scenario 2: Virtual Machine with Multiple Adapters

VM has management interface and camera network interface:

```go
// Use the camera network IP directly
opts := &discovery.DiscoverOptions{
    NetworkInterface: "192.168.200.50",  // Camera network segment
}
devices, _ := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
```

### Scenario 3: Docker Container with Custom Network

```go
// Container has multiple networks, specify which one
opts := &discovery.DiscoverOptions{
    NetworkInterface: "172.20.0.10",  // Custom bridge network IP
}
devices, _ := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
```

### Scenario 4: CLI Tool with User Selection

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    ifaceFlag := flag.String("interface", "", "Network interface to use")
    flag.Parse()

    if *ifaceFlag == "" {
        // List available if not specified
        interfaces, _ := discovery.ListNetworkInterfaces()
        fmt.Println("Available interfaces:")
        for _, i := range interfaces {
            fmt.Printf("  %s\n", i.Name)
        }
        fmt.Println("Use -interface flag to specify")
        return
    }

    opts := &discovery.DiscoverOptions{
        NetworkInterface: *ifaceFlag,
    }

    devices, _ := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
    fmt.Printf("Found %d devices\n", len(devices))
}
```

**Usage:**
```bash
# List interfaces
./app

# Available interfaces:
#   eth0
#   wlan0

# Discover on specific interface
./app -interface eth0
./app -interface wlan0
./app -interface 192.168.1.100
```

## Error Handling

### Interface Not Found

```go
opts := &discovery.DiscoverOptions{
    NetworkInterface: "nonexistent-interface",
}

devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
if err != nil {
    fmt.Println(err)
    // Output:
    // network interface "nonexistent-interface" not found. 
    // Available interfaces: [eth0 [192.168.1.100] wlan0 [192.168.88.50] ...]
}
```

### Invalid IP Address

```go
opts := &discovery.DiscoverOptions{
    NetworkInterface: "192.168.999.999",  // Invalid IP
}

devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
if err != nil {
    // Error: network interface not found
    log.Fatal(err)
}
```

## Migration Guide

### From: Using Default Discovery

```go
// Old code - still works!
devices, err := discovery.Discover(ctx, 5*time.Second)
```

### To: Using Specific Interface

```go
// New code - with interface selection
opts := &discovery.DiscoverOptions{
    NetworkInterface: "eth0",
}
devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
```

No breaking changes - old code continues to work!

## Troubleshooting

### "No devices found on interface X"

**Possible causes:**
1. Cameras are on a different network segment
2. Interface is not connected to the camera network
3. Firewall is blocking multicast on that interface
4. Camera network interface name is different than expected

**Solution:**
```go
// List interfaces to verify
interfaces, _ := discovery.ListNetworkInterfaces()
for _, i := range interfaces {
    if i.Up && i.Multicast {
        fmt.Printf("Try: %s (%v)\n", i.Name, i.Addresses)
    }
}
```

### "Network interface not found"

**Possible causes:**
1. Interface name typo (e.g., "eth0" vs "eth1")
2. Interface is down
3. IP address not assigned to any interface

**Solution:**
- Check spelling: `discovery.ListNetworkInterfaces()`
- Verify interface is up: `Up: true`
- Verify IP is correct: Check `Addresses` field

### Multicast Not Supported

```go
interfaces, _ := discovery.ListNetworkInterfaces()
for _, i := range interfaces {
    if i.Multicast {
        fmt.Printf("%s supports multicast\n", i.Name)
    }
}
```

## Best Practices

1. **Always list interfaces first** if uncertain:
   ```go
   interfaces, _ := discovery.ListNetworkInterfaces()
   // Show user and let them choose
   ```

2. **Validate interface exists** before discovery:
   ```go
   opts := &discovery.DiscoverOptions{
       NetworkInterface: userInput,
   }
   // Try with empty timeout first to validate
   ```

3. **Try multiple interfaces** for robust applications:
   ```go
   for _, iface := range interfaces {
       if iface.Up && iface.Multicast {
           opts := &discovery.DiscoverOptions{
               NetworkInterface: iface.Name,
           }
           devices, _ := discovery.DiscoverWithOptions(ctx, 2*time.Second, opts)
           if len(devices) > 0 {
               return devices
           }
       }
   }
   ```

4. **Check interface capabilities**:
   ```go
   for _, i := range interfaces {
       if i.Up && i.Multicast {
           // Good candidate for discovery
       }
   }
   ```

## Testing

```bash
# Run discovery tests
go test -v ./discovery/

# Run with specific interface test
go test -v ./discovery/ -run TestDiscoverWithOptions
```

## Related Documentation

- [QUICKSTART](../QUICKSTART.md) - Getting started with onvif-go
- [discovery/discovery.go](./discovery.go) - Source code
- [discovery/discovery_test.go](./discovery_test.go) - Test examples
