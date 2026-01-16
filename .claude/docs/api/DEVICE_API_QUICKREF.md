# ONVIF Device API Quick Reference

Quick reference for the most commonly used ONVIF Device Management APIs.

## Getting Started

```go
import "github.com/0x524a/onvif-go"

// Create client
client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
    onvif.WithCredentials("admin", "password"))
```

## Core Information

```go
// Device information
info, _ := client.GetDeviceInformation(ctx)
// Returns: Manufacturer, Model, FirmwareVersion, SerialNumber, HardwareID

// All capabilities
caps, _ := client.GetCapabilities(ctx)
// Returns: Analytics, Device, Events, Imaging, Media, PTZ capabilities

// Specific service capabilities
serviceCaps, _ := client.GetServiceCapabilities(ctx)
// Returns: Network, Security, System capabilities

// Available services
services, _ := client.GetServices(ctx, true) // include capabilities
// Returns: Namespace, XAddr, Version for each service

// Endpoint reference (device GUID)
guid, _ := client.GetEndpointReference(ctx)
```

## Network Configuration

```go
// Network interfaces
interfaces, _ := client.GetNetworkInterfaces(ctx)
for _, iface := range interfaces {
    fmt.Printf("%s: %s\n", iface.Info.Name, iface.Info.HwAddress)
}

// Network protocols (HTTP, HTTPS, RTSP)
protocols, _ := client.GetNetworkProtocols(ctx)
for _, proto := range protocols {
    fmt.Printf("%s: enabled=%v, ports=%v\n", proto.Name, proto.Enabled, proto.Port)
}

// Set protocol
client.SetNetworkProtocols(ctx, []*onvif.NetworkProtocol{
    {Name: onvif.NetworkProtocolHTTP, Enabled: true, Port: []int{80}},
    {Name: onvif.NetworkProtocolRTSP, Enabled: true, Port: []int{554}},
})

// Default gateway
gateway, _ := client.GetNetworkDefaultGateway(ctx)
client.SetNetworkDefaultGateway(ctx, &onvif.NetworkGateway{
    IPv4Address: []string{"192.168.1.1"},
})

// Zero configuration (auto IP)
zeroConf, _ := client.GetZeroConfiguration(ctx)
client.SetZeroConfiguration(ctx, "eth0", true)
```

## DNS & NTP

```go
// DNS configuration
dns, _ := client.GetDNS(ctx)
client.SetDNS(ctx, false, []string{"example.com"}, []onvif.IPAddress{
    {Type: "IPv4", IPv4Address: "8.8.8.8"},
})

// NTP configuration
ntp, _ := client.GetNTP(ctx)
client.SetNTP(ctx, false, []onvif.NetworkHost{
    {Type: "DNS", DNSname: "pool.ntp.org"},
})

// Dynamic DNS
ddns, _ := client.GetDynamicDNS(ctx)
client.SetDynamicDNS(ctx, onvif.DynamicDNSClientUpdates, "mycamera.dyndns.org")

// Hostname
hostname, _ := client.GetHostname(ctx)
client.SetHostname(ctx, "camera-01")
rebootNeeded, _ := client.SetHostnameFromDHCP(ctx, false)
```

## Discovery & Scopes

```go
// Discovery mode
mode, _ := client.GetDiscoveryMode(ctx)
client.SetDiscoveryMode(ctx, onvif.DiscoveryModeDiscoverable)

// Remote discovery
remoteMode, _ := client.GetRemoteDiscoveryMode(ctx)
client.SetRemoteDiscoveryMode(ctx, onvif.DiscoveryModeDiscoverable)

// Scopes
scopes, _ := client.GetScopes(ctx)
client.AddScopes(ctx, []string{
    "onvif://www.onvif.org/location/building/floor1",
    "onvif://www.onvif.org/name/camera-entrance",
})
removed, _ := client.RemoveScopes(ctx, []string{"old-scope"})
client.SetScopes(ctx, []string{"scope1", "scope2"}) // replaces all
```

## System Date & Time

```go
// Get current time
sysTime, _ := client.FixedGetSystemDateAndTime(ctx)
fmt.Printf("Mode: %s\n", sysTime.DateTimeType) // Manual or NTP
fmt.Printf("TZ: %s\n", sysTime.TimeZone.TZ)
fmt.Printf("UTC: %d-%02d-%02d %02d:%02d:%02d\n",
    sysTime.UTCDateTime.Date.Year,
    sysTime.UTCDateTime.Date.Month,
    sysTime.UTCDateTime.Date.Day,
    sysTime.UTCDateTime.Time.Hour,
    sysTime.UTCDateTime.Time.Minute,
    sysTime.UTCDateTime.Time.Second)

// Set time (manual mode)
client.SetSystemDateAndTime(ctx, &onvif.SystemDateTime{
    DateTimeType:    onvif.SetDateTimeManual,
    DaylightSavings: true,
    TimeZone:        &onvif.TimeZone{TZ: "EST5EDT,M3.2.0,M11.1.0"},
    UTCDateTime: &onvif.DateTime{
        Date: onvif.Date{Year: 2024, Month: 1, Day: 15},
        Time: onvif.Time{Hour: 10, Minute: 30, Second: 0},
    },
})

// Set time (NTP mode)
client.SetSystemDateAndTime(ctx, &onvif.SystemDateTime{
    DateTimeType:    onvif.SetDateTimeNTP,
    DaylightSavings: true,
    TimeZone:        &onvif.TimeZone{TZ: "EST5EDT,M3.2.0,M11.1.0"},
})
```

## User Management

```go
// List users
users, _ := client.GetUsers(ctx)
for _, user := range users {
    fmt.Printf("%s: %s\n", user.Username, user.UserLevel)
}

// Create user
client.CreateUsers(ctx, []*onvif.User{
    {Username: "operator1", Password: "SecurePass123", UserLevel: "Operator"},
})

// Modify user
client.SetUser(ctx, &onvif.User{
    Username: "operator1", Password: "NewPass456", UserLevel: "Administrator",
})

// Delete user
client.DeleteUsers(ctx, []string{"operator1"})

// Remote user (for connecting to other devices)
remoteUser, _ := client.GetRemoteUser(ctx)
client.SetRemoteUser(ctx, &onvif.RemoteUser{
    Username:           "admin",
    Password:           "password",
    UseDerivedPassword: true,
})
```

## Security & Access Control

```go
// IP address filter
filter, _ := client.GetIPAddressFilter(ctx)
client.SetIPAddressFilter(ctx, &onvif.IPAddressFilter{
    Type: onvif.IPAddressFilterAllow,
    IPv4Address: []onvif.PrefixedIPv4Address{
        {Address: "192.168.1.0", PrefixLength: 24},
        {Address: "10.0.0.0", PrefixLength: 8},
    },
})

// Add IP to filter
client.AddIPAddressFilter(ctx, &onvif.IPAddressFilter{
    Type: onvif.IPAddressFilterAllow,
    IPv4Address: []onvif.PrefixedIPv4Address{
        {Address: "172.16.0.0", PrefixLength: 12},
    },
})

// Remove IP from filter
client.RemoveIPAddressFilter(ctx, &onvif.IPAddressFilter{
    Type: onvif.IPAddressFilterAllow,
    IPv4Address: []onvif.PrefixedIPv4Address{
        {Address: "172.16.0.0", PrefixLength: 12},
    },
})

// Password complexity
pwdConfig, _ := client.GetPasswordComplexityConfiguration(ctx)
client.SetPasswordComplexityConfiguration(ctx, &onvif.PasswordComplexityConfiguration{
    MinLen:                  10,
    Uppercase:               2,
    Number:                  2,
    SpecialChars:            1,
    BlockUsernameOccurrence: true,
    PolicyConfigurationLocked: false,
})

// Password history
pwdHistory, _ := client.GetPasswordHistoryConfiguration(ctx)
client.SetPasswordHistoryConfiguration(ctx, &onvif.PasswordHistoryConfiguration{
    Enabled: true,
    Length:  5, // remember last 5 passwords
})

// Authentication failure warnings
authConfig, _ := client.GetAuthFailureWarningConfiguration(ctx)
client.SetAuthFailureWarningConfiguration(ctx, &onvif.AuthFailureWarningConfiguration{
    Enabled:         true,
    MonitorPeriod:   60,  // seconds
    MaxAuthFailures: 5,
})
```

## Relay & IO Control

```go
// Get relay outputs
relays, _ := client.GetRelayOutputs(ctx)
for _, relay := range relays {
    fmt.Printf("Relay %s: %s, idle=%s\n",
        relay.Token, relay.Properties.Mode, relay.Properties.IdleState)
}

// Configure relay
client.SetRelayOutputSettings(ctx, "relay1", &onvif.RelayOutputSettings{
    Mode:      onvif.RelayModeBistable,
    IdleState: onvif.RelayIdleStateClosed,
})

// Control relay state
client.SetRelayOutputState(ctx, "relay1", onvif.RelayLogicalStateActive)   // ON
client.SetRelayOutputState(ctx, "relay1", onvif.RelayLogicalStateInactive) // OFF
```

## Auxiliary Commands

```go
// Wiper control
client.SendAuxiliaryCommand(ctx, "tt:Wiper|On")
client.SendAuxiliaryCommand(ctx, "tt:Wiper|Off")

// IR illuminator
client.SendAuxiliaryCommand(ctx, "tt:IRLamp|On")
client.SendAuxiliaryCommand(ctx, "tt:IRLamp|Off")
client.SendAuxiliaryCommand(ctx, "tt:IRLamp|Auto")

// Washer
client.SendAuxiliaryCommand(ctx, "tt:Washer|On")
client.SendAuxiliaryCommand(ctx, "tt:Washer|Off")

// Full washing procedure
client.SendAuxiliaryCommand(ctx, "tt:WashingProcedure|On")
```

## System Maintenance

```go
// System logs
systemLog, _ := client.GetSystemLog(ctx, onvif.SystemLogTypeSystem)
accessLog, _ := client.GetSystemLog(ctx, onvif.SystemLogTypeAccess)
fmt.Println(systemLog.String)

// System URIs (for HTTP download)
logUris, supportUri, backupUri, _ := client.GetSystemUris(ctx)
// Download via HTTP GET from returned URIs

// Support information
supportInfo, _ := client.GetSystemSupportInformation(ctx)
fmt.Println(supportInfo.String)

// Backup
backupFiles, _ := client.GetSystemBackup(ctx)
for _, file := range backupFiles {
    fmt.Printf("Backup: %s (%s)\n", file.Name, file.Data.ContentType)
}

// Restore
client.RestoreSystem(ctx, backupFiles)

// Factory reset
client.SetSystemFactoryDefault(ctx, onvif.FactoryDefaultSoft) // soft reset
client.SetSystemFactoryDefault(ctx, onvif.FactoryDefaultHard) // hard reset

// Reboot
message, _ := client.SystemReboot(ctx)
fmt.Println(message)
```

## Firmware Upgrade

```go
// Start firmware upgrade (HTTP POST method)
uploadUri, delay, downtime, _ := client.StartFirmwareUpgrade(ctx)
// 1. Wait for delay duration
// 2. HTTP POST firmware file to uploadUri
// 3. Device will reboot after upgrade

// Start system restore (HTTP POST method)
uploadUri, downtime, _ := client.StartSystemRestore(ctx)
// 1. HTTP POST backup file to uploadUri
// 2. Device will restore and reboot
```

## Error Handling

All functions return errors that should be checked:

```go
info, err := client.GetDeviceInformation(ctx)
if err != nil {
    log.Fatalf("GetDeviceInformation failed: %v", err)
}

// Context timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

info, err := client.GetDeviceInformation(ctx)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out")
    } else {
        log.Printf("Error: %v", err)
    }
}
```

## Best Practices

1. **Always use context with timeout** for network operations
2. **Check capabilities first** before calling optional features
3. **Handle errors gracefully** - devices may not support all operations
4. **Use TLS skip verify** for self-signed certificates: `WithInsecureSkipVerify()`
5. **Check reboot requirements** when changing network settings
6. **Backup configuration** before factory reset or firmware upgrade
7. **Test on non-production devices** first

## Common Patterns

### Check if feature is supported
```go
caps, _ := client.GetCapabilities(ctx)
if caps.Device != nil && caps.Device.Network != nil {
    if caps.Device.Network.IPFilter {
        // IP filtering is supported
        filter, _ := client.GetIPAddressFilter(ctx)
    }
}
```

### Safe configuration change
```go
// 1. Get current config
currentConfig, _ := client.GetNetworkProtocols(ctx)

// 2. Modify
newConfig := currentConfig
newConfig[0].Port = []int{8080}

// 3. Apply
err := client.SetNetworkProtocols(ctx, newConfig)
if err != nil {
    // Restore original if needed
    log.Printf("Failed to apply config: %v", err)
}
```

### Batch operations
```go
// Create multiple users at once
client.CreateUsers(ctx, []*onvif.User{
    {Username: "user1", Password: "pass1", UserLevel: "Operator"},
    {Username: "user2", Password: "pass2", UserLevel: "User"},
    {Username: "admin2", Password: "pass3", UserLevel: "Administrator"},
})

// Delete multiple users
client.DeleteUsers(ctx, []string{"user1", "user2"})

// Add multiple scopes
client.AddScopes(ctx, []string{"scope1", "scope2", "scope3"})
```

## Geo Location & Discovery

```go
// Get device location (GPS coordinates)
locations, _ := client.GetGeoLocation(ctx)
for _, loc := range locations {
    fmt.Printf("%s: (%.4f, %.4f) elevation %.1fm\n",
        loc.Entity, loc.Lat, loc.Lon, loc.Elevation)
}

// Set location
client.SetGeoLocation(ctx, []onvif.LocationEntity{
    {
        Entity:    "Main Building",
        Token:     "loc1",
        Fixed:     true,
        Lon:       -122.4194,
        Lat:       37.7749,
        Elevation: 10.5,
    },
})

// Get WS-Discovery multicast addresses
dpAddresses, _ := client.GetDPAddresses(ctx)
for _, addr := range dpAddresses {
    fmt.Printf("%s: %s / %s\n", addr.Type, addr.IPv4Address, addr.IPv6Address)
}

// Set discovery addresses (empty list restores defaults)
client.SetDPAddresses(ctx, []onvif.NetworkHost{
    {Type: "IPv4", IPv4Address: "239.255.255.250"},
    {Type: "IPv6", IPv6Address: "ff02::c"},
})

// Get device access policy
policy, _ := client.GetAccessPolicy(ctx)
if policy.PolicyFile != nil {
    fmt.Printf("Policy: %d bytes of %s\n",
        len(policy.PolicyFile.Data),
        policy.PolicyFile.ContentType)
}
```

## See Also

- [DEVICE_API_STATUS.md](DEVICE_API_STATUS.md) - Complete API implementation status
- [README.md](README.md) - Main project documentation
- [ONVIF Specification](https://www.onvif.org/specs/DocMap-2.6.html)
