# onvif-go

[![CI](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml/badge.svg)](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mickeyzzc/onvif-go/v2.svg)](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/mickeyzzc/onvif-go/v2)](https://goreportcard.com/report/github.com/mickeyzzc/onvif-go/v2)
[![License](https://img.shields.io/github/license/mickeyzzc/onvif-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

> 面向 Go 的 ONVIF 客户端 + 虚拟相机服务器——服务门面 API、主动与被动双模发现、
> 能对付各种怪固件的鉴权梯队、PTZ / 媒体 / 事件 / 成像。零第三方依赖。
> 由 [@mickeyzzc](https://github.com/mickeyzzc) 维护。

> [**English**](README.md)

一个经过生产锤炼的 Go 库，用于与 ONVIF 兼容的网络摄像头、NVR 和监控设备通信。
设备兼容性来自真实部署：海康、安讯士（Axis）、大华、博世（Bosch）、Amcrest、
海思 OEM 硬件、极简嵌入式实现（ESP32），以及那种「不管问什么都回 SOAP Fault」
的固件。它是 [MiBeeNvr](https://github.com/Mi-Bee-Studio/MiBeeNvr) 的 ONVIF
底层——后者是跑在 ARM64 上 7×24 小时的 NVR。

## 该用哪个 onvif-go？

本仓库是 [0x524a/onvif-go](https://github.com/0x524a/onvif-go) 的持续维护
谱系（其 v1.0.0–v1.1.7 tag 保留在本仓历史中，见
[血统与许可](#血统与许可)）。上游项目仍在维护、两者同为 MIT——按需选择：

| 你要什么 | 选哪个 |
|---|---|
| 可嵌入自家二进制的虚拟相机**服务器**——状态可注入的 provider、按动作的鉴权策略、metrics 接缝、PullPoint 事件 | 本仓库（`/v2`） |
| 服务门面客户端 API、鉴权梯队、主动 + 被动 + 定向三种发现 | 本仓库（`/v2`） |
| 现有代码已在用 `github.com/0x524a/onvif-go` 且满足需求 | 上游 `v1.1.7`——留在那边没问题 |
| 从 v1 代码要上面的 v2 能力 | 本仓库；[MIGRATION.md](MIGRATION.md) 给出迁移映射（多数情况只需换 import 路径） |

两者模块路径不同（`…/onvif-go/v2` 与 `…/onvif-go`），迁移期间甚至可以在
同一个构建里共存。

## 功能特性

**客户端** —— 每个 ONVIF 操作挂在其所属的服务对象上，与 ONVIF 服务模型一一对应：

- `client.Device()` —— 设备信息、能力（带缓存、single-flight、弱设备可降级到
  最小能力集）、网络/系统/DNS/NTP 配置、存储、WiFi、证书、用户管理；
  类型化时间同步（`GetDeviceTime` / `SetDeviceTime` / `SetNTPServer`，
  写入路径带 `ErrNoCredentials` 凭据门——免鉴权的时钟读取恰好在时钟漂移
  打破 digest 鉴权时也可用）
- `client.Media()` —— profile 与主/子码流选择助手、流 URI
  （`GetStreamURIWithOptions`：RTSP/HTTP/UDP × 单播/组播，命名空间宽容解析）、
  快照、编码/音频/OSD 配置
- `client.PTZ()` —— 转动、状态、预置位、配置与选项、Home 位、辅助命令（雨刷/加热器）、节点、服务能力
- `client.Imaging()` —— 曝光、聚焦、成像设置、移动选项、规范名 `Stop`、服务能力、成像预置位（18.12）
- `client.Events()` —— 托管 PullPoint 订阅（后台轮询、自动续订、
  `ErrEventsNotSupported` 哨兵）+ 原始原语；WS-BaseNotification 推送
  （`Subscribe` 探测、`ErrPushNotSupported` 即回落 PullPoint；
  Renew/Unsubscribe 走订阅自身的 Manager 端点；`ParseNotifyEvents` +
  现成的 `NotifyHandler` 消费端点；托管 `SubscribePushEvents` 半 TTL
  续订循环）；标准运动话题助手（`ParseMotionAlarm` / `IsMotionTopic` ——
  MotionAlarm `State` 与 CellMotionDetector `IsMotion`）
- `client.DeviceIO()`、`client.Security()` —— 继电器、I/O、用户管理、音频输出配置族（音频**解码器**配置按 WSDL 归属 Media 服务，已在 media 包覆盖）

**为真实固件打造的鉴权** —— `WithAuthMode` 选择 digest / 明文 token /
HTTP Basic / 不鉴权；`WithAuthFallback` 提供自动回退梯队，并记住设备接受的
第一个模式。`errors.Is(err, onvif.ErrUnauthorized)` 可靠分类鉴权失败
（HTTP 401/403、NotAuthorized Fault、200 带 Fault）。`WithAutoClockSkew`
自动测量并校正设备时钟偏差（海康 "sender not authorized" 陷阱），
`DiagnoseAuth` 能区分时钟偏差、密码错误和拒绝 ONVIF 鉴权的设备。

**三种发现方式** —— `discovery.Discover`（组播探测）、`discovery.Listener`
（被动：听相机上电 Hello 和别人探测的应答，与主动发现共存；默认在回调前
丢弃非 ONVIF 应答者，Types/Scopes 约束可选）、
`discovery.ProbeEndpoint` / `ProbeSerial`（纯 HTTP 定向探测，跨子网触达组播
够不着的设备）。后处理：`FilterONVIFDevices`（滤掉 Synology/Windows/打印机
幽灵应答者）、`ParseScopes`、并行 `EnrichDevices`（限并发身份补全）。

**健壮性** —— 每次调用都做 Fault 检测（200 带 Fault 绝不会被判成成功）、
结构化 `FaultError`/`HTTPStatusError`、空 URI 响应变成带响应体摘要的显式
错误、能力 XAddr 修复（相机漫游后的陈旧宣传地址）、以及有文档有测试的并发
契约：一个 `Client`、多个 goroutine、无需外部加锁。

**虚拟相机服务器** —— `server/` 模拟 ONVIF 相机，不用硬件就能测你的录像软件：
设备/媒体/PTZ/成像 + 可选的事件 PullPoint 服务（`SupportEvents`），
宿主可经 `PublishEvent` 接缝注入自家分析产生的事件通知。

范围说明：本库只做 ONVIF 信令——产出和消费流 **URI**（RTSP/HTTP/UDP），
但不搬运媒体本身；媒体面请搭配 RTSP 服务器或客户端。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2
```

要求 Go **1.26+**。库模块零第三方依赖。

## 快速上手

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mickeyzzc/onvif-go/v2"
)

func main() {
    client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "pass"),
        onvif.WithAutoClockSkew()) // 应对时钟偏差的相机
    if err != nil {
        log.Fatal(err)
    }
    if err := client.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    info, err := client.Device().GetDeviceInformation(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s %s（固件 %s，序列号 %s）\n",
        info.Manufacturer, info.Model, info.FirmwareVersion, info.SerialNumber)

    profiles, err := client.Media().GetProfiles(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    mainToken := onvif.SelectMainProfile(profiles) // 分辨率最高者，而非 profiles[0]
    uri, err := client.Media().GetStreamURI(context.Background(), mainToken)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("主码流:", uri.URI)
}
```

固件拒绝 digest 鉴权？加一条梯队：

```go
client, err := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

发现——主动、被动或定向：

```go
devices, _ := discovery.Discover(ctx, 5*time.Second)          // 组播探测
usable := discovery.FilterONVIFDevices(devices)               // 滤掉幽灵应答者
discovery.EnrichDevices(ctx, usable)                          // 并行补全身份

listener, _ := discovery.NewListener("", func(d *discovery.Device) { // 被动
    fmt.Println("相机上线:", d.GetDeviceEndpoint())
})
go listener.Start(ctx)
defer listener.Stop()

dev := discovery.ProbeEndpoint(ctx, "192.168.2.50", 80, 1200*time.Millisecond) // 跨子网
serial, ok := discovery.ProbeSerial(ctx, "192.168.2.50", nil)                  // 身份锚点
```

事件——托管、自动续订：

```go
sub, err := client.Events().SubscribeEvents(ctx, func(msg onvif.NotificationMessage) {
    fmt.Println(msg.Topic, msg.Message.Data)
}, nil)
if errors.Is(err, onvif.ErrEventsNotSupported) {
    // 设备没有事件服务；缓存阴性结果
}
defer sub.Unsubscribe(context.Background())
```

事件，设备支持时走推送——先探测、失败回落、再起消费端点：

```go
push, err := client.Events().SubscribePushEvents(ctx, "http://nvr:8080/onvif/notify", nil)
if errors.Is(err, onvif.ErrPushNotSupported) {
    // 回落到上面的托管 PullPoint 订阅
}
defer push.Unsubscribe(context.Background())

// 设备 POST wsnt:Notify 的端点（自行包一层来源/token 校验——推送协议不携带凭据）：
http.Handle("/onvif/notify", onvif.NotifyHandler(func(msgs []onvif.NotificationMessage) {
    for _, m := range msgs {
        fmt.Println(m.Topic, m.Message.Data)
    }
}))
```

鉴权怎么都不通时，做个分诊：

```go
diag, _ := client.DiagnoseAuth(ctx)
// diag.Status: "ok" | "clock-skew"（修相机 NTP）| "bad-credentials"
```

## 协议对齐与路线图

线格式契约以官方 ONVIF WSDL/XSD 集（[onvif/specs](https://github.com/onvif/specs)，
全部 schema `elementFormDefault="qualified"`）为真源：服务 WSDL 内本地声明的
元素归属该 WSDL 命名空间，`ver10/schema` 类型的子元素归属 `tt`。全部已服务
响应面由命名空间严格解码的契约套件钉死；客户端↔模拟器一致性回环覆盖全操作
矩阵；客户端解码路径有敌意响应 fuzz。

明确的范围边界（留待后续能力包）：

- **Media2（ver20/media/wsdl）**——编解码无关的配置模型已具备
  （`onvif.Client.Media2()`：带内联配置的 profile、视频编码器配置与按编码
  的选项——每个编解码一条、`Encoding` 为自由名（H264/H265/AV1/…）、写入
  原样透传、取流 URI、快照 URI、编码器实例数、组播起停、
  `SetSynchronizationPoint`、服务能力、音频编码器与元数据配置族）。
  `Initialize` 支持纯 Media2 设备探测：GetCapabilities 无 ver10 Media
  XAddr 时经 GetServices 探测 ver20/media 端点并接给 `Media2()`。Profile
  增删与 OSD 配置族尚未实现；ver10 media 仍是全覆盖面。
- **Profile M**——事件 PullPoint 族、配置中的元数据配置、analytics 服务
  客户端（规则/模块配置 CRUD，`onvif.Client.Analytics()`，v2.1.0）均已具备；
  analytics `ConfigDescription` 的消息描述与元数据流的
  PTZ/Event/SensorData 成员尚未建模。
- **Profile T**——服务端已支持 TLS 传输（`Config.TLSCertFile`/`TLSKeyFile`）；
  高级安全置备（PKI、802.1X、UsernameToken 之外的 WS-Security）不在范围。
- **GetEventProperties** 应答规格完整（主题命名空间/方言位置、两个强制
  主题表达式方言——pull-point 过滤器如实执行——以及规范规定的空消息内容
  方言）；消息内容过滤不应用也不广播。
- **WebRTC**——保持关注，未实现。

## CLI 工具

四个辅助 CLI 自 v2.0.0 起随[每个 release](https://github.com/mickeyzzc/onvif-go/releases)
发布预编译零依赖单文件：linux（amd64/arm64/arm）、macOS（amd64/arm64）、
Windows（amd64）——下载、`chmod +x`、对照 `SHA256SUMS` 校验即可；或
`go install github.com/mickeyzzc/onvif-go/v2/cmd/<工具>@v2.2.0`。

哪个工具干什么：`discover` 扫描局域网（WS-Discovery）；`onvif-quick`
交互式体检单台相机（连接、流地址、带防护的 PTZ 演示）；`onvif-diagnostics`
对十一个主要操作做全量体检并出 JSON 报告——[提 issue](https://github.com/mickeyzzc/onvif-go/issues)
前先跑它；`onvif-server` 起一台虚拟多镜头相机，无硬件测试。各工具的
分场景手册见下方文档中心。

## 文档

专题手册已迁移至 MiBee 文档中心——库手册唯一真源（双语）：

> **https://www.mlsbs.top/docs/mibeelibs**

手册修改向文档中心提 PR（流程见该仓 GOVERNANCE）。本仓 [`docs/README.md`](docs/README.md) 保留引导页。
## 项目结构

| 路径 | 用途 |
|---|---|
| `client.go` 等（根目录） | `Client`（鉴权梯队、时钟偏差、下载）+ v1 兼容 alias |
| `device/` `security/` `deviceio/` `media/` `ptz/` `imaging/` `events/` | 域服务包（v2 拆分，issue #20） |
| `types/` | 共享数据模型 leaf |
| `discovery/` | WS-Discovery：主动探测、被动监听、定向 HTTP 探测、后处理 |
| `internal/soap/` | SOAP 传输 + WS-Security（digest/明文模式、Fault 检测） |
| `server/` | 虚拟相机模拟器 + 可嵌入服务端：DeviceIO（继电器/数字输入，opt-in）、Media2 最小面（opt-in）、媒体 OSD 闭环 + 音频空集族 |
| `internal/onviftesting/` | 测试助手：mock server、抓包回放、golden 文件（内部包，不进消费方二进制） |
| `testdata/captures/` | 真机 SOAP 抓包回归 fixture |
| `docs/` | 引导至文档中心 |
| `cmd/` | 辅助 CLI：`discover`、`onvif-quick`、`onvif-diagnostics`、`onvif-server`（另有 `generate-tests`，抓包→fixture 的开发者工具） |
| `examples/` | 按功能划分的可运行示例：discovery、device-info、imaging-settings、ptz-control、events（PullPoint 订阅）、media2（H.265 选项）、analytics（Profile M，含规则增删改+选项探测）、metadata（流解析）、simple-server、tls-server、onvif-server、complete-demo |

## 开发

```bash
make build    # go build ./...
make test     # go test -race ./...
make lint     # golangci-lint run（安装：make lint-install）
make fmt      # 通过 golangci-lint fmt 执行 gofumpt + goimports
```

CI（[ci.yml](.github/workflows/ci.yml)）在每次推送到 `main` 时运行 lint +
格式检查 + race 测试 + 构建；分支要求三个任务全绿，且只接受 PR 合并。

## 库卫生（rc4 加固）

- `server.DefaultConfig()` **不再内置 admin/admin 凭据**（默认为空；未配凭据
  即文档化的全开放模式，暴露真实网络前请显式设置 `Username`/`Password`）；
- 公开的 `testing` 包移入 `internal/onviftesting`（不再把 `httptest` 链进
  消费方二进制）；
- 三个可变默认切片改为返回新切片的函数：`server/soap.DefaultProtectedPrefixes()`、
  `server.DefaultScopes()`、`discovery.DefaultProbePorts()`；
- `testdata/` 只含合成数据（RFC 5737 文档网段），真实抓包请放 gitignored
  的 `tmp/`；
- 仓库根 `hygiene_test.go` 源码扫描守卫钉死以上各条。迁移细节见
  [CHANGELOG.md](CHANGELOG.md) 与 [MIGRATION.md](MIGRATION.md)。

## 血统与许可

本项目是 [0x524a/onvif-go](https://github.com/0x524a/onvif-go) 的持续维护版
（其本身源自更早的 ONVIF Go 尝试——见 git 历史）。模块路径在 **v1.2.0** 起变更为
`github.com/mickeyzzc/onvif-go`；此前以 `github.com/0x524a/onvif-go` 发布
（tag v1.0.0–v1.1.7 保留在本仓库作历史存档，永不删除）。v1.2.0 之后的延续性
变更记录在[变更日志](CHANGELOG.md)。

基于 [MIT 许可证](LICENSE) 发布。衷心感谢原作者的贡献，它们仍受同一许可保护。
