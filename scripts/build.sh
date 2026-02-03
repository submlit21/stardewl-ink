#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Stardewl-Ink...${NC}"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# 创建输出目录
mkdir -p dist

# 构建核心库（作为Go模块，不需要单独构建）
echo -e "${YELLOW}Checking core library...${NC}"
cd core
go test -c -o ../dist/core.test  # 构建测试二进制文件用于验证
cd ..

# 构建信令服务器
echo -e "${YELLOW}Building signaling server...${NC}"
cd signaling
go build -o ../dist/stardewl-signaling
cd ..

# 构建示例演示程序
echo -e "${YELLOW}Building example demo...${NC}"
go build -o dist/stardewl-demo examples/simple_demo.go

# 构建CLI应用
echo -e "${YELLOW}Building CLI application...${NC}"
go build -o dist/stardewl cmd/stardewl/main.go

echo -e "${GREEN}Build completed!${NC}"
echo -e "Output files in ${YELLOW}dist/${NC}:"
ls -la dist/

echo -e "\n${YELLOW}To start the signaling server:${NC}"
echo -e "  ./dist/stardewl-signaling"
echo -e "\n${YELLOW}To run the CLI application:${NC}"
echo -e "  ./dist/stardewl --interactive"
echo -e "  ./dist/stardewl --host"
echo -e "  ./dist/stardewl --join=123456"
echo -e "\n${YELLOW}To run the example demo:${NC}"
echo -e "  ./dist/stardewl-demo"