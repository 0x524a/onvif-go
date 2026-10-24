# 并发模型

一个 `Client` 可被多个 goroutine 并发使用。本文说明该保证的依据与用法。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 契约

- **一个 `Client`、多个 goroutine、无需外部加锁。** 在录制、快照、PTZ、
  事件等组件间共享同一个 client——这是预期用法，对 ESP32 级弱设备尤其重要
  （每多一条 HTTP 连接都算数）。
- client 的全部可变状态——凭据、时钟偏差、鉴权梯队 sticky 结论、能力缓存、
  服务端点——由内部 `sync.RWMutex` 保护。
- 每个操作在共享的 `*http.Client`（自身并发安全）上构建独立的无状态 SOAP
  交换；请求之间没有任何状态留在 `Client` 上。
- 配置 setter（`SetCredentials`、`SetClockSkew`、`InvalidateCapabilitiesCache`、
  `ResetAuthLadder`）可与进行中的调用并发执行；每个调用使用派发时刻的一致
  快照。

## 审计修掉了什么

这个保证曾经只是「几乎成立」，有三处缺口（issue #12）：

1. **端点竞态** —— Media/PTZ/Imaging 的服务端点读取无锁，与 `Initialize`
   的写入竞争。
2. **裸读且无回退** —— PTZ 和 Imaging 直接读端点字段，未 Initialize 就调用
   会把请求发到空 URL。现在全部端点访问走带锁 getter，并回退到设备端点。
3. **无锁写入** —— `Initialize` 现在持有写锁。

## 如何锁死保证

`concurrency_test.go` 在一个共享 client 上跑混合操作矩阵——Device / Media /
PTZ / Events 操作与 `SetCredentials` / `SetClockSkew` /
`InvalidateCapabilitiesCache` / `Initialize` 并发——CI 的测试任务在每次推送时
以 `-race` 运行它。

## 实践建议

- **每台设备共享一个 `Client`。** 每次调用新建 client 会重复握手，可能压垮
  ESP32 级固件。
- 热路径用 `GetCapabilitiesCached` 而非 `GetCapabilities`——它对并发首调
  single-flight（见[架构](onvif-go-architecture.md)）。
- **不要**在**不同设备**间共享 `Client`；端点身份与鉴权结论都是按设备的。
- 托管事件订阅自带后台 goroutine，通过 `Unsubscribe` / 上下文取消停止
  （见[事件](onvif-go-events.md)）。
