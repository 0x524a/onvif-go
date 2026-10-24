# 测试

测试套件默认完全离线：所有行为测试都跑在 `httptest` mock 设备或原始 XML
fixture 上。真机集成测试存在，但由环境变量门控，从不在 CI 运行。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## 运行

```bash
make test        # go test -race ./...
make lint        # golangci-lint run（要求零 findings）
make fmt         # 通过 golangci-lint fmt 执行 gofumpt + goimports
make check       # lint + test
```

CI（`.github/workflows/ci.yml`）在每次推送到 `main` 时运行 lint、格式检查、
race 测试与构建；三个任务都是必需状态检查。

少数发现测试需要主机能加入 WS-Discovery 组播组。没有可用组播路由的主机
（开着 VPN 的笔记本、部分 CI runner）会自动 skip——套件保持绿色。

## 测试分层

| 层 | 位置 | 覆盖 |
|---|---|---|
| mock 设备测试 | 代码旁的 `*_test.go` | 对 `httptest` SOAP 设备的完整 client 行为：鉴权梯队迁移、Fault 处理、响应形态变体、托管订阅生命周期 |
| 原始 fixture | `testdata/captures/*.xml` | 把真实形态的信封回放进 client（如 issue #3 背后的 GetStreamUri 命名空间变体） |
| 解析器单测 | 如 `TestParseScopes`、`TestLooseExtractURI`、`TestSelectMainProfile` | 纯函数、无 I/O |
| 一致性回环 | client↔simulator 套件 | 本库自己的 client 与 server 走真实 HTTP 跑全操作矩阵（device、PTZ、imaging——含 GetScopes 元素形态往返） |
| 命名空间契约套件 | wire-namespace 测试 | 每个响应面序列化后按命名空间严格反解——元素解析到 WSDL 指定命名空间才赋值——外加显式前缀端到端检查 |
| Fuzz 目标 | 原生 `FuzzXxx`（device/media/events） | 真实 per-op 响应解析器对抗恶意报文，种子取自契约线格式 |
| 并发矩阵 | `concurrency_test.go` | 单个共享 client 上的混合操作 + 配置变更，`-race` 下有实际意义 |
| 抓包回放助手 | `internal/onviftesting/` 包 | mock server、抓包注册表、golden 文件——大型套件在用 |

## 真机集成测试

部分测试文件（`device_real_camera_test.go`、`media_real_camera_test.go`）
在提供连接信息时（且仅在当时）对真实硬件运行：

```bash
export ONVIF_TEST_ENDPOINT="http://192.168.x.x/onvif/device_service"
export ONVIF_TEST_USERNAME="..."
export ONVIF_TEST_PASSWORD="..."
go test -v -run RealCamera ./...
```

文档和 issue 里一律使用占位符凭据——绝不放真实凭据。设备行为异常时，抓下
它的原始 SOAP 响应，作为 fixture 加进 `testdata/captures/`，让回归在相机
离场后依然存活。`cmd/onvif`（见 [CLI 工具](onvif-go-cli.md)）就是为产出
这些抓包而生的。

## 测试驱动开发

本库以 TDD 方式开发：**先写失败的测试再写实现**——修 bug 先写复现
用例，新服务操作先锁定线上契约（设备会分支判断的请求字段 + 需要
解析的响应字段）再实现。本轮补测就挖出四个被"看起来能跑"的代码
掩盖的真实缺陷（`imaging.Move` 从未编码焦点参数、`imaging.GetOptions`
静默丢弃大部分选项组、死端点回落守卫、`Stop` 要等满一个读超时周期
才能解除阻塞）。

约定：

- 服务操作测试统一走 `internal/testutil.FakeCaller`——与真实传输完全
  相同的 `xml.Unmarshal` 解码路径、无套接字、微秒级延迟。只有 HTTP
  行为本身是被测对象时才用 `httptest`。
- **所有 channel 等待必须有界**：`waitFor`/`mustReceive` 辅助在 5 秒
  内判失败——死锁必须表现为测试失败，绝不能变成 CI 超时。
- 后台循环（`events.EventStream`、`discovery.Listener`、
  `server/discovery.Responder`）必须**立即**感知取消，而不是等下一个
  定时周期——Stop 路径关闭底层套接字打断阻塞读；测试断言 `Done()`
  及时关闭。
- 时间靠注入而不是睡眠：生命周期测试通过选项传入短时长；时间戳
  断言用相对 `time.Now()` 渲染夹具（时区无关）。
- 时长预算：一个表驱动用例微秒级；包级全套件保持个位数秒（`events`
  里唯一的 1 秒退避恢复测试是刻意的例外）。CI 以 `-race` 跑同一套
  件——注册表和共享状态必须加锁保护（`server/soap.Handler` 是范例）。

## 新测试的约定

- 优先用 `httptest` server 而非手写 mock transport；按请求体内容分类请求，
  不按调用顺序——传输细节变化时测试仍然有效。
- 对设备会「看请求内容行事」的部分断言**报文形态**（如 `GetStreamUri`
  请求的 StreamSetup 字段）。
- 时间敏感的循环（托管订阅、发现监听器）应暴露确定性退出信号（`Done()`
  channel），不依赖 sleep。
