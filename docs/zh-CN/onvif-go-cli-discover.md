# CLI：discover

`discover` 用一条命令回答一个问题：*此刻这个网络上有哪些 ONVIF
设备？* 它发出多播 WS-Discovery Probe，打印每个应答者的 endpoint
reference、XAddr 与 scopes——在还不知道任何凭证之前的脚本友好起点。

## 安装

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
# 或从 release 产物直接下载预构建二进制（六平台）
```

## 用法

```bash
discover [-interface eth0] [-timeout 10s]
```

| 旗标 | 默认 | 含义 |
|---|---|---|
| `-interface` | 全部 | 探测走出的网卡——按名指定（`eth0`、`en0`） |
| `-timeout` | `10s` | 整体发现窗口（Go 时长格式：`5s`、`30s`、`1m`） |

输出按相机分块：`Endpoint`（发现层 endpoint reference）、`XAddr`
行（实际服务地址——一台设备可能通告多个）、`Scopes`（携带厂商、
型号、硬件线索的 ONVIF scope URI）。零发现以 0 退出并提示
"No cameras found."；传输层失败以 1 退出。

```text
Found 2 camera(s):

Camera 1:
  Endpoint: uuid:1b2e3d47-0000-0000-9cfa-5e0a9c4b2d1f
  XAddr: http://192.168.1.201:80/onvif/device_service
  Scopes:
    - onvif://www.onvif.org/Manufacturer/Hikvision
    - onvif://www.onvif.org/Model/DS-2CD2143G2-I
```

## 场景

**多网卡主机。** VPN + 有线 + docker 网桥并存的机器上，走错网卡的
探测要么一无所获、要么被虚拟网络应答。钉住物理网卡：

```bash
discover -interface en0 -timeout 5s
```

**资产巡检。** 固定格式的输出让它可以当巡检脚本的一行前端
（解析 `XAddr:` 行）；拿到凭证后再交给 `onvif quick` /
`onvif diagnostics`。

**慢网络。** 断电重启后的相机上线慢；与其循环跑工具不如拉长窗口——
整个超时期间应答者会持续到达：

```bash
discover -timeout 30s
```

## 边界与下一步

多播只到本子网；跨路由子网的相机请在与它们同 LAN 的主机上做定向
探测——`ProbeEndpoint` / `ProbeSerial` API 见
[discovery 手册](onvif-go-discovery.md)，那里也讲了用
`discovery.FilterONVIFDevices` 滤掉同样会应答 Probe 的
Synology/Windows/打印机幽灵应答者。
