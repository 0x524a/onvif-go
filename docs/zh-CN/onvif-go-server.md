# ONVIF 服务端

`server/` 是一个可嵌入也可独立运行的 ONVIF 服务端实现：带 Device、
Media、PTZ、Imaging 服务的虚拟摄像头模拟器。v2 传输层（`server/soap`）
面向真实设备嵌入重构——请求上下文、按动作鉴权、字节级可预测的 XML
输出。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 传输层架构

```http
POST /onvif/device_service
        │
        ▼
soap.Handler.ServeHTTP
        │  1. 提取动作（local name，规范 化：GetStreamURI → GetStreamUri）
        │  2. 解码信封（WS-Security 头 + 原始 inner-XML 请求体）
        │  3. requiresAuth(action)? → authenticate(header)
        │  4. 分发：ContextHandler(RequestContext, body []byte)
        ▼
响应写出器（golden 测试锁定的字节布局，RawXML 直通）
```

请求体以**原始 inner XML 字节**到达 handler——用 `soap.ParseRequest`
（或直接 `encoding/xml`）解码成具体请求类型。v1 传输层交给 handler 的
值是 `encoding/xml` 永远填不进去的空值，请求参数（如 `ProfileToken`）
静默丢失；v2 修好了这条管道。

## 请求上下文

以 `RegisterContextHandler` 注册的 handler 能感知请求来源——多网卡
地址通告的前提：

```go
handler.RegisterContextHandler("GetStreamUri", func(rc *soap.RequestContext, body []byte) (interface{}, error) {
    ip := rc.RemoteIP              // 客户端 IP（RemoteAddr 的 host 部分）
    ctx := rc.Context()            // 请求 context：取消、超时
    _ = rc.Request                 // 底层 *http.Request
    ...
})
```

`rc.Action` 携带 WSDL 规范动作名。旧签名
`RegisterHandler(action, func(body []byte) ...)` 作为薄包装继续可用。

### XAddr 回显

真实摄像头对每个客户端回答**从该客户端网络可达**的 URL。模拟器默认
采取同样行为：

- `Config.AdvertiseHost` 为空（默认）：把请求方源 IP 回显为所有通告
  URL 的 host——GetCapabilities/GetServices 的 XAddr、流/快照 URI。
- `Config.AdvertiseHost` 非空：该 host 处处生效（固定 DNS 名、反向
  代理等场景）。

## 按动作鉴权（#16）

`server.DefaultConfig()` 的**凭据默认为空** —— 库的默认值绝不能内置可猜
测的账号对。未配置凭据时全部动作开放（适合本地模拟器）；把服务端暴露到
真实网络前，必须显式设置 `Username`/`Password`：

```go
cfg := server.DefaultConfig()
cfg.Username = "admin"      // 自行决定
cfg.Password = "..."        // 来自宿主配置，绝不硬编码提交
```

配置了凭据时，默认策略只对**写类动作**要求认证——`Set*`、`Remove*`、
`Create*`、`Go*`，外加 `SystemReboot` 和 `Config.AuthProtectedActions`
里列出的名字。读操作对无凭据的发现客户端保持开放。未配置凭据时全部
开放（行为不变）。

WS-Security UsernameToken 两种口令形式都接受：

| 形式 | 校验 |
|---|---|
| PasswordDigest（默认） | `Base64(SHA1(nonce + created + password))`，常量时间比较 |
| PasswordText | 明文比较；`AuthPolicy{AllowPasswordText: false}` 可关闭 |

```go
handler := soap.NewHandlerWithOptions(soap.HandlerOptions{
    Username: "admin",
    Password: "secret",
    Auth: &soap.AuthPolicy{          // nil → DefaultAuthPolicy()
        Prefixes: []string{"Set", "Remove", "Create", "Go"},
        Actions: []string{"SystemReboot"},
        All:     false,              // true = 严格模式（v1 行为）
    },
})
```

## 字节级 XML 输出（#18）

响应写出器手工拼装信封；字节布局由 golden 测试锁定，做字节级匹配的
消费方因此保持稳定。

**规范大小写。** 响应元素使用 ONVIF WSDL 拼写——
`GetStreamUriResponse`/`GetSnapshotUriResponse`（不是 `...URI...`），
内层 `MediaUri`/`Uri`。进入的旧拼写（`GetStreamURI`）仍分发到规范
handler。

**默认形式**（与历史线上格式一致）：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
      <MediaUri>
        <Uri>rtsp://198.51.100.7:8554/stream0</Uri>
```

**显式前缀**（`Config.ExplicitPrefixes` 或
`HandlerOptions.ExplicitPrefixes`）：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <trt:Uri>rtsp://198.51.100.7:8554/stream0</trt:Uri>
```

前缀表：`tds:` `trt:` `tptz:` `timg:` `tev:` `tt:` `trc:` `tan:`。
无约定前缀的命名空间保持默认 `xmlns` 声明形式。

**Raw 通道。** 任意 handler 返回 `soap.RawXML` 即可把预构建字节原样
嵌入——完全绕过 `encoding/xml`，精确控制元素名、前缀、格式。Raw 输出
不会被前缀模式改写。

## 可插拔 provider（M3）

SOAP 层背后的每个状态来源都是 `server/provider` 里的接口——模拟器
（`server/simulator`）只是默认实现：

| Provider | 支撑 | 用于替换 |
|---|---|---|
| `DeviceInfoProvider` | GetDeviceInformation | 硬件标识 / EEPROM |
| `StreamURIProvider`（可选 `StreamURISetter`） | GetStreamUri、RTSP 路径 | 真实 RTSP 挂载点 |
| `SnapshotProvider` | HTTP `/snapshot` 端点 | JPEG 帧缓冲 |
| `ImagingProvider` | Get/SetImagingSettings、GetOptions、Move | ISP / 镜头控制 |
| `PTZProvider` | 移动、状态、预置位 | 电机驱动（或继续用模拟器） |

```go
srv, err := server.New(config,
    server.WithDeviceInfoProvider(hwInfo),
    server.WithStreamURIProvider(rtspMounts),
    server.WithSnapshotProvider(frameBuffer),
    server.WithImagingProvider(isp),
)
```

不传选项 = 完整模拟器行为（CLI 与示例零改动）。SOAP handler 现在是
无状态的翻译层；`Server.Handle*` 签名、导出的模型类型、领域错误哨兵
保持 `server.*` 拼写（别名兼容）。

**预置位存储（v2.2.0）：** `PTZProvider` 继续负责移动/状态/GotoPreset；
可选的 `PTZPresetReader` 让 GetPresets 超越静态配置，`PTZPresetWriter`
解锁 SetPreset / RemovePreset——模拟器两者都实现了，带每 profile 可变
存储（set → list → goto → remove 开箱即用）。GetConfigurations 与
GetNodes 按档案配置以 WSDL 正确命名空间应答。

## TLS 传输（Profile T 基线）

v2.1.0 起设置 `Config.TLSCertFile` 与 `Config.TLSKeyFile`（两者同设或
同不设——只设一个会在校验时报错）后，服务端走 HTTPS。没有单独的启动
调用：TLS 在 `Start` 内部决定。`Start` 显式绑定监听，`Server.ListenAddr()`
启动即可读——`Port: 0`（内核分配）时也可观察，适合测试：

```go
config := server.DefaultConfig()
config.TLSCertFile = "cert.pem"
config.TLSKeyFile = "key.pem"

srv, _ := server.New(config)
go srv.Start(ctx)                           // https://…/onvif/device_service
fmt.Println(srv.ListenAddr())               // 如 0.0.0.0:8443
```

范围：纯 TLS 传输（Profile T 基线）。高级置备——PKI、802.1X、
UsernameToken 之外的 WS-Security——刻意不做。

可运行的 `onvif server` 二进制只说纯 HTTP；TLS 是嵌入包的面
（[onvif server CLI](onvif-go-cli-server.md)）。

## 去向

可插拔 provider（上文）直接前接真实相机栈；pull-point 事件服务
（`SupportEvents` + `PublishEvent` 接缝）见事件手册。分层见
[v2-architecture.md](onvif-go-v2-architecture.md)。
