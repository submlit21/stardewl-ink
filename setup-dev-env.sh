#!/bin/bash

# Stardewl-Ink 开发环境配置脚本
# 适用于 Ubuntu 24.04 / Debian 12+

set -e

echo "🔧 配置 Stardewl-Ink 开发环境..."
echo "=========================================="

# 检查系统
if [ "$EUID" -eq 0 ]; then
    echo "❌ 请不要使用 root 用户运行此脚本"
    exit 1
fi

# 更新系统
echo "📦 更新系统包..."
sudo apt-get update

# 1. 安装 GCC 13.3
echo "🔧 安装 GCC 13.3..."
sudo apt-get install -y gcc-13 g++-13
sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-13 100
sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-13 100

# 2. 安装 Java JDK 21
echo "☕ 安装 Java JDK 21..."
sudo apt-get install -y openjdk-21-jdk
echo 'export JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64' >> ~/.bashrc
echo 'export PATH=$JAVA_HOME/bin:$PATH' >> ~/.bashrc

# 3. 安装 .NET 9.0
echo "🌐 安装 .NET 9.0..."
wget -q https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
sudo dpkg -i packages-microsoft-prod.deb
rm packages-microsoft-prod.deb
sudo apt-get update
sudo apt-get install -y dotnet-sdk-9.0

# 4. 安装 Maven 3.9
echo "📚 安装 Maven 3.9..."
sudo apt-get install -y maven
echo 'export MAVEN_HOME=/usr/share/maven' >> ~/.bashrc
echo 'export PATH=$MAVEN_HOME/bin:$PATH' >> ~/.bashrc

# 5. 安装 Go (如果还没有)
if ! command -v go &> /dev/null; then
    echo "🚀 安装 Go..."
    wget -q https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
    rm go1.22.2.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
fi

# 6. 安装 Git
echo "📝 安装 Git..."
sudo apt-get install -y git

# 7. 安装其他开发工具
echo "🛠️  安装其他开发工具..."
sudo apt-get install -y \
    build-essential \
    pkg-config \
    cmake \
    curl \
    wget \
    unzip \
    tree \
    htop \
    net-tools

# 8. 配置 Go 代理（中国用户）
echo "🌍 配置 Go 代理..."
mkdir -p ~/.config/go
echo "GOPROXY=https://goproxy.cn,direct" > ~/.config/go/env
echo "GOSUMDB=off" >> ~/.config/go/env

# 9. 验证安装
echo "✅ 验证安装..."
source ~/.bashrc

echo ""
echo "📊 安装结果："
echo "------------------------------------------"
gcc --version | head -1
java --version 2>/dev/null | head -1 || echo "Java: 未安装"
dotnet --version 2>/dev/null || echo ".NET: 未安装"
mvn --version 2>/dev/null | head -1 || echo "Maven: 未安装"
go version 2>/dev/null || echo "Go: 未安装"
echo "------------------------------------------"

echo ""
echo "🎉 开发环境配置完成！"
echo ""
echo "下一步："
echo "1. 重新打开终端或运行: source ~/.bashrc"
echo "2. 克隆项目: git clone git@github.com:submlit21/stardewl-ink.git"
echo "3. 进入项目: cd stardewl-ink"
echo "4. 构建项目: make build"
echo "5. 运行CLI: ./dist/stardewl --interactive"
echo ""
echo "💡 提示：如果某些包下载失败，请检查网络连接或手动下载。"