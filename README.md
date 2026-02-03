# Stardewl-Ink

星露谷物语联机工具，使用WebRTC实现P2P连接，无需端口转发。

## 功能特性

- 🚀 **WebRTC P2P连接**：使用连接码配对，无需端口转发
- 🔗 **简单配对**：主客户端生成连接码，客户端输入连接码即可连接
- 📁 **Mod检查**：自动扫描并对比两端Mod文件
- 🛠️ **跨平台支持**：核心使用Go，各平台使用原生UI技术
- 🔒 **无账号系统**：无需登录、好友或社区功能

## 架构设计

```
stardewl-ink/
├── core/          # 核心WebRTC连接库 (Go)
├── signaling/     # 信令服务器 (Go)
├── client/        # 各平台客户端
│   ├── windows/   # Windows客户端
│   ├── macos/     # macOS客户端
│   └── linux/     # Linux客户端
├── scripts/       # 构建脚本
└── config/        # 配置文件
```

## 快速开始

### 1. 启动信令服务器
```bash
cd signaling
go run main.go
```

### 2. 使用客户端
1. 主客户端：点击"生成连接码"
2. 客户端：输入连接码并连接
3. 连接成功后自动检查Mod一致性

## 技术栈

- **核心**: Go + Pion WebRTC
- **信令**: Go + WebSocket
- **Windows**: WinUI 3 / WPF
- **macOS**: SwiftUI
- **Linux**: GTK 4 / Qt

## 开发

```bash
# 初始化项目
make init

# 构建核心库
make build-core

# 构建信令服务器
make build-signaling

# 运行测试
make test
```

## 许可证

MIT License