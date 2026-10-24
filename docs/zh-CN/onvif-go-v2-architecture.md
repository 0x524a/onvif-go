# v2 架构（破坏性版本）

> 状态：已实施——M1–M4 已落地（#21 #22 #23 #24），收尾见 [#25](https://github.com/mickeyzzc/onvif-go/issues/25)，由 [v2 总跟踪](https://github.com/mickeyzzc/onvif-go/issues/20)追踪。本文档是唯一权威计划。

v2 是一次刻意为之的破坏性版本，两大动因：

1. **客户端规模**：约 2.4 万行代码被「门面 API 必须同包」的结构约束锁死在根目录平铺。
2. **服务端雄心**：#15–#19 要求 `server/` 从测试模拟器升级为**可嵌入真实相机的设备端框架**（MiBee Eye / rpi3b-cam）。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 模块版本策略

Go modules 强制 v2+ 使用 `/v2` 后缀：模块路径变为
`github.com/mickeyzzc/onvif-go/v2`。客户端包位于 `onvif/` 子目录
（import 路径 `github.com/mickeyzzc/onvif-go/v2/onvif`，包标识符不变），
仓库根目录不再放置任何 Go 源码。v1 各 tag 原样保留给现有消费者。
在 1.x 里做破坏会耗尽公开库的 semver 信用——不做。

## v2 解开的死结：Caller 接口

v1 无法拆包的根因是导入环：门面访问器（`Client.Media()`）需要服务类型，
服务方法又需要 `Client`。v2 采用 aws-sdk-go-v2 的标准形态：

```go
// internal/api —— 极小 leaf 包
type Caller interface {
    Call(ctx context.Context, endpoint, action string, request, response any) error
    EndpointFor(svc Service) string // 已解析的服务端点（带回退）
}
```

- `Client`（`onvif` 包）实现 `Caller`——鉴权梯队、时钟偏差、能力缓存留在这一层
- 每个服务包只依赖 `internal/api`
- `onvif` 包 import 服务包并提供访问器——单向依赖，无环
- 服务为**长驻实例**：Client 构造时创建，访问器返回同一指针，服务因此可以持有自身状态（能力缓存随 `device.Service`）

## 目标布局

```text
github.com/mickeyzzc/onvif-go/v2
├── onvif/               客户端包（包标识符仍为 `onvif`）
│   ├── client.go        NewClient + Client（实现 api.Caller）+ 选项
│   ├── auth.go          鉴权模式 / 梯队 / 时钟偏差 / DiagnoseAuth
│   ├── errors.go        跨域哨兵（自 leaf 包 alias）
│   └── aliases.go       v1 兼容再导出
├── types/               共享数据模型 leaf（IPAddress、IntRectangle 等）
├── device/              tds：身份、能力、系统、网络、DNS/NTP、存储、WiFi
│                        ——能力缓存随本包
├── security/            用户、远程用户、访问策略、证书
├── media/               trt：profile、主/子码流选择、StreamSetup、
│                        编码/音频/OSD 配置
├── ptz/   imaging/   events/（含托管订阅）   deviceio/
├── media2/              tr2：编解码无关的 Media2 配置模型——按编码的
│                        options、自由名 Encoding（H.265/AV1 的答案）、
│                        逐字透传的 set（v2.1.0）
├── analytics/           tan：Profile M 规则/分析模块配置客户端（v2.1.0）
├── metadata/            tt:MetadataStream 解析器——Profile M 分析输出
│                        通道（帧、目标、几何）（v2.1.0）
├── discovery/           客户端发现（探测/监听/定向/后处理）
├── wsdiscovery/         WS-Discovery 报文编解码 leaf——客户端与
│                        设备端应答器共享（#15）
├── server/              设备端框架（一等公民）
│   ├── server.go        生命周期、HTTP 监听、可选发现应答器挂载
│   ├── provider.go      可插拔状态后端（#19）：DeviceInfo / StreamURI /
│   │                    Snapshot JPEG / Imaging（范围+枚举校验）/ PTZ
│   ├── simulator/       现有模拟器状态实现 → 默认 provider
│   ├── services/        无状态 handler：SOAP ↔ provider 调用翻译
│   ├── soap/            设备端传输：dispatch、请求上下文（#17）、
│   │                    按动作鉴权 + PasswordText（#16）、
│   │                    XML 写出器：规范大小写、显式前缀、
│   │                    raw bytes 通道（#18）
│   └── discovery/       设备侧 WS-Discovery 应答器（#15）
├── internal/{api,soap,httpdigest}
└── cmd/  examples/  testing/  testdata/captures/  docs/{en,zh}
```

## 服务端框架原则

- **services 无状态**：`server/services/*` 不持有状态，只做 SOAP 操作到
  provider 调用的翻译；一切状态决策退到接口之后
- **模拟器即默认**：现有模拟器行为变成 `server/simulator`，未注入
  provider 时自动选用——`cmd/onvif` 与示例零改动
- **传输层管横切关注**：请求上下文（#17）、按动作鉴权策略（#16）、
  XML 字节级控制（#18）全部落在 `server/soap`，一次实现全服务受益
- **报文可预期**：写出器按 WSDL 规范大小写输出、可选显式命名空间前缀、
  提供绕过 encoding/xml 的 raw 通道——golden fixture 测试锁定

## 兼容策略

允许破坏，但不浪费破坏：

- 根包 **type alias**（`type Profile = media.Profile`）让多数消费者换掉
  import 路径即可编译
- 门面调用形态不变：`client.Media().GetProfiles(ctx)`
- 一切迁移写入 `MIGRATION.md` 完整映射表（如
  `onvif.SelectMainProfile` → `media.SelectMain`）
- 域错误随包走（`media.ErrEmptyMediaURI`、`events.ErrEventsNotSupported`），
  跨域哨兵在根包保留 alias

## 里程碑

| 阶段 | Issue | 范围 |
|---|---|---|
| M1 | #21 | ✅ 客户端拆包、`/v2` 路径、alias——零行为变化 |
| M2 | #22 | ✅ `server/soap` 传输层：上下文、鉴权策略、XML 写出器（#17 #16 #18） |
| M3 | #23 | ✅ provider 接口 + 模拟器抽离（#19） |
| M4 | #24 | ✅ 设备侧发现应答器 + 共享编解码（#15） |
| M5 | #25 | ✅ 文档、迁移指南、CHANGELOG、`v2.0.0-rc1`、MiBeeNvr 参考迁移 |

## 流程规则

- 一切变更走 PR（main 已锁：必需检查、PR-only、管理员同权、线性历史）
- 每个 PR 同步更新 `docs/{en,zh}` 与 CHANGELOG
