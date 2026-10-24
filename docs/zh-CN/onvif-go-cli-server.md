# CLI：onvif server

`onvif server` 是 `server/` 包的可运行形态：一台虚拟多目 ONVIF
相机。启动它，你的录像机、NVR 或 ONVIF 客户端就看到 1–10 个可配置的
相机 profile——带 PTZ 节点、成像服务、可选 pull-point 事件——完全
无需硬件。请按地址添加：该二进制只讲 SOAP over HTTP，不响应
WS-Discovery 多播探测。库自身的一致性回环测试驱动的就是同一个模拟器。

## 安装

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## 用法

```bash
onvif server -password 'test-pass' [-port 8080] [-profiles 3] \
    [-manufacturer onvif-go] [-model 'Virtual Multi-Lens Camera'] \
    [-ptz=true] [-imaging=true] [-events=false] [-info] [-version]
```

| 旗标 | 默认 | 含义 |
|---|---|---|
| `-host` / `-port` | `0.0.0.0` / `8080` | 监听地址 |
| `-username` / `-password` | `admin` / **必填** | 设备凭证 |
| `-manufacturer` `-model` `-firmware` `-serial` | `onvif-go` / `Virtual Multi-Lens Camera` / `1.0.0` / `SN-12345678` | 通告的身份 |
| `-profiles` | `3` | profile 数量，1–10 |
| `-ptz` `-imaging` `-events` | 开 / 开 / 关 | 服务开关 |
| `-info` | — | 打印解析后的配置并退出 |
| `-version` | — | 打印版本并退出 |

密码必填——旗标或 `ONVIF_SERVER_PASSWORD` 环境变量。模拟器拒绝空
凭证启动：空凭证是嵌入包文档化的全开放模式，不适合真网络上的演示
二进制。

profile 由十个轮换模板生成（主 1080p、广角 720p、长焦、低照度、
4K UHD、紧凑 VGA、PTZ 球机、鱼眼、热成像、车牌 60fps），各带
H264 编码器与快照配置；带 PTZ 的模板获得 pan/tilt/zoom 节点与两个
预置位。服务树挂载在给定端口的 `/onvif` 下。`SIGINT`/`SIGTERM`
优雅停机。

## 场景

**无硬件的录像机接入。** 开发并演示录像机的按地址添加相机流程——
端点即 `http://<host>:<port>/onvif/device_service`，
`GetDeviceInformation` 返回你设定的身份、
profile 与流地址照常应答——相机上桌之前全链路可跑：

```bash
onvif server -password demo -manufacturer TestCam -model VC-1 -serial SIM-0001
```

**单进程模拟一排相机。** `-profiles 10` 得到十台各异的虚拟相机
（模板轮换）——对外是一个设备地址、`GetProfiles` 展开一排；适合
UI 布局与主/子流选择测试。

**PTZ 界面测试。** 开 `-ptz` 后 PTZ 服务应答 `GetStatus`、连续/绝对
转动与预置位——足够驱动摇杆组件或预置位巡航逻辑，且永不磨损。

**事件集成。** `-events` 启用 pull-point 服务；把订阅代码指向模拟器
即可离线开发 MotionAlarm 处理（托管客户端 API 见
[events 手册](onvif-go-events.md)）。

**CI 浸润靶。** 二进制单进程、磁盘零状态，受测录像机可随意增删它；
放进集成套件旁的容器就是一个确定性 ONVIF 对端。

## TLS 与嵌入

可运行二进制只说纯 HTTP。要 HTTPS 用嵌入包的
`Config.TLSCertFile`/`TLSKeyFile`（TLS 在 `Start` 内自动启用）——见
[server 手册](onvif-go-server.md)，那里还有可插拔状态提供者、按
动作鉴权策略，以及二进制未暴露的 `PublishEvent` 接缝。
