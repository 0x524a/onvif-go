# CLI：onvif diagnostics

`onvif diagnostics` 对一台相机跑本库的全部读取面——十一个编号操作、
逐操作的成功/失败与响应耗时——并写出 JSON 报告。加 `-capture-xml`
还会把每次原始 SOAP 交互录进一个设计为回归夹具的 `tar.gz` 归档。
**提 issue 前就跑它**：JSON + 归档携带维护者需要的全部现场。

## 安装

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## 用法

```bash
onvif diagnostics -endpoint http://192.168.1.201/onvif/device_service \
    -username service -password '***' [-output ./camera-logs] \
    [-timeout 30] [-verbose] [-capture-xml] [-capture-all]
```

| 旗标 | 默认 | 含义 |
|---|---|---|
| `-endpoint` | —（必填） | 设备服务 endpoint URL |
| `-username` / `-password` | —（必填） | 相机凭证 |
| `-output` | `./camera-logs` | 输出目录 |
| `-timeout` | `30` | 单请求超时（秒） |
| `-verbose` | 关 | 逐操作的过程细节 |
| `-capture-xml` | 关 | 录制原始 SOAP 请求/响应对 → `tar.gz` |
| `-capture-all` | 关 | 全面模式：全部读操作（隐含 `-capture-xml`） |

## 它跑什么

标准流程按序执行：`GetDeviceInformation`、`GetSystemDateAndTime`、
`GetCapabilities`、服务 endpoint 发现（`Initialize`）、`GetProfiles`，
再逐 profile 跑 `GetStreamUri`、`GetSnapshotUri`、
`GetVideoEncoderConfigurations`、`GetImagingSettings`、`GetStatus`
（PTZ）、`GetPresets`（PTZ）。每个结果带 `success`、`data`、`error`、
`response_time` 落进报告；失败同时追加进全局 `errors` 日志。

输出文件按设备自身身份命名（可用时）：

```text
camera-logs/
├── Hikvision_DS-2CD2143G2-I_5.7.1_diag_20260920-1030.json
└── Hikvision_DS-2CD2143G2-I_5.7.1_xmlcapture_20260920-1030.tar.gz
```

归档内每次 SOAP 交互一个文件，外加描述设备与采集上下文的
`metadata.json`（V2 格式）。

## 场景

**提 issue。** 标准流程 + `-capture-xml` 跑一遍，抹掉 JSON 里的凭证，
两个文件都附上。报告精确指出哪个操作、以何种方式失败（fault 码、
HTTP 状态、超时）；归档让维护者能重放你相机的一字不差的线上行为。

**变成回归夹具。** 抓取归档喂给 `generate-tests`，它把交互转成对
录制响应的 Go 测试，并把相机登记进 `testdata/captures/registry.json`
——真实固件怪癖就此变成永久回归覆盖（见
[testing 手册](onvif-go-testing.md)）。

**固件对比。** 相机固件升级前后各跑一遍，diff 两份 JSON 报告，哪个
操作的响应形状或耗时变了一目了然。

**慢或抖动的相机。** 调大 `-timeout`（某些嵌入式固件跑
`GetCapabilities` 要几十秒）；`-verbose` 逐操作实时输出而非只看汇总。

## 隐私注意

JSON 含 endpoint 与用户名；XML 归档含完整 SOAP 报文——包括相机回显
的任何内容。分享前先抹凭证与内网地址（贴进 issue 的内容适用文档
中心的脱敏红线）。
