# 快速开始：发现并取流

`onvif-go/v2` 是 ONVIF 客户端（也含服务端）。本页给最小可编译示例：
连接一台相机、列出 profile、用启发式选出主 / 子码流。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

> v2 已稳定（v2.0.0 于 2026-09-17 发布，当前线为 v2.2.0）；生产仍固定
> 到具体版本号。

## 最小示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
	client, err := onvif.NewClient("192.0.2.10")
	if err != nil {
		log.Fatal(err)
	}
	profiles, err := client.Media().GetProfiles(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	mainToken := onvif.SelectMainProfile(profiles)
	subToken := onvif.SelectSubProfile(profiles, mainToken)
	fmt.Println("main profile:", mainToken, "sub profile:", subToken)
}
```

要点：

- `NewClient` 接受 IP 或 URL；鉴权用 `WithCredentials` 等选项按需追加
  （[鉴权与安全](onvif-go-authentication.md)）。
- 取流地址用 `client.Media().GetStreamURI`，可经 `StreamSetup` 指定
  RTSP / HTTP 等传输（[媒体与码流](onvif-go-media.md)）。
- 事件订阅一句话搞定：`client.Events().SubscribeEvents`（[事件](onvif-go-events.md)）。
- H.265/AV1 配置在 Media2 模型：`client.Media2().GetVideoEncoderConfigurationOptions`
  按支持的编码逐条上报、`Encoding` 为自由名（[媒体](onvif-go-media.md#media2-h265av1)）。
- Profile M 面：`client.Analytics()`（规则/分析模块配置）与 `metadata`
  包（解析分析输出流）。

## 下一步

- [CLI 工具](onvif-go-cli.md)：写代码前先用 `discover` / `onvif quick` /
  `onvif diagnostics` 摸一遍相机。
- [设备发现](onvif-go-discovery.md)：WS-Discovery 探测网内相机。
- [媒体与码流](onvif-go-media.md)：profile 选择启发式与流 URI。
- [架构](onvif-go-architecture.md) / [v2 架构](onvif-go-v2-architecture.md)：分层与迁移。
