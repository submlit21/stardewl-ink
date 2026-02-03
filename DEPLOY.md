# 部署指南

## 快速部署

### 方法1：GitHub网页上传（推荐）
1. 访问 https://github.com/submlit21
2. 点击"New repository"
3. 仓库名: `stardewl-ink`
4. **不要**初始化README、.gitignore或license
5. 创建后，点击"Upload files"
6. 上传 `stardewl-ink-complete.tar.gz`
7. GitHub会自动解压并提交

### 方法2：使用Git命令
```bash
# 解压项目
tar -xzf stardewl-ink-complete.tar.gz
cd stardewl-ink

# 初始化Git仓库
git init
git add .
git commit -m "Initial commit: Complete Stardewl-Ink project"

# 添加GitHub remote（使用SSH）
git remote add origin git@github.com:submlit21/stardewl-ink.git

# 或者使用HTTPS（需要token）
git remote add origin https://github.com/submlit21/stardewl-ink.git

# 推送到GitHub
git push -u origin main
```

## 环境配置

### Go环境设置（国内用户）
```bash
# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 下载依赖
go mod download
```

### 构建项目
```bash
# 使用Makefile
make build

# 或手动构建
./scripts/build.sh
```

## 运行测试

### 信令服务器
```bash
# 启动服务器（端口8080）
./dist/stardewl-signaling

# 验证服务器运行
curl http://localhost:8080/health
```

### 演示程序
```bash
# 运行完整演示
go run examples/simple_demo.go

# 或构建后运行
go build -o dist/stardewl-demo examples/simple_demo.go
./dist/stardewl-demo
```

## 各平台客户端开发

### Windows (WinUI 3)
1. 安装Visual Studio 2022
2. 选择"使用C++的桌面开发"和".NET桌面开发"
3. 创建WinUI 3项目
4. 使用P/Invoke调用核心库

### macOS (SwiftUI)
1. 安装Xcode
2. 创建SwiftUI项目
3. 创建C桥接文件调用核心库

### Linux (GTK 4)
1. 安装GTK 4开发包
2. 创建C项目
3. 链接核心库

## 生产部署

### 信令服务器部署
```bash
# 构建生产版本
GOOS=linux GOARCH=amd64 go build -o stardewl-signaling-linux ./signaling

# 使用systemd服务
sudo cp stardewl-signaling-linux /usr/local/bin/
sudo cp config/stardewl-signaling.service /etc/systemd/system/
sudo systemctl enable stardewl-signaling
sudo systemctl start stardewl-signaling
```

### 客户端打包
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o stardewl-client.exe ./client/windows

# macOS
GOOS=darwin GOARCH=arm64 go build -o stardewl-client-macos ./client/macos

# Linux
GOOS=linux GOARCH=amd64 go build -o stardewl-client-linux ./client/linux
```

## 监控和维护

### 日志查看
```bash
# 信令服务器日志
journalctl -u stardewl-signaling -f

# 客户端日志（如果启用）
./stardewl-client --log-level=debug
```

### 性能监控
```bash
# 查看连接数
curl http://localhost:8080/health | jq .connections

# 查看服务器状态
systemctl status stardewl-signaling
```

## 故障排除

### 常见问题

1. **Go依赖下载失败**
   ```bash
   go env -w GOPROXY=https://goproxy.cn,direct
   go env -w GOSUMDB=off
   ```

2. **信令服务器端口冲突**
   ```bash
   # 修改端口
   ./dist/stardewl-signaling -port=9090
   
   # 或停止占用进程
   sudo lsof -i :8080
   sudo kill -9 <PID>
   ```

3. **WebRTC连接失败**
   - 检查防火墙设置
   - 验证STUN服务器可访问
   - 检查NAT类型

4. **Mod扫描失败**
   ```bash
   # 手动指定Mods路径
   ./stardewl-client --mods-path="/path/to/Mods"
   ```

### 调试模式
```bash
# 启用调试日志
export STARDEWL_DEBUG=1
./dist/stardewl-signaling

# 或客户端
./stardewl-client --debug
```

## 更新和维护

### 更新依赖
```bash
# 更新所有依赖
go get -u ./...
go mod tidy

# 更新特定包
go get -u github.com/pion/webrtc/v3
```

### 版本发布
```bash
# 打标签
git tag v1.0.0
git push origin v1.0.0

# 创建发布
gh release create v1.0.0 --title "v1.0.0" --notes "初始版本"
```

## 安全建议

1. **生产环境配置**
   - 使用TLS/SSL加密
   - 限制允许的来源
   - 启用身份验证

2. **监控和告警**
   - 设置连接数监控
   - 配置错误告警
   - 定期日志分析

3. **备份策略**
   - 定期备份配置
   - 版本控制所有更改
   - 测试恢复流程

## 支持

- 问题报告: GitHub Issues
- 文档: 查看 `docs/` 目录
- 示例: 查看 `examples/` 目录