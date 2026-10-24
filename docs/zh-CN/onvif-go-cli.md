# CLI 工具总览

`cmd/` 附带四个辅助可执行文件——发现探测、交互式快速上手工具、深度诊断采集器，以及可独立运行形态的虚拟相机模拟器。它们都是零依赖单二进制，每次发布为六个平台构建。

## 哪个工具干哪件事？

| 你想…… | 运行 |
|---|---|
| 看局域网上有哪些 ONVIF 设备（脚本友好输出） | `onvif discover` |
| 交互式摆弄一台相机——连接、云台点动、取流地址 | `onvif quick` |
| 采集一台相机的全操作报告（提 issue、留档抓包） | `onvif diagnostics` |
| 无硬件测试录像机/NVR 对接虚拟相机 | `onvif server` |
| 把抓到的 SOAP 交互变成回归测试（开发者） | `generate-tests` |

各子命令手册：[discover](onvif-go-cli-discover.md) ·
[onvif quick](onvif-go-cli-quick.md) ·
[onvif diagnostics](onvif-go-cli-diagnostics.md) ·
[onvif server](onvif-go-cli-server.md)

接触一台未知相机的典型顺序是前三个接力：`onvif discover` 找到它，
`onvif quick` 证明本库能与它通信，`onvif diagnostics -capture-xml`
记录你需要的其余一切。

## 安装

**v2.1.0** 起每个 [release](https://github.com/mickeyzzc/onvif-go/releases)
都附带六平台预构建产物（linux/amd64、linux/arm64、linux/arm、
darwin/amd64、darwin/arm64、windows/amd64）及 `SHA256SUMS`——下载即用，
无需 Go 工具链：

```bash
gh release download v2.6.0 -R mickeyzzc/onvif-go -p 'onvif_linux_amd64'
```

或按版本锚定从源码安装：

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0   # discover/quick/diagnostics/server 全在一个二进制
```

或在检出目录构建（`make build` 会连开发者工具 `generate-tests`
一起构建；它不进发布产物）。

## 共同约定

- 退出码 0 = 成功；非 0 且 stderr 有信息 = 失败。
- 相机凭证从不读配置文件——只认命令行旗标（`onvif server` 另支持
  `ONVIF_SERVER_PASSWORD` 环境变量），密钥不落 dotfile。
- 所见即库之所为：这些工具只是公开 API 的薄壳，没有工具侧特例——
  在工具里观察到的行为，在你自己的集成代码里同样成立。

`discover` 与 `onvif quick` 背后的发现机制见
[discovery 手册](onvif-go-discovery.md)；嵌入模拟器而非运行二进制见
[server 手册](onvif-go-server.md)。
