# onvif-cli Non-Interactive Mode Guide

## Overview

`onvif-cli` now supports both **interactive mode** (default) and **non-interactive mode** with command-line arguments. This makes it suitable for:

- Shell scripts and automation
- Docker containers
- Continuous integration/deployment (CI/CD)
- Batch operations
- Programmatic camera management
- Cron jobs

## Modes

### Interactive Mode (Default)

```bash
./onvif-cli
# Menu-driven interface with prompts
```

### Non-Interactive Mode

```bash
./onvif-cli -e <endpoint> -u <username> -p <password> -op <operation>
# Direct command execution without prompts
```

## Command-Line Flags

### Required Flags (for non-discovery operations)

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `-endpoint` | `-e` | Camera endpoint URL | `http://192.168.1.100/onvif/device_service` |
| `-username` | `-u` | Username | `admin` |
| `-password` | `-p` | Password | `mypassword` |
| `-operation` | `-op` | Operation to perform | `info`, `profiles`, `stream`, etc. |

### Optional Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `-interface` | `-i` | Network interface for discovery | (system default) |
| `-timeout` | `-t` | Request timeout in seconds | `30` |
| `-non-interactive` | `-ni` | Force non-interactive mode | false |
| `-help` | `-h` | Show help message | false |

## Supported Operations

### Non-Discovery Operations (require endpoint + credentials)

| Operation | Description | Output |
|-----------|-------------|--------|
| `info` | Get device information | Manufacturer, model, firmware, serial number |
| `capabilities` | Get device capabilities | List of supported services |
| `profiles` | Get media profiles | Profile names and encoding info |
| `stream` | Get stream URI | RTSP stream URL |
| `snapshot` | Get snapshot URI | Snapshot URL |
| `datetime` | Get system date/time | Device system time |

### Discovery Operations (no credentials needed)

| Operation | Description |
|-----------|-------------|
| `discover` | Discover cameras on network |

## Usage Examples

### Example 1: Get Device Information

```bash
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op info
```

**Output:**
```
🔗 Connecting to http://192.168.1.100/onvif/device_service...
✅ Connected to Hikvision DS-2CD2143G2-I

📋 Device Information:
  Manufacturer: Hikvision
  Model: DS-2CD2143G2-I
  Firmware: V5.4.41 build 201111
  Serial Number: DS-2CD2143G2-I5C28D1234
  Hardware ID: 2cd2
```

### Example 2: Get Media Profiles

```bash
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op profiles
```

**Output:**
```
✅ Found 2 profile(s):

Profile 1: Profile000
  Token: Profile000
  Encoding: H264

Profile 2: Profile001
  Token: Profile001
  Encoding: H265
```

### Example 3: Get Stream URI

```bash
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op stream
```

**Output:**
```
✅ Stream URI: rtsp://192.168.1.100:554/stream1
```

### Example 4: Get Capabilities

```bash
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op capabilities
```

**Output:**
```
✅ Capabilities:
  ✓ Device Service
  ✓ Media Service (Streaming)
  ✓ PTZ Service
  ✓ Imaging Service
  ✓ Events Service
```

### Example 5: Discover Cameras (Default Interface)

```bash
onvif-cli -op discover -t 5
```

**Output:**
```
🔍 Discovering ONVIF cameras...
✅ Found 2 camera(s):

Camera 1:
  Endpoint: http://192.168.1.100:8080/onvif/device_service
  Name: Office Camera

Camera 2:
  Endpoint: http://192.168.1.101:8080/onvif/device_service
  Name: Conference Room Camera
```

### Example 6: Discover on Specific Interface

```bash
# By interface name
onvif-cli -op discover -i eth0 -t 5

# By IP address
onvif-cli -op discover -i 192.168.1.100 -t 5
```

### Example 7: Custom Timeout

```bash
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op info \
          -t 60  # 60 second timeout
```

## Scripting Examples

### Shell Script: Discover and Get Endpoints

```bash
#!/bin/bash

# Discover cameras on eth0
cameras=$(onvif-cli -op discover -i eth0 -t 5)

if echo "$cameras" | grep -q "No ONVIF cameras"; then
    echo "No cameras found"
    exit 1
fi

echo "Cameras found:"
echo "$cameras"
```

### Shell Script: Get Info from Multiple Cameras

```bash
#!/bin/bash

declare -a CAMERAS=(
    "http://192.168.1.100/onvif/device_service"
    "http://192.168.1.101/onvif/device_service"
)

for endpoint in "${CAMERAS[@]}"; do
    echo "Getting info from $endpoint..."
    onvif-cli -e "$endpoint" -u admin -p password -op info
    echo ""
done
```

### Shell Script: Get Stream URIs and Save to File

```bash
#!/bin/bash

OUTPUT_FILE="stream_urls.txt"
> "$OUTPUT_FILE"  # Clear file

for i in {1..10}; do
    ip="192.168.1.$((100+i))"
    endpoint="http://$ip/onvif/device_service"
    
    stream=$(onvif-cli -e "$endpoint" -u admin -p password -op stream 2>/dev/null | grep "Stream URI")
    
    if [ -n "$stream" ]; then
        echo "$ip: $stream" >> "$OUTPUT_FILE"
    fi
done

echo "Stream URLs saved to $OUTPUT_FILE"
```

### Python Script: Query Cameras

```python
#!/usr/bin/env python3

import subprocess
import json
import sys

def get_camera_info(endpoint, username, password):
    """Get camera information using onvif-cli"""
    cmd = [
        "onvif-cli",
        "-e", endpoint,
        "-u", username,
        "-p", password,
        "-op", "info"
    ]
    
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        return result.stdout
    except subprocess.TimeoutExpired:
        return None

def get_stream_uri(endpoint, username, password):
    """Get RTSP stream URL"""
    cmd = [
        "onvif-cli",
        "-e", endpoint,
        "-u", username,
        "-p", password,
        "-op", "stream"
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    return result.stdout.strip()

# Example: Get info from multiple cameras
cameras = [
    ("http://192.168.1.100/onvif/device_service", "admin", "password"),
    ("http://192.168.1.101/onvif/device_service", "admin", "password"),
]

for endpoint, username, password in cameras:
    print(f"\n=== {endpoint} ===")
    info = get_camera_info(endpoint, username, password)
    print(info)
    
    stream_uri = get_stream_uri(endpoint, username, password)
    print(f"Stream: {stream_uri}")
```

### Docker Usage

```bash
# Build image
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o onvif-cli ./cmd/onvif-cli

FROM alpine:latest
COPY --from=builder /app/onvif-cli /usr/local/bin/

# Usage
CMD ["onvif-cli", "-e", "http://camera:8080/onvif/device_service", \
     "-u", "admin", "-p", "password", "-op", "info"]
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (camera not found, connection failed, etc.) |

## Error Handling

```bash
#!/bin/bash

onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op info

if [ $? -eq 0 ]; then
    echo "✅ Camera info retrieved successfully"
else
    echo "❌ Failed to get camera info"
    exit 1
fi
```

## Tips & Best Practices

### 1. Use Environment Variables for Credentials

```bash
export CAMERA_IP="192.168.1.100"
export CAMERA_USER="admin"
export CAMERA_PASS="mypassword"

onvif-cli -e "http://$CAMERA_IP/onvif/device_service" \
          -u "$CAMERA_USER" -p "$CAMERA_PASS" \
          -op profiles
```

### 2. Batch Processing with Timeout

```bash
# Set a timeout for each operation
timeout 10 onvif-cli -e http://192.168.1.100/onvif/device_service \
                     -u admin -p password \
                     -op info
```

### 3. Logging Output

```bash
# Log to file with timestamp
{
    echo "=== $(date) ==="
    onvif-cli -e http://192.168.1.100/onvif/device_service \
              -u admin -p password \
              -op capabilities
} >> camera_query.log
```

### 4. Discovery with Interface Selection

```bash
# First list available interfaces
./onvif-cli -h  # Shows help

# Then discover on specific interface
onvif-cli -op discover -i eth0

# Or by IP
onvif-cli -op discover -i 192.168.1.0
```

### 5. Handling Errors in Scripts

```bash
#!/bin/bash

check_camera() {
    local endpoint="$1"
    local user="$2"
    local pass="$3"
    
    if onvif-cli -e "$endpoint" -u "$user" -p "$pass" -op info &>/dev/null; then
        echo "✅ Camera responsive"
        return 0
    else
        echo "❌ Camera not responsive"
        return 1
    fi
}

# Check multiple cameras
for i in {1..5}; do
    check_camera "http://192.168.1.$((100+i))/onvif/device_service" \
                 "admin" "password"
done
```

## Comparison: Interactive vs Non-Interactive

| Aspect | Interactive | Non-Interactive |
|--------|-------------|-----------------|
| User prompts | Yes | No |
| Automation | Poor | Excellent |
| Scripts | Not suitable | Perfect |
| Docker/CI | Difficult | Ideal |
| Learning curve | Easy | Medium |
| Speed | Slow | Fast |

## Troubleshooting

### Problem: "Connection refused"

```bash
# Check if endpoint is reachable
curl -I http://192.168.1.100/onvif/device_service

# Try with explicit timeout
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p password \
          -op info \
          -t 60
```

### Problem: "Invalid credentials"

```bash
# Verify username and password
# Try interactive mode first to test credentials
./onvif-cli

# Then use correct credentials in non-interactive mode
onvif-cli -e http://192.168.1.100/onvif/device_service \
          -u admin -p correctpassword \
          -op info
```

### Problem: Discovery finds no cameras

```bash
# List available interfaces first
./onvif-cli -h

# Try specific interface
onvif-cli -op discover -i eth0 -t 10

# Try different interface
onvif-cli -op discover -i wlan0 -t 10
```

## Advanced: Creating Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias camera-info='onvif-cli -e http://192.168.1.100/onvif/device_service -u admin -p password -op info'
alias camera-stream='onvif-cli -e http://192.168.1.100/onvif/device_service -u admin -p password -op stream'
alias discover-cameras='onvif-cli -op discover -t 5'

# Usage
camera-info
camera-stream
discover-cameras
```

## API Integration

### In Go Programs

```go
package main

import (
	"os/exec"
	"strings"
)

func getCameraInfo(endpoint, username, password string) (string, error) {
	cmd := exec.Command("onvif-cli",
		"-e", endpoint,
		"-u", username,
		"-p", password,
		"-op", "info")
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}
```

## Summary

Non-interactive mode makes `onvif-cli` suitable for:
- ✅ Automation and scripting
- ✅ Docker containers
- ✅ CI/CD pipelines
- ✅ Batch processing
- ✅ Integration with other tools
- ✅ Programmatic access

All while maintaining backward compatibility with the interactive mode!
