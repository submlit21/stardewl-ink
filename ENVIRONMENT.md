# 开发环境配置

## 当前环境状态

### 已安装
- ✅ **GCC 13.3** - 已安装并配置为默认编译器
- ✅ **Go 1.22** - 已安装（项目要求）
- ✅ **Git** - 已安装

### 需要安装
- ⏳ **Java JDK 21** - 安装中/需要安装
- ⏳ **.NET 9.0** - 需要安装
- ⏳ **Maven 3.9** - 需要安装

## 安装方法

### 方法1：使用提供的脚本（推荐）
```bash
# 给予执行权限
chmod +x setup-dev-env.sh

# 运行配置脚本
./setup-dev-env.sh

# 重新加载环境变量
source ~/.bashrc
```

### 方法2：手动安装

#### 1. Java JDK 21
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y openjdk-21-jdk

# 验证
java --version
```

#### 2. .NET 9.0
```bash
# 添加Microsoft仓库
wget https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb
sudo dpkg -i packages-microsoft-prod.deb
rm packages-microsoft-prod.deb

# 安装.NET SDK
sudo apt-get update
sudo apt-get install -y dotnet-sdk-9.0

# 验证
dotnet --version
```

#### 3. Maven 3.9
```bash
# Ubuntu/Debian
sudo apt-get install -y maven

# 验证
mvn --version
```

## 环境变量配置

将以下内容添加到 `~/.bashrc` 或 `~/.zshrc`：

```bash
# Java
export JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64
export PATH=$JAVA_HOME/bin:$PATH

# Go (如果使用脚本安装)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Go 代理（中国用户）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off

# Maven
export MAVEN_HOME=/usr/share/maven
export PATH=$MAVEN_HOME/bin:$PATH
```

## 项目特定要求

### Go 版本
- **最低**: Go 1.22
- **推荐**: Go 1.22.2+
- **已验证**: Go 1.22 (当前环境), Go 1.25.5 (你的环境)

### 构建要求
```bash
# 验证构建环境
cd stardewl-ink
make build

# 如果遇到权限问题
chmod +x scripts/build.sh
```

### 网络要求
- **GitHub访问**: 需要访问 `github.com` 拉取代码
- **Go模块**: 需要访问 `proxy.golang.org` 或配置镜像
- **.NET NuGet**: 需要访问 `nuget.org`

## 故障排除

### 常见问题

#### 1. Go模块下载失败
```bash
# 设置国内镜像
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 清理并重试
go clean -modcache
go mod download
```

#### 2. 权限问题
```bash
# 修复脚本权限
chmod +x scripts/*.sh

# 修复构建输出权限
sudo chown -R $(whoami) dist/
```

#### 3. 依赖冲突
```bash
# 更新所有依赖
go get -u ./...
go mod tidy

# 清理构建缓存
make clean
```

#### 4. 端口冲突（信令服务器）
```bash
# 检查端口占用
sudo lsof -i :8080

# 使用不同端口
./dist/stardewl-signaling -port=9090
```

## 验证安装

运行验证脚本：
```bash
# 检查所有组件
./setup-dev-env.sh --check

# 或手动检查
gcc --version
java --version
dotnet --version
mvn --version
go version
git --version
```

## 下一步

环境配置完成后：
1. 克隆项目：`git clone git@github.com:submlit21/stardewl-ink.git`
2. 进入目录：`cd stardewl-ink`
3. 构建项目：`make build`
4. 测试运行：`./dist/stardewl --interactive`

## 支持

如果遇到安装问题：
1. 检查网络连接
2. 查看具体错误信息
3. 尝试手动安装单个组件
4. 联系维护者获取帮助