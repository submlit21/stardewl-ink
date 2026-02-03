# 交叉编译指南

## 概述

Stardewl-Ink 支持从 Linux 环境交叉编译到 Windows、macOS 和其他 Linux 平台。这使得你可以在单一开发环境中为所有目标平台构建可执行文件。

## 快速开始

### 使用 Makefile（推荐）

```bash
# 构建所有平台
make cross-build-all

# 或构建特定平台
make cross-build-windows    # 构建 Windows .exe 文件
make cross-build-macos      # 构建 macOS 二进制文件
make cross-build-linux      # 构建 Linux 二进制文件

# 别名（更方便）
make build-windows
make build-macos
make build-linux
```

### 使用交叉编译脚本

```bash
# 显示帮助
./scripts/cross-build.sh

# 列出可用平台
./scripts/cross-build.sh list

# 构建所有平台
./scripts/cross-build.sh all

# 构建特定平台
./scripts/cross-build.sh windows
./scripts/cross-build.sh macos
./scripts/cross-build.sh linux
```

## 输出目录结构

```
dist/
├── windows/           # Windows 可执行文件
│   ├── stardewl.exe
│   ├── stardewl-signaling.exe
│   └── stardewl-demo.exe
├── macos/             # macOS 二进制文件
│   ├── stardewl
│   ├── stardewl-signaling
│   └── stardewl-demo
└── linux/             # Linux 二进制文件
    ├── stardewl
    ├── stardewl-signaling
    └── stardewl-demo
```

## 平台配置

| 平台 | GOOS | GOARCH | 文件扩展名 | 目标系统 |
|------|------|--------|------------|----------|
| Windows | `windows` | `amd64` | `.exe` | Windows 10/11 (64位) |
| macOS | `darwin` | `arm64` | 无 | macOS (Apple Silicon) |
| Linux | `linux` | `amd64` | 无 | Linux (64位) |

## 使用示例

### 为 Windows 用户构建
```bash
# 构建 Windows 版本
make cross-build-windows

# 打包为 ZIP 文件（方便分发）
cd dist/windows && zip -r stardewl-windows.zip *.exe
```

### 为 macOS 用户构建
```bash
# 构建 macOS 版本
make cross-build-macos

# 验证构建
file dist/macos/stardewl
# 应该显示: Mach-O 64-bit executable arm64
```

### 为 Linux 用户构建
```bash
# 构建 Linux 版本
make cross-build-linux

# 测试在 Linux 上运行
./dist/linux/stardewl --version
```

### 一次性构建所有平台
```bash
# 清理并构建所有平台
make clean
make cross-build-all

# 查看所有构建的文件
find dist -type f -name "stardewl*" | sort
```

## 高级用法

### 自定义构建选项
```bash
# 手动交叉编译（不使用 Makefile）
GOOS=windows GOARCH=amd64 go build -o stardewl.exe ./cmd/stardewl

# 使用 CGO（如果需要）
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o stardewl.exe ./cmd/stardewl
```

### 构建特定版本
```bash
# 设置版本信息
VERSION="1.0.0"
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$VERSION" -o stardewl.exe ./cmd/stardewl
```

### 最小化二进制大小
```bash
# 使用压缩和优化
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o stardewl.exe ./cmd/stardewl
```

## 环境要求

### 必需
- Go 1.22+（支持交叉编译）
- 基本的构建工具（make, git）

### 可选（用于 CGO）
- Windows: `x86_64-w64-mingw32-gcc`
- macOS: `osxcross` 工具链
- Linux: `gcc` 或交叉编译工具链

## 故障排除

### 常见问题

#### 1. 交叉编译失败
```bash
# 检查 Go 版本
go version

# 检查目标平台支持
go tool dist list | grep -E "(windows|darwin|linux)"
```

#### 2. Windows 构建缺少依赖
```bash
# 如果使用 CGO，需要安装 mingw-w64
sudo apt-get install -y mingw-w64

# 然后使用 CGO_ENABLED=1
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build ...
```

#### 3. macOS 构建问题
```bash
# macOS 交叉编译需要特定的 SDK
# 如果没有 macOS SDK，可以禁用 CGO
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ...
```

#### 4. 文件权限问题
```bash
# 修复执行权限
chmod +x dist/linux/stardewl
chmod +x dist/macos/stardewl

# Windows .exe 文件通常不需要特殊权限
```

### 验证构建

```bash
# 检查文件类型
file dist/windows/stardewl.exe
# 应该显示: PE32+ executable (console) x86-64

file dist/macos/stardewl
# 应该显示: Mach-O 64-bit executable arm64

file dist/linux/stardewl
# 应该显示: ELF 64-bit LSB executable, x86-64
```

## 分发建议

### Windows
- 打包为 ZIP 文件
- 包含简单的使用说明（README.txt）
- 考虑使用 Inno Setup 或 NSIS 创建安装程序

### macOS
- 创建 .dmg 磁盘映像
- 或打包为 .tar.gz 文件
- 可能需要公证（Notarization）以绕过 Gatekeeper

### Linux
- 打包为 .tar.gz 文件
- 或创建 .deb/.rpm 包
- 考虑发布到 Snap Store 或 FlatHub

## 持续集成

### GitHub Actions 示例
```yaml
name: Cross-compile
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Build for all platforms
        run: make cross-build-all
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: stardewl-binaries
          path: dist/
```

## 性能提示

1. **并行构建**：Go 构建系统本身支持并行编译
2. **缓存依赖**：使用 `go mod download` 预下载依赖
3. **增量构建**：只重新编译更改的文件
4. **最小化构建**：使用 `-trimpath` 和 `-ldflags="-s -w"` 减少二进制大小

## 相关资源

- [Go 交叉编译文档](https://go.dev/doc/install/source#environment)
- [Go 构建约束](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [Go 编译器和链接器标志](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)

## 支持

如果遇到交叉编译问题：
1. 检查 Go 版本是否支持目标平台
2. 查看具体的错误信息
3. 尝试禁用 CGO（`CGO_ENABLED=0`）
4. 在项目 Issues 中报告问题