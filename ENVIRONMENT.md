# 开发环境配置

## 必需环境

### Go 开发环境
- **Go 版本**: 1.22 或更高版本
- **Go 代理**（中国用户推荐）:
  ```bash
  go env -w GOPROXY=https://goproxy.cn,direct
  go env -w GOSUMDB=off
  ```

### 构建工具
- **GCC/Clang**: 用于编译（大多数系统已预装）
- **Git**: 版本控制

## 快速安装

### Ubuntu/Debian
```bash
# 安装 Go
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
rm go1.22.2.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# 安装 Git
sudo apt-get update
sudo apt-get install -y git
```

### macOS
```bash
# 使用 Homebrew 安装
brew install go git

# 或从官网下载 Go 安装包
```

### Windows
1. 从 https://go.dev/dl/ 下载 Go 安装程序
2. 从 https://git-scm.com/ 下载 Git
3. 按照安装向导完成安装

## 项目设置

### 1. 克隆项目
```bash
git clone git@github.com:submlit21/stardewl-ink.git
cd stardewl-ink
```

### 2. 下载依赖
```bash
go mod download
```

### 3. 构建项目
```bash
# 构建当前平台
make build

# 或交叉编译（从 Linux 构建 Windows 版本）
make cross-build-windows
```

### 4. 运行
```bash
# 启动信令服务器（需要先运行）
./dist/stardewl-signaling

# 运行 CLI 应用
./dist/stardewl --interactive
```

## 验证安装

```bash
# 检查版本
go version
git --version

# 验证项目构建
make test-build  # 只构建不运行测试
```

## 故障排除

### Go 模块下载失败
```bash
# 设置国内镜像
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 清理缓存
go clean -modcache
go mod download
```

### 构建失败
```bash
# 清理并重新构建
make clean
make build

# 或直接使用 Go 构建
go build -o dist/stardewl ./cmd/stardewl
```

### 端口冲突
```bash
# 信令服务器默认使用 8080 端口
# 如果端口被占用，可以修改代码或停止占用进程
sudo lsof -i :8080
sudo kill -9 <PID>
```

## 开发工作流

1. **修改代码**
2. **构建测试**: `make build`
3. **运行测试**: `./dist/stardewl --interactive`
4. **提交更改**: `git add . && git commit -m "message" && git push`

## 生产部署

### 信令服务器
```bash
# 构建生产版本
GOOS=linux GOARCH=amd64 go build -o stardewl-signaling ./signaling

# 使用 systemd 服务
sudo cp stardewl-signaling /usr/local/bin/
sudo systemctl enable stardewl-signaling
sudo systemctl start stardewl-signaling
```

### 客户端分发
```bash
# 构建所有平台
make cross-build-all

# 打包分发
cd dist/windows && zip -r stardewl-windows.zip *.exe
cd ../macos && tar -czf stardewl-macos.tar.gz *
cd ../linux && tar -czf stardewl-linux.tar.gz *
```

## 支持

- 项目文档: `docs/` 目录
- 问题报告: GitHub Issues
- 快速开始: `QUICKSTART.md`