# 快速开始指南

## 1. 环境要求

- Go 1.22 或更高版本
- Git
- 星露谷物语（可选，用于测试Mod功能）

## 2. 获取代码

```bash
# 克隆仓库
git clone https://github.com/submlit21/stardewl-ink.git
cd stardewl-ink

# 设置Go代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 下载依赖
go mod download
```

## 3. 构建项目

```bash
# 使用Makefile（推荐）
make build

# 或手动构建
./scripts/build.sh
```

构建完成后，在 `dist/` 目录下会生成：
- `stardewl-signaling` - 信令服务器
- `libstardewl-core.a` - 核心库（C ABI）

## 4. 启动信令服务器

```bash
# 启动服务器（默认端口8080）
./dist/stardewl-signaling

# 或使用Makefile
make run-signaling
```

服务器启动后，可以通过以下方式验证：
- 健康检查：`curl http://localhost:8080/health`
- 创建房间：`curl -X POST http://localhost:8080/create`

## 5. 运行示例程序

```bash
# 运行演示程序
go run examples/simple_demo.go

# 或构建后运行
go build -o dist/stardewl-demo examples/simple_demo.go
./dist/stardewl-demo
```

## 6. 测试完整流程

### 步骤1：启动信令服务器
```bash
./dist/stardewl-signaling
```

### 步骤2：创建主机客户端（终端1）
```bash
cat > host.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"time"
	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	config := core.ClientConfig{
		SignalingURL: "ws://localhost:8080/ws",
		ConnectionID: "test-room-001",
		IsHost:       true,
		ICEServers:   core.GetDefaultICEServers(),
	}

	client, err := core.NewStardewlClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	client.SetModsCheckedHandler(func(comparison core.ModComparison) {
		fmt.Println("\n=== Mod对比结果 ===")
		fmt.Println(core.FormatComparisonResult(comparison))
	})

	fmt.Println("主机已启动，连接码: test-room-001")
	fmt.Println("等待客户端连接...")
	
	time.Sleep(30 * time.Second)
}
EOF

go run host.go
```

### 步骤3：创建客户端（终端2）
```bash
cat > client.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"time"
	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	config := core.ClientConfig{
		SignalingURL: "ws://localhost:8080/ws",
		ConnectionID: "test-room-001",
		IsHost:       false,
		ICEServers:   core.GetDefaultICEServers(),
	}

	client, err := core.NewStardewlClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	client.SetModsCheckedHandler(func(comparison core.ModComparison) {
		fmt.Println("\n=== Mod对比结果 ===")
		fmt.Println(core.FormatComparisonResult(comparison))
	})

	fmt.Println("客户端已启动")
	fmt.Println("连接到房间: test-room-001")
	
	time.Sleep(30 * time.Second)
}
EOF

go run client.go
```

## 7. 各平台客户端开发

### Windows (WinUI 3)
```csharp
// 使用P/Invoke调用核心库
[DllImport("libstardewl-core.a")]
public static extern IntPtr CreateClient(string config);

// 界面使用WinUI 3 XAML
```

### macOS (SwiftUI)
```swift
// 使用C桥接调用核心库
import stardewl_core

// 界面使用SwiftUI
```

### Linux (GTK 4)
```c
// 使用FFI调用核心库
#include <stardewl_core.h>

// 界面使用GTK 4
```

## 8. 配置说明

### 信令服务器配置
编辑 `config/config.yaml`：
```yaml
signaling:
  host: "0.0.0.0"
  port: 8080
  allowed_origins: ["*"]
```

### 客户端配置
```go
config := core.ClientConfig{
	SignalingURL: "ws://your-server:8080/ws",
	ConnectionID: "your-room-id",
	IsHost:       true,
	ModsPath:     "C:/Path/To/StardewValley/Mods",
	ICEServers:   core.GetDefaultICEServers(),
}
```

## 9. 故障排除

### 常见问题

1. **Go依赖下载失败**
   ```bash
   go env -w GOPROXY=https://goproxy.cn,direct
   go env -w GOSUMDB=off
   ```

2. **信令服务器启动失败**
   - 检查端口8080是否被占用
   - 检查防火墙设置

3. **WebRTC连接失败**
   - 检查ICE服务器配置
   - 检查网络连接（STUN服务器需要外网访问）

4. **Mod扫描失败**
   - 检查文件路径权限
   - 确认星露谷物语已安装

### 日志查看
```bash
# 查看信令服务器日志
./dist/stardewl-signaling 2>&1 | tee server.log

# 客户端调试
export STARDEWL_DEBUG=1
go run examples/simple_demo.go
```

## 10. 下一步

1. 阅读 [架构文档](docs/ARCHITECTURE.md)
2. 查看 [API 文档](docs/API.md)
3. 参与开发或报告问题

## 许可证

MIT License