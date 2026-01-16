# CLI Tools with Network Interface Support

This guide shows how to use the enhanced CLI tools with network interface discovery capabilities.

## Overview

Both `onvif-cli` and `onvif-quick` now support explicit network interface selection when discovering ONVIF cameras. This is useful when you have multiple network interfaces on your system.

## onvif-cli - Full-featured CLI

### Building onvif-cli

```bash
# From the project root
go build -o onvif-cli ./cmd/onvif-cli
```

### Running onvif-cli

```bash
./onvif-cli
```

### Main Menu Features

```
📋 Main Menu:
  1. Discover Cameras on Network
  2. List Network Interfaces
  3. Connect to Camera
  4. Device Operations
  5. Media Operations
  6. PTZ Operations
  7. Imaging Operations
  0. Exit
```

### Feature 1: List Network Interfaces

Select option `2` to see all available network interfaces:

```
🌐 Available Network Interfaces
================================
✅ Found 3 interface(s):

📡 lo (⬆️  Up, Multicast: ✓)
   └─ 127.0.0.1
   └─ ::1

📡 eth0 (⬆️  Up, Multicast: ✓)
   └─ 192.168.1.100
   └─ fe80::1

📡 wlan0 (⬆️  Up, Multicast: ✓)
   └─ 192.168.88.50

💡 Use interface name or IP address when discovering cameras
   Example: eth0 or 192.168.1.100
```

### Feature 2: Discover with Interface Selection

Select option `1` for camera discovery:

```
🔍 Discovering ONVIF cameras...
This may take a few seconds...
Use specific network interface? (y/n) [n]: y

🌐 Available network interfaces:
  1. lo
     └─ 127.0.0.1
     (Up: true, Multicast: No)
  2. eth0
     └─ 192.168.1.100
     (Up: true, Multicast: Yes)
  3. wlan0
     └─ 192.168.88.50
     (Up: true, Multicast: Yes)

Enter interface name or IP address: eth0
🎯 Using interface: eth0

✅ Found 2 camera(s):

📹 Camera #1:
   Endpoint: http://192.168.1.101:8080/onvif/device_service
   Name: Office Camera
   Location: Conference Room A
   Types: [...]
   XAddrs: [...]
```

### Usage Scenarios

#### Scenario 1: Quick Camera Discovery (Default Interface)

```bash
./onvif-cli
# Select: 1 (Discover)
# Answer: n (use default interface)
# Result: Discovers on system default interface
```

#### Scenario 2: Discover on Specific Ethernet Interface

```bash
./onvif-cli
# Select: 2 (List interfaces)
# See eth0 is available with 192.168.1.100
# Select: 1 (Discover)
# Answer: y (use specific interface)
# Enter: eth0 or 192.168.1.100
# Result: Discovers only on eth0
```

#### Scenario 3: Discover on WiFi Interface

```bash
./onvif-cli
# Select: 2 (List interfaces)
# See wlan0 is available with 192.168.88.50
# Select: 1 (Discover)
# Answer: y (use specific interface)
# Enter: wlan0
# Result: Discovers only on wlan0
```

#### Scenario 4: Connect and Control

```bash
./onvif-cli
# Select: 1 (Discover) -> Find camera -> Connect
# Or: Select: 3 (Connect) -> Enter endpoint manually
# Then use options 4-7 for device/media/ptz/imaging control
```

## onvif-quick - Quick Demo Tool

### Building onvif-quick

```bash
# From the project root
go build -o onvif-quick ./cmd/onvif-quick
```

### Running onvif-quick

```bash
./onvif-quick
```

### Main Menu Features

```
What would you like to do?
1. 🔍 Discover cameras
2. 🌐 List network interfaces
3. 📹 Connect to camera
4. 🎮 PTZ demo
5. 📡 Get stream URLs
0. Exit
```

### Feature 1: List Network Interfaces

Select option `2`:

```
🌐 Network Interfaces
====================
✅ Found 3 interface(s):

📡 lo (Up, Multicast: No)
   └─ 127.0.0.1
   └─ ::1

📡 eth0 (Up, Multicast: Yes)
   └─ 192.168.1.100
   └─ fe80::1

📡 wlan0 (Up, Multicast: Yes)
   └─ 192.168.88.50
```

### Feature 2: Quick Discovery with Interface Selection

Select option `1`:

```
🔍 Discovering cameras on network...
Use specific network interface? (y/n) [n]: y

Available interfaces:
  1. lo (127.0.0.1, ::1)
  2. eth0 (192.168.1.100, fe80::1)
  3. wlan0 (192.168.88.50)

Enter interface name or IP: eth0
✅ Found 1 camera(s):
  1. Office Camera (http://192.168.1.101:8080/onvif/device_service)
```

### Quick Demo Workflows

#### Workflow 1: List Interfaces → Discover → Check Streams

```bash
./onvif-quick
# Select: 2 (List interfaces)
# See which interfaces are available
# Select: 1 (Discover)
# Choose eth0
# Specify credentials when found
# Select: 5 (Get stream URLs) to see RTSP streams
```

#### Workflow 2: PTZ Demo on Specific Interface

```bash
./onvif-quick
# Select: 1 (Discover) on eth0
# Find PTZ-capable camera
# Select: 4 (PTZ demo)
# Test pan/tilt/zoom movements
```

## Common Workflows

### Workflow A: Multi-Network Environment

You have a system with both Ethernet (192.168.1.0/24) and WiFi (192.168.88.0/24):

```bash
./onvif-cli

# Step 1: List interfaces
1 (Discover)
n (default)
# No results?

# Step 2: Try Ethernet explicitly
1 (Discover)
y (specific interface)
eth0
# Found cameras on ethernet!

# Step 3: Try WiFi
1 (Discover)
y (specific interface)
wlan0
# Found different cameras on WiFi!
```

### Workflow B: Docker Container with Multiple Networks

Container has management (172.17.0.x) and camera (172.20.0.x) networks:

```bash
./onvif-quick

# Step 1: See available networks
2 (List interfaces)
# Output shows two networks with different IPs

# Step 2: Discover on camera network
1 (Discover)
y (specific interface)
172.20.0.10  # Use the camera network IP
# Discovers cameras on the camera network
```

### Workflow C: Network Troubleshooting

Discovery not working as expected?

```bash
./onvif-cli

# Step 1: Check all interfaces
2 (List interfaces)
# Look for:
# - Interfaces marked "Up: true"
# - Multicast support: Yes
# - Expected IP addresses

# Step 2: Try discovery on each interface
1 (Discover)
y (use specific interface)
# Try each interface name one by one
# See which one finds cameras

# Result: Identifies which network has your cameras
```

## Tips & Best Practices

### 1. Check Interface Status First

Always start with option 2 to see:
- Interface names (eth0, wlan0, docker0, etc.)
- IP addresses assigned
- Whether multicast is supported
- Whether the interface is up/down

```bash
# Quick check
./onvif-cli
2 (List interfaces)
```

### 2. Use Interface Names When Possible

Interface names are more reliable than IP addresses:

```
Good:  eth0, wlan0
Less good: 192.168.1.100 (may change)
```

### 3. Check Multicast Support

Ensure the interface supports multicast (required for WS-Discovery):

```
Look for: "Multicast: Yes" or "Multicast: ✓"
```

### 4. Isolate Discovery to One Network

If you have many interfaces, disable the ones you don't need:

```bash
./onvif-cli
1 (Discover)
y (specify eth0)
# Only discovers on eth0, ignores other interfaces
```

### 5. Scripting and Automation

For automation, you can pipe input:

```bash
# Non-interactive discovery on eth0
(echo 1; echo y; echo eth0; sleep 2; echo 0) | ./onvif-cli

# Or with timeout
timeout 30 bash -c '(echo 1; echo y; echo eth0) | ./onvif-cli'
```

## Troubleshooting

### Problem: "Use specific network interface?" appears on every discovery

**Solution**: This is the normal behavior in onvif-cli. To skip it, answer `n` to use the system default interface.

### Problem: Interface listed but discovery fails

**Possible causes**:
1. Interface doesn't support multicast (check "Multicast: Yes")
2. Cameras aren't on that network segment
3. Firewall blocking UDP 3702

**Solution**:
```bash
./onvif-cli
2 (List interfaces)
# Check Multicast: Yes
# Check interface is "Up: true"
1 (Discover)
y (use specific interface)
# Try the confirmed interface
```

### Problem: "network interface not found" error

**Solution**: 
1. Use `2 (List interfaces)` to see exact interface names
2. Copy the exact name from the list
3. Try again with correct interface name

```bash
# Wrong:  eth-0 or ethnet0
# Right:  eth0 (from list)
```

### Problem: No cameras found on any interface

**Possible causes**:
1. Cameras on different subnet
2. Firewall blocking discovery
3. ONVIF not enabled on cameras

**Solution**:
```bash
# Try each interface individually
./onvif-cli
2 (List interfaces)
# For each interface that shows "Multicast: Yes" and "Up: true"
1 (Discover)
y (use that interface)
# Check if cameras found
```

## Integration with Other Tools

### Using Discovered Camera with VLC

```bash
./onvif-cli
1 (Discover)
y (eth0)
# Get stream URL from discovered camera
2 (Get stream URIs)
# Copy RTSP URL
# Paste into VLC: File → Open Network Stream
```

### Scripting Camera Discovery

```bash
#!/bin/bash
# discover_cameras.sh

# List all interfaces with multicast support
./onvif-cli << EOF
2
q
EOF | grep "Multicast: ✓" | grep -o "📡 [^ ]*" | cut -d' ' -f2 | while read iface; do
    echo "Discovering on $iface..."
    # Could add automated discovery here
done
```

## Related Documentation

- [NETWORK_INTERFACE_GUIDE.md](../discovery/NETWORK_INTERFACE_GUIDE.md) - Detailed discovery API guide
- [QUICKSTART.md](../QUICKSTART.md) - Quick start guide
- [examples/discovery/](../examples/discovery/) - Discovery code examples
- [ONVIF Specification](https://www.onvif.org/) - Official ONVIF specs

## Command Reference

### onvif-cli Commands

| Option | Feature | Purpose |
|--------|---------|---------|
| 1 | Discover Cameras | Find ONVIF cameras (with interface selection) |
| 2 | List Interfaces | See all network interfaces |
| 3 | Connect to Camera | Manual endpoint connection |
| 4 | Device Operations | Info, capabilities, datetime, reboot |
| 5 | Media Operations | Profiles, streams, snapshots, video settings |
| 6 | PTZ Operations | Pan/tilt/zoom control and presets |
| 7 | Imaging Operations | Brightness, contrast, saturation, etc. |
| 0 | Exit | Quit the application |

### onvif-quick Commands

| Option | Feature | Purpose |
|--------|---------|---------|
| 1 | Discover Cameras | Find ONVIF cameras (quick, with interface selection) |
| 2 | List Interfaces | See all network interfaces |
| 3 | Connect to Camera | Quick connection and info |
| 4 | PTZ Demo | Quick PTZ movement demonstration |
| 5 | Get Stream URLs | Display all stream and snapshot URLs |
| 0 | Exit | Quit the application |

## Version History

- **Current**: Network interface selection support added
- **Previous**: Basic discovery and camera control
