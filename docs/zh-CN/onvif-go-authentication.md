# 鉴权与安全

真实相机固件对鉴权的接受度千差万别——同一台设备的不同服务、甚至不同固件版本
都可能不一样。本文说明鉴权模式、回退梯队、时钟偏差处理与诊断 API。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 鉴权模式

```go
type AuthMode string

const (
    AuthDigest       AuthMode = "digest"        // WS-Security PasswordDigest（默认）
    AuthPasswordText AuthMode = "password-text" // UsernameToken 明文密码
    AuthHTTPBasic    AuthMode = "http-basic"    // HTTP Basic，无 WS-Security
    AuthNone         AuthMode = "none"          // 完全不鉴权
)
```

用 `WithAuthMode` 选择。默认 `AuthDigest` 与历史行为完全一致。

各模式的实际来源：

| 设备行为 | 可用模式 |
|---|---|
| 标准 ONVIF 固件 | `AuthDigest` |
| 特定服务（PTZ、GetUsers）拒绝 digest 但接受明文 token | `AuthPasswordText` |
| imaging 服务拒绝 WS-Security 但接受 HTTP Basic（因固件而异） | `AuthHTTPBasic` |
| ESP32 级极简固件，拒绝一切带鉴权的请求 | `AuthNone` |

## 回退梯队

无法预知设备接受什么时，让 client 自己一次性摸清：

```go
client, _ := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

遇鉴权类失败（分类见下）自动尝试下一模式；**第一个可用的模式会被记住**
（sticky），后续调用直达，不必每次付出全梯队代价。更换凭据
（`SetCredentials`）会清除记忆——结论只对旧凭据有效。`ResetAuthLadder()`
手动清除（设备侧变更后）；`AuthLadderMode()` 返回当前生效模式。

非鉴权错误绝不会换模式重试：网络故障与鉴权方案无关，盲目重试只增加延迟。

## 鉴权失败分类

`errors.Is(err, onvif.ErrUnauthorized)` 匹配一切「鉴权形态」的失败：

- HTTP 401 / 403；
- 携带 NotAuthorized code 的 SOAP Fault（如 `ter:NotAuthorized`）；
- **HTTP 200 却带 Fault**——ONVIF 设备的常见怪癖；
- 回退梯队全部耗尽。

```go
_, err := client.Device().GetDeviceInformation(ctx)
if errors.Is(err, onvif.ErrUnauthorized) {
    // 凭据问题——不是网络 bug，也不是解析 bug
}
```

## 时钟偏差（海康陷阱）

WS-Security digest 内嵌 `Created` 时间戳，设备会拒绝超出重放窗口（通常
±5 分钟）的 digest。相机时钟与你偏离时，**每个 digest 都像被篡改过**，
返回一句极具误导性的 "sender not authorized"——这是最会骗人的鉴权失败。

两个 API 对付它：

```go
// 选项：Initialize 时测量一次（在第一个带鉴权的调用之前）并静默应用；
// 测量失败则退回本机时钟。
client, _ := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAutoClockSkew())

// 手动：RTT 补偿测量（本地基准取往返中点，网络延迟不污染测量值）。
skew, err := client.MeasureClockSkew(ctx)
client.SetClockSkew(skew)
```

## DiagnoseAuth——三种根因分诊

鉴权「就是不通」时，根因有三种完全不同的可能：时钟偏差（修相机的 NTP）、
凭据错误、设备根本拒绝 ONVIF 鉴权。`DiagnoseAuth` 把它们分开：

```go
diag, err := client.DiagnoseAuth(ctx)
switch diag.Status {
case onvif.AuthStatusOK:             // digest 正常
case onvif.AuthStatusClockSkew:      // 用设备时间就能通——修 NTP，
                                     // 或保持 WithAutoClockSkew
    fmt.Println(diag.ClockSkew)      // 测得的偏差
case onvif.AuthStatusBadCredentials: // 用设备时间也失败
}
```

流程：探测 digest → 失败则测偏差 → 偏差超 2 分钟就用设备时间重试：成功即
坐实时钟偏差，失败则指向凭据。
