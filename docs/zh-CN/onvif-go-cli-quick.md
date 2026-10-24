# CLI：onvif quick

`onvif quick` 是与相机初次接触的交互式五菜单伴侣：发现设备、列
网卡、连接识别、点动云台、打印全部取流/快照地址。没有旗标、没有
配置——运行后跟着提示走即可。它为在一分钟内回答"本库到底能不能
和我的相机通信"而存在。

## 安装

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## 用法

```bash
onvif quick
```

菜单循环直到退出：

| 选项 | 动作 |
|---|---|
| 1 — 发现相机 | 多播探测（可选钉住单网卡，带列表选择器）；逐台打印名称 + endpoint |
| 2 — 列出网络接口 | 每块网卡的 up/down 状态、多播能力与地址——多网卡主机上选项 1 的预检 |
| 3 — 连接相机 | 依次询问 IP / 用户名（默认 `admin`）/ 密码 → 设备信息、profile 数量、首条流地址 |
| 4 — PTZ 演示 | 连接后查 `PTZ GetStatus`，再执行一次预设动作：右/左/上/下（连续转动 2 秒后停止）或绝对回中 |
| 5 — 取流地址 | 逐 profile：RTSP 流地址、快照地址、编码 + 分辨率 |

相机 endpoint 按 `http://<ip>/onvif/device_service` 拼接——绝大多数
设备开放的传统 ONVIF 路径。

## 场景

**新相机上手。** 依序选：1（它在不在网上？）、3（凭证对不对、它是
谁？）、5（抓 RTSP 地址给 VLC 验证）。选项 5 结束时恰好给出所需
地址：流用 VLC 打开、快照用浏览器打开。

**云台链路检查。** 选项 4 一次验证整条链——鉴权、服务发现、
`GetProfiles`、PTZ `GetStatus`——并让相机真实移动。若提示
"PTZ not supported"，是当前 profile 没挂 PTZ 配置；换一个 profile
再试，别急着断定相机没有云台机构。

**发现前的网卡分诊。** VPN 重载的笔记本先跑选项 2：标着
`Multicast: No` 的网卡载不动探测。把网卡名喂给选项 1 的选择器
（或 `discover -interface`）。

## 解读结果

- 选项 3 打出 "Connected!" 意味着给定凭证下摘要/无鉴权路径成功；
  鉴权梯队行为（密码明文、HTTP Basic 回退）见
  [authentication 手册](onvif-go-authentication.md)。
- 流地址逐字来自 `GetStreamUri`——有的相机把凭证嵌进 URL，有的
  走 RTSP 鉴权；工具收尾提示两者都提到。

## 边界

一切都是一次性交互——要可重复运行、脚本化采集、或给 issue 附证据，
用 [onvif diagnostics](onvif-go-cli-diagnostics.md)。
