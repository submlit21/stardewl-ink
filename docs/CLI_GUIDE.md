# CLI 使用指南

## 概述

Stardewl-Ink 提供了一个功能完整的命令行界面，让你可以快速开始使用星露谷联机工具，无需等待GUI开发完成。

## 快速开始

### 1. 构建CLI应用
```bash
# 克隆项目
git clone git@github.com:submlit21/stardewl-ink.git
cd stardewl-ink

# 设置Go代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 下载依赖
go mod download

# 构建
make build
```

### 2. 查看帮助
```bash
./dist/stardewl --help
```

## 使用模式

### 交互模式（推荐新手）
```bash
./dist/stardewl --interactive
```

交互模式提供菜单驱动的界面：
```
🎮 星露谷联机工具 - 交互模式
========================================

请选择模式:
1. 作为主机运行（创建房间）
2. 作为客户端运行（加入房间）
3. 检查本地Mods
4. 启动信令服务器
5. 退出
```

### 命令行模式

#### 作为主机运行
```bash
# 基本用法
./dist/stardewl --host

# 指定Mods路径
./dist/stardewl --host --mods="/path/to/StardewValley/Mods"

# 使用自定义信令服务器
./dist/stardewl --host --signaling="ws://example.com:8080/ws"

# 显示详细日志
./dist/stardewl --host --verbose
```

#### 作为客户端运行
```bash
# 加入房间
./dist/stardewl --join=123456

# 指定连接码和Mods路径
./dist/stardewl --join=123456 --mods="~/Games/StardewValley/Mods"
```

#### 检查Mods
```bash
# 列出本地Mods
./dist/stardewl --list-mods --mods="/path/to/Mods"

# 只检查不连接
./dist/stardewl --check-only --mods="/path/to/Mods"
```

## 完整工作流程

### 步骤1：启动信令服务器
```bash
# 终端1 - 启动信令服务器
./dist/stardewl-signaling
```

### 步骤2：主机创建房间
```bash
# 终端2 - 作为主机运行
./dist/stardewl --host --mods="~/StardewValley/Mods"

# 输出示例：
# === 主机模式 ===
# 连接码: 784532
# 等待客户端连接...
```

### 步骤3：客户端加入房间
```bash
# 终端3 - 作为客户端运行
./dist/stardewl --join=784532 --mods="~/StardewValley/Mods"

# 输出示例：
# === 客户端模式 ===
# 连接码: 784532
# 正在连接到主机...
```

### 步骤4：Mods自动对比
连接成功后，双方会自动交换Mod信息并显示对比结果：
```
==================================================
Mods对比结果:
只在本地存在的Mod:
  - ExpandedPreconditionsUtility (a1b2c3d4, 1048576 bytes)

版本不同的Mod:
  - StardewValleyExpanded:
    本地: e5f6g7h8 (20971520 bytes)
    远程: i9j0k1l2 (21045248 bytes)

相同的Mod (15个):
  - ContentPatcher
  - JsonAssets
  - SpaceCore
  - ...（更多）

==================================================
⚠️  发现Mod差异！
请确保双方Mod一致后再开始游戏。
```

## 高级功能

### 自动路径检测
如果不指定 `--mods` 参数，工具会自动检测星露谷Mods路径：
- Windows: `%APPDATA%\StardewValley\Mods`
- macOS: `~/Library/Application Support/StardewValley/Mods`
- Linux: `~/.local/share/StardewValley/Mods`

### 心跳检测
连接建立后会自动进行心跳检测，保持连接活跃。

### 错误处理
- 网络断开自动重连
- Mods扫描失败友好提示
- 详细的错误日志（使用 `--verbose`）

## 实用命令示例

### 批量检查多个路径
```bash
# 检查多个可能的Mods路径
for path in \
  "$HOME/StardewValley/Mods" \
  "$HOME/.local/share/StardewValley/Mods" \
  "/mnt/c/Users/$(whoami)/AppData/Roaming/StardewValley/Mods"
do
  if [ -d "$path" ]; then
    echo "=== 检查 $path ==="
    ./dist/stardewl --list-mods --mods="$path"
    echo
  fi
done
```

### 生成Mods报告
```bash
# 生成详细的Mods报告
./dist/stardewl --list-mods --mods="/path/to/Mods" > mods-report.txt

# 生成JSON格式报告（需要jq）
./dist/stardewl --list-mods --mods="/path/to/Mods" --verbose 2>&1 | \
  grep -A 1000 "找到.*个Mod文件" | \
  tail -n +2 > mods-list.txt
```

### 自动化脚本
```bash
#!/bin/bash
# auto-connect.sh - 自动连接脚本

CONNECTION_CODE="$1"
MODS_PATH="$2"

if [ -z "$CONNECTION_CODE" ]; then
  # 作为主机运行
  echo "作为主机运行..."
  ./dist/stardewl --host --mods="$MODS_PATH"
else
  # 作为客户端运行
  echo "连接到 $CONNECTION_CODE..."
  ./dist/stardewl --join="$CONNECTION_CODE" --mods="$MODS_PATH"
fi
```

## 故障排除

### 常见问题

#### 1. "未检测到星露谷Mods路径"
```bash
# 手动指定路径
./dist/stardewl --host --mods="/完整/路径/到/StardewValley/Mods"
```

#### 2. 信令服务器连接失败
```bash
# 检查服务器是否运行
curl http://localhost:8080/health

# 使用不同的端口
./dist/stardewl-signaling -port=9090
./dist/stardewl --host --signaling="ws://localhost:9090/ws"
```

#### 3. WebRTC连接失败
```bash
# 启用详细日志
./dist/stardewl --host --verbose

# 检查防火墙设置
# 确保STUN服务器可访问
```

### 调试模式
```bash
# 启用所有调试信息
./dist/stardewl --host --verbose 2>&1 | tee debug.log
```

## 与GUI版本的关系

CLI版本和未来的GUI版本共享相同的核心库：
- **相同的核心功能**：WebRTC连接、Mods对比
- **相同的信令协议**：兼容相同的信令服务器
- **可并行使用**：CLI和GUI可以同时连接

## 下一步

CLI版本已经足够用于：
- ✅ 测试核心连接功能
- ✅ 验证Mods对比算法
- ✅ 实际游戏联机测试
- ✅ 自动化脚本集成

未来可以添加：
- 文件传输功能（Mods同步）
- 语音聊天支持
- 游戏状态监控
- 更丰富的统计信息