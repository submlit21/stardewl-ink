#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Stardewl-Ink for current platform...${NC}"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# 获取当前平台信息
CURRENT_OS=$(go env GOOS)
CURRENT_ARCH=$(go env GOARCH)
echo -e "Platform: ${YELLOW}${CURRENT_OS}/${CURRENT_ARCH}${NC}"

# 创建输出目录
mkdir -p dist

# 检查核心库
echo -e "${YELLOW}Checking core library...${NC}"
cd core
go build ./...  # 只构建，不运行测试
cd ..

# 确定文件扩展名
EXE_EXT=""
if [ "$CURRENT_OS" = "windows" ]; then
    EXE_EXT=".exe"
fi

# 构建信令服务器
echo -e "${YELLOW}Building signaling server...${NC}"
go build -o "dist/stardewl-signaling${EXE_EXT}" ./signaling

# 构建示例演示程序
echo -e "${YELLOW}Building example demo...${NC}"
go build -o "dist/stardewl-demo${EXE_EXT}" ./examples/simple_demo.go

# 构建CLI应用
echo -e "${YELLOW}Building CLI application...${NC}"
go build -o "dist/stardewl${EXE_EXT}" ./cmd/stardewl

echo -e "${GREEN}Build completed!${NC}"
echo -e "Output files in ${YELLOW}dist/${NC}:"
ls -la dist/

echo -e "\n${YELLOW}To start the signaling server:${NC}"
echo -e "  ./dist/stardewl-signaling${EXE_EXT}"
echo -e "\n${YELLOW}To run the CLI application:${NC}"
echo -e "  ./dist/stardewl${EXE_EXT} --interactive"
echo -e "  ./dist/stardewl${EXE_EXT} --host"
echo -e "  ./dist/stardewl${EXE_EXT} --join=123456"
echo -e "\n${YELLOW}To run the example demo:${NC}"
echo -e "  ./dist/stardewl-demo${EXE_EXT}"

echo -e "\n${YELLOW}Cross-compilation targets (via Makefile):${NC}"
echo -e "  make cross-build-all      # Build for all platforms"
echo -e "  make cross-build-windows  # Build Windows .exe files"
echo -e "  make cross-build-macos    # Build macOS binaries"
echo -e "  make cross-build-linux    # Build Linux binaries"