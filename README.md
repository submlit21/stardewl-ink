# Stardewl-Ink 🌀

星露谷物语联机工具，使用 WebRTC 实现 P2P 连接，无需端口转发。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![WebRTC](https://img.shields.io/badge/WebRTC-P2P-blue)](https://webrtc.org/)

## ✨ 功能特性

- 🚀 **WebRTC P2P 连接** - 使用连接码配对，无需端口转发或复杂配置
- 🔗 **简单配对系统** - 主客户端生成连接码，客户端输入即可连接
- 📁 **智能 Mod 检查** - 自动扫描并对比两端 Mod 文件，提示差异
- 🛠️ **真正的跨平台** - 核心使用 Go，各平台使用原生 UI 技术
- 🔒 **隐私保护** - 无账号系统，无需登录、好友或社区功能
- ⚡ **高性能** - 基于 Pion WebRTC，稳定高效的 P2P 连接

## 🏗️ 项目结构

```
stardewl-ink/
├── core/                 # 核心 WebRTC 连接库 (Go)
│   ├── connection.go    # WebRTC 连接管理
│   ├── mods.go         # Mod 文件扫描和对比
│   ├── messages.go     # 消息协议定义
│   └── core.go         # 客户端主逻辑
├── signaling/           # 信令服务器 (Go)
│   └── main.go         # WebSocket 信令服务器
├── examples/           # 示例代码
├── config/             # 配置文件
├── scripts/            # 构建脚本
├── docs/              # 文档
└── dist/              # 构建输出
```

## 🚀 快速开始

### 1. 环境准备
```bash
# 克隆项目
git clone git@github.com:submlit21/stardewl-ink.git
cd stardewl-ink

# 设置 Go 代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 下载依赖
go mod download
```

### 2. 构建项目
```bash
# 使用 Makefile
make build

# 或手动构建
./scripts/build.sh
```

### 3. 使用 CLI 应用（立即开始联机！）
```bash
# 交互模式（推荐新手）
./dist/stardewl --interactive

# 或直接使用命令行
./dist/stardewl --host          # 作为主机创建房间
./dist/stardewl --join=123456   # 作为客户端加入房间
```

### 4. 启动信令服务器（如果需要）
```bash
./dist/stardewl-signaling
# 服务器将在 http://localhost:8080 启动
```

### 5. 运行示例
```bash
# 运行演示程序
./dist/stardewl-demo
```

## 📖 详细文档

- [快速开始指南](QUICKSTART.md) - 完整的安装和使用教程
- [架构设计](docs/ARCHITECTURE.md) - 系统架构和设计原理
- [API 文档](docs/API.md) - 核心库 API 参考

## 🎮 使用流程

### 作为主机（创建游戏）
1. 启动客户端应用
2. 点击"生成连接码"
3. 将连接码分享给朋友
4. 等待客户端连接
5. 连接成功后自动检查 Mod 一致性

### 作为客户端（加入游戏）
1. 启动客户端应用
2. 输入朋友分享的连接码
3. 点击"连接"
4. 连接成功后自动检查 Mod 一致性

## 🔧 技术栈

### 核心层
- **语言**: Go 1.22+
- **WebRTC**: [Pion WebRTC](https://github.com/pion/webrtc) v3
- **网络**: WebSocket + STUN/TURN

### 信令服务器
- **框架**: 标准库 + gorilla/websocket
- **协议**: JSON over WebSocket
- **特性**: 房间管理、心跳检测、连接保活

### 客户端界面（各平台）
- **Windows**: WinUI 3 / WPF (C#)
- **macOS**: SwiftUI (Swift)
- **Linux**: GTK 4 (C) / Qt (C++)
- **通信**: C ABI 调用核心库

## 🧪 开发指南

### 运行测试
```bash
make test
```

### 代码格式化
```bash
make fmt
```

### 代码检查
```bash
make lint
```

### 构建所有目标
```bash
make clean build test
```

## 📁 Mod 支持

### 支持的 Mod 格式
- `.mod` 文件
- `.dll` 文件
- `.zip` 压缩包

### 自动路径检测
工具会自动检测以下平台的星露谷物语 Mods 路径：
- **Windows**: `%APPDATA%\StardewValley\Mods`
- **macOS**: `~/Library/Application Support/StardewValley/Mods`
- **Linux**: `~/.local/share/StardewValley/Mods`
- **Steam Deck**: Flatpak 兼容路径

## 🔗 通信协议

### 信令消息
```json
{
  "type": "offer|answer|ice_candidate",
  "data": {
    "connection_id": "房间ID",
    "sdp": "SDP描述",
    "candidate": "ICE候选"
  }
}
```

### 应用消息
```json
{
  "type": "mods_list|mods_comparison|game_ready",
  "payload": {
    "mods": [...],
    "comparison": {...}
  }
}
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Pion WebRTC](https://github.com/pion/webrtc) - 优秀的 Go WebRTC 实现
- [gorilla/websocket](https://github.com/gorilla/websocket) - Go WebSocket 库
- 星露谷物语社区 - 灵感来源

---

**Happy Farming!** 🌾