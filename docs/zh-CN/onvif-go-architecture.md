# 架构

onvif-go 是一个零第三方依赖的 Go 库，用于与 ONVIF 设备通信，另附一个虚拟摄像头服务器，
让你在没有实体相机的情况下测试录像软件。本文说明各组件如何协作。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go@v1.2.0
```

## 分层

```mermaid
flowchart TD
    subgraph tools["cmd/ · examples/"]
        clis["辅助 CLI 与可运行示例"]
    end

    subgraph rootpkg["根包 — Client"]
        facades["服务门面<br/>Device · Media · PTZ · Imaging<br/>Events · DeviceIO · Security"]
        dispatcher["鉴权调度器 Client.call<br/>梯队 + sticky 记忆"]
        state["共享状态（互斥锁保护）<br/>凭据 · 时钟偏差 ·<br/>能力缓存 · 服务端点"]
    end

    subgraph internals["internal/"]
        soap["internal/soap<br/>信封构建/解析 · WS-Security ·<br/>Fault 检测 · 鉴权模式"]
        digest["internal/httpdigest<br/>HTTP Digest RoundTripper<br/>（快照下载）"]
    end

    subgraph disc["discovery/ — 自包含，不依赖根包与 internal/"]
        discbox["主动探测 · 被动监听 ·<br/>定向 HTTP 探测 · 后处理"]
    end

    subgraph testbed["server/ · testing/ · testdata/captures/"]
        sim["虚拟相机 · mock server · 真机 fixture"]
    end

    clis --> facades
    facades --> dispatcher
    dispatcher --> state
    dispatcher --> soap
    dispatcher --> digest
    sim --> soap
```

## 包布局（v2）

v1 的单包门面让根目录平铺。v2（issue #20）按域拆分——`internal/api.Caller`
接口解开了此前无法拆包的导入环：

| 包 | 职责 |
|---|---|
| 根包（`onvif`） | `Client`（实现 `api.Caller`）、鉴权策略/梯队/诊断、下载、v1 兼容 alias |
| `types/` | 共享数据模型 leaf（IPAddress、IntRectangle、区间、SimpleItem、共享哨兵） |
| `device/` | tds：身份、能力（+缓存）、系统、网络、DNS/NTP、存储、WiFi |
| `security/` | 用户、访问策略、证书、IP 过滤 |
| `deviceio/` | tmd：继电器、数字 I/O、串口 |
| `media/` | trt：profile、主/子码流选择、StreamSetup、编码/音频/OSD |
| `ptz/`、`imaging/` | tptz / timg 域 |
| `events/` | tev：PullPoint 原语 + 托管订阅 |
| `discovery/` | 客户端发现（探测/监听/定向/后处理） |
| `server/` | 设备端框架（传输、handler、模拟器） |
| `internal/api` | Client 与服务之间的 `Caller` 接口 |
| `internal/soap`、`internal/httpdigest` | SOAP 传输 + WS-Security；HTTP Digest RoundTripper |

```mermaid
flowchart TD
    subgraph rootpkg["根包 — Client 实现 api.Caller"]
        client["NewClient + 鉴权梯队 + 时钟偏差"]
    end
    subgraph svcs["服务包"]
        device["device"]
        media["media"]
        others["ptz · imaging · events · deviceio · security"]
    end
    apileaf["internal/api — Caller 接口"]
    subgraph leaves["共享 leaf"]
        types["types/"]
        soap["internal/soap · internal/httpdigest"]
    end
    client --> apileaf
    device --> apileaf
    media --> apileaf
    others --> apileaf
    device --> types
    media --> types
    client --> soap
    media --> ptz
```

v1 风格的调用面通过根包 alias 继续可编译（`type Profile = media.Profile`、
门面访问器不变）；迁移映射见 [v2-architecture.md](onvif-go-v2-architecture.md)。


## 服务门面模型

`Client` 是「连接 + 配置」的载体；每个 ONVIF 操作挂在其所属的服务对象上，
与 ONVIF 服务模型一一对应：

```go
client.Device().GetDeviceInformation(ctx)
client.Media().GetProfiles(ctx)
client.PTZ().ContinuousMove(ctx, token, speed, timeout)
```

门面是无状态视图：访问器按需构造，因此单个 `Client` 可以安全地跨 goroutine
共享，门面层无需加锁（见 [并发模型](onvif-go-concurrency.md)）。

## 单一调用路径

全部约 220 个服务操作都经过同一个调度器 `Client.call`，它负责：

1. 读锁快照当前鉴权配置（主模式、回退梯队、sticky 结论）；
2. 为所选模式构建无状态 SOAP 交换；
3. 遇鉴权类失败时按梯队重试，并记住第一个可用的模式；
4. 包装鉴权类失败，保证 `errors.Is(err, ErrUnauthorized)` 恒成立。

单一可审计路径也让并发契约可证明——梯队语义详见[鉴权与安全](onvif-go-authentication.md)。

## SOAP 传输层（internal/soap）

每次调用构建一个无状态 client，职责：

- 按配置的模式构建信封与 WS-Security 头（digest、password-text 或不加），
  或改用 HTTP Basic；
- **无论 HTTP 状态如何都检测 SOAP Fault**——ONVIF 设备经常以 HTTP 200 携带
  Fault，在引入 Fault 检测之前，这类响应在无返回值操作上会被误判为成功；
- 返回结构化错误：`*FaultError`（code/subcode/reason，兼容 SOAP 1.1/1.2
  两种布局、任意命名空间前缀）与 `*HTTPStatusError`。

## 能力 XAddr 修复

相机在 `GetCapabilities` 里宣传的服务端点（XAddr）经常带着回环地址，或
漫游换 IP 后的陈旧地址。由于客户端是通过 device_service URL 连上设备的，
该 URL 的主机名才是权威依据：`Initialize` 时会把不一致的 XAddr 主机改写
（保留服务专属端口）。

## 为什么 discovery 不依赖根包

`discovery` 包刻意自包含：自带最小化的 SOAP 探测（HTTP POST + 正则/XML
提取），只需要发现功能的消费者不必引入完整 client，也永远不可能与根包
形成导入环。
