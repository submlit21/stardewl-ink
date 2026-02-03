#!/bin/bash

# Stardewl-Ink 开发环境配置脚本
# 只包含项目必需的环境

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

# 1. 安装 Go
if ! command -v go &> /dev/null; then
    echo "🚀 安装 Go..."
    wget -q https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
    rm go1.22.2.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
else
    echo "✅ Go 已安装: $(go version)"
fi

# 2. 安装 Git
if ! command -v git &> /dev/null; then
    echo "📝 安装 Git..."
    sudo apt-get install -y git
else
    echo "✅ Git 已安装: $(git --version)"
fi

# 3. 安装构建工具
echo "🛠️  安装构建工具..."
sudo apt-get install -y \
    build-essential \
    pkg-config \
    curl \
    wget \
    unzip

# 4. 配置 Go 代理（中国用户）
echo "🌍 配置 Go 代理..."
mkdir -p ~/.config/go
echo "GOPROXY=https://goproxy.cn,direct" > ~/.config/go/env
echo "GOSUMDB=off" >> ~/.config/go/env

# 5. 验证安装
echo "✅ 验证安装..."
source ~/.bashrc 2>/dev/null || true

echo ""
echo "📊 安装结果："
echo "------------------------------------------"
go version 2>/dev/null || echo "Go: 未安装"
git --version 2>/dev/null | head -1 || echo "Git: 未安装"
echo "------------------------------------------"

echo ""
echo "🎉 开发环境配置完成！"
echo ""
echo "下一步："
echo "1. 重新打开终端或运行: source ~/.bashrc"
echo "2. 克隆项目: git clone git@github.com:submlit21/stardewl-ink.git"
echo "3. 进入项目: cd stardewl-ink"
echo "4. 下载依赖: go mod download"
echo "5. 构建项目: make build"
echo "6. 运行CLI: ./dist/stardewl --interactive"
echo ""
echo "💡 提示：如果网络连接有问题，请检查代理设置。"