# 设备发现

onvif-go 提供三种互补的发现模式和一个后处理层，全部位于自包含的 `discovery` 包。

| 模式 | 函数 | 覆盖范围 |
|---|---|---|
| 主动探测 | `Discover` / `DiscoverWithOptions` | 同子网（UDP 组播） |
| 被动监听 | `Listener` | 同子网、零延迟 |
| 定向探测 | `ProbeEndpoint` / `ProbeSerial` | 任意地址、跨子网、纯 HTTP |


## 发现交互时序

```mermaid
sequenceDiagram
    participant C as discovery 客户端
    participant D as ONVIF 设备
    C->>D: UDP Probe（组播 3702）
    D-->>C: ProbeMatches（XAddrs）
    C->>D: HTTP POST GetCapabilities
    D-->>C: 能力描述（Media/Device 地址）
    C->>D: GetProfiles / GetStreamUri
    D-->>C: Profile 与拉流地址
```

主动探测与定向探测都最终落到同一套 Probe/ProbeMatches 解析；被动 `Listener` 只监听网络上既有的 Probe 与 ProbeMatches，不主动发帧。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 主动：组播探测

```go
devices, err := discovery.Discover(ctx, 5*time.Second)
// 多网卡主机可指定组播接口（名称或 IP）：
devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second,
    &discovery.DiscoverOptions{NetworkInterface: "eth0"})
```

每分钟轮询一次是常见用法——但相机上电瞬间广播的 Hello 就浪费了，除非有人在
听。这正是被动监听器的职责。

## 被动：常驻监听器

```go
listener, _ := discovery.NewListener("", func(d *discovery.Device) {
    fmt.Println("相机上线:", d.GetDeviceEndpoint())
})
go func() { _ = listener.Start(ctx) }()
defer listener.Stop()
```

- 同时接收 **Hello**（上电自宣）与 **ProbeMatches**（设备应答别人的探测——
  白捡的第二发现源）；Bye 与畸形报文一律忽略。
- **与主动发现共存**：与 `DiscoverWithOptions` 完全相同的方式加入 WS-Discovery
  组播组（SO_REUSEADDR），Linux/macOS 内核会给每个绑定的 socket 各投一份——
  同进程里监听器 + 按需探测互不抢流量。
- `ifaceName` 传 `""` = 内核默认接口（单网卡主机推荐）。
- handler 在监听 goroutine 上执行：迅速返回，慢活自己开 goroutine；
  handler panic 会被隔离。
- `Stop()` 幂等；停止后拒绝重启。循环完全退出时 `Done()` 关闭。

## 定向：纯 HTTP 探测

组播不跨子网路由，部分设备也不回组播探测。相机 IP 自愈（换 IP 后按序列号
寻回）需要纯 HTTP：

```go
// 两级策略依次尝试：把 WS-Discovery Probe POST 到
// http://host:port/onvif/device_service，失败则发无鉴权
// GetDeviceInformation。返回 nil = 非 ONVIF / 离线
// （从外部无法区分，两者都算「没找到」）。
dev := discovery.ProbeEndpoint(ctx, "192.0.2.50", 80, 1200*time.Millisecond)

// 常用端口扫描（默认 80、8080、8000）取序列号——命名空间无关提取；
// 序列号是同一实体相机跨协议关联的身份锚点。
serial, ok := discovery.ProbeSerial(ctx, "192.0.2.50", nil)
```

来自任意主机的畸形响应都有 recover 防护（绝不 panic）；401/405 类应答按
「探测意义下非 ONVIF」处理。

## 后处理

三个助手把原始发现结果变成可用的设备清单：

```go
usable := discovery.FilterONVIFDevices(devices) // 滤掉幽灵应答者
info := discovery.ParseScopes(dev.Scopes)       // name / hardware / location
discovery.EnrichDevices(ctx, usable)            // 并行补全身份信息
```

- **`FilterONVIFDevices`** —— 通用 WS-Discovery 应答者（Synology DSM、
  Windows、打印机）对任何 Probe 都应答，混进来就成了相机清单里永远 pending
  的幽灵。保留条件（宽松取或）：Types 含 `NetworkVideoTransmitter`（按
  local-part 匹配，前缀无关）**或**任一 scope 以 `onvif://www.onvif.org/`
  开头——边缘实现不被误杀。
- **`ParseScopes`** —— 解析 ONVIF scope 约定（`onvif://www.onvif.org/
  name/X`、`.../hardware/Y`、`.../location/Z`），值做百分号解码。包内产出的
  每个 `Device` 也直接填好了结构化字段（`d.Name`、`d.Hardware`、`d.Location`）。
- **`EnrichDevices`** —— 并行对每台设备发无鉴权 `GetDeviceInformation`
  （`WithEnrichConcurrency` 限并发、`WithEnrichTimeout` 限单台超时），填充
  `d.Info`（厂商/型号/固件/序列号）。尽力而为：不可达设备静默跳过、绝不
  致命；已有 `Info` 的设备不动。

## 关于设备身份

`Device.EndpointRef` 是 WS-Discovery 端点地址（`urn:uuid:...` 形态）——
**不是**设备序列号。跨协议关联同一台相机（ONVIF vs GB28181）必须用
`Device.Info.SerialNumber`；拿 `EndpointRef` 比对序列号会静默地永远对不上。

## 设备侧：WS-Discovery 应答器

同一套编解码驱动设备侧（`server/discovery`）：常驻组播 `:3702` 监听，
以 ProbeMatches 应答 Probe（单播回探测方套接字），启动时发 Hello、
停止时发 Bye，并通过 `http.Handler` 应答定向的 Probe-over-HTTP POST
——即 `ProbeEndpoint` 跨子网探测所用的传输。

```go
responder := serverdiscovery.NewResponder(serverdiscovery.Config{
    Types:   []string{"tds:Device", "dp0:NetworkVideoTransmitter"},
    Scopes:  []string{"onvif://www.onvif.org/name/MiBeeEye"},
    Port:    8080, // 派生 XAddrs: http://<请求方 IP>:8080/onvif/device_service
})
err := responder.Start(ctx)
defer responder.Stop()
mux.Handle("/onvif/device_service", soapHandler)   // SOAP
mux.Handle("/onvif/discovery", responder)          // 定向 HTTP 探测
```

`XAddrs` 为空时应答器回显请求方的 IP——与 SOAP 层 XAddr 响应相同的
"对等可达"规则。应答内容（Types/Scopes/XAddrs/MetadataVersion）完全
可配；Types 过滤不匹配的探测按 WS-Discovery 语义忽略。

共存：应答器、`Discover()`、被动 `Listener` 以相同方式绑定组播组；
Linux/macOS 上内核会把每个组播数据报复制给所有绑定套接字，设备侧与
客户端在同一进程共存互不抢流量。应答永远是单播。

线上编解码本体在 `wsdiscovery/`（两侧共享）：`BuildProbe`/
`ParseProbe`、`BuildProbeMatches`/`ParseProbeMatches`、`BuildHello`/
`BuildBye`、`ParseAnnouncement`——一套定义，客户端与设备侧不再有
私有解析漂移。
