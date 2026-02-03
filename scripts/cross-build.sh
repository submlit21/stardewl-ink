#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Stardewl-Ink Cross-compilation Tool${NC}"
echo -e "${BLUE}===================================${NC}"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# 显示帮助
show_help() {
    echo -e "${YELLOW}Usage:${NC} $0 [platform]"
    echo -e ""
    echo -e "${YELLOW}Platforms:${NC}"
    echo -e "  windows     Build for Windows (amd64)"
    echo -e "  macos       Build for macOS (arm64)"
    echo -e "  linux       Build for Linux (amd64)"
    echo -e "  all         Build for all platforms"
    echo -e "  list        List available platforms"
    echo -e ""
    echo -e "${YELLOW}Examples:${NC}"
    echo -e "  $0 windows          # Build Windows executables"
    echo -e "  $0 all              # Build for all platforms"
    echo -e "  $0                  # Show this help"
    echo -e ""
    echo -e "${YELLOW}Output directories:${NC}"
    echo -e "  dist/windows/       Windows .exe files"
    echo -e "  dist/macos/         macOS binaries"
    echo -e "  dist/linux/         Linux binaries"
}

# 平台配置
declare -A PLATFORMS
PLATFORMS[windows]="GOOS=windows GOARCH=amd64"
PLATFORMS[macos]="GOOS=darwin GOARCH=arm64"
PLATFORMS[linux]="GOOS=linux GOARCH=amd64"

# 列出平台
list_platforms() {
    echo -e "${YELLOW}Available platforms:${NC}"
    for platform in "${!PLATFORMS[@]}"; do
        echo -e "  ${GREEN}${platform}${NC} - ${PLATFORMS[$platform]}"
    done
}

# 构建单个平台
build_platform() {
    local platform=$1
    local env_vars=${PLATFORMS[$platform]}
    
    echo -e "\n${YELLOW}Building for ${platform}...${NC}"
    echo -e "Environment: ${env_vars}"
    
    # 确定文件扩展名
    local exe_ext=""
    if [ "$platform" = "windows" ]; then
        exe_ext=".exe"
    fi
    
    # 创建输出目录
    local output_dir="dist/${platform}"
    mkdir -p "$output_dir"
    
    # 构建核心库（检查兼容性）
    echo -e "  🔧 Checking core library..."
    eval $env_vars go build ./core/...
    
    # 构建CLI应用
    echo -e "  🖥️  Building CLI application..."
    eval $env_vars go build -o "${output_dir}/stardewl${exe_ext}" ./cmd/stardewl
    
    # 构建信令服务器
    echo -e "  🌐 Building signaling server..."
    eval $env_vars go build -o "${output_dir}/stardewl-signaling${exe_ext}" ./signaling
    
    # 构建演示程序
    echo -e "  🧪 Building example demo..."
    eval $env_vars go build -o "${output_dir}/stardewl-demo${exe_ext}" ./examples/simple_demo.go
    
    echo -e "  ✅ ${platform} builds saved to ${output_dir}/"
    
    # 显示构建信息
    echo -e "\n  ${BLUE}Build info:${NC}"
    echo -e "  Platform: ${platform}"
    echo -e "  Files:"
    for file in "${output_dir}"/*; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo -e "    - $(basename "$file") (${size})"
        fi
    done
}

# 构建所有平台
build_all() {
    echo -e "${YELLOW}Building for all platforms...${NC}"
    
    for platform in "${!PLATFORMS[@]}"; do
        build_platform "$platform"
    done
    
    echo -e "\n${GREEN}✅ All cross-platform builds completed!${NC}"
    echo -e "\n${BLUE}Summary:${NC}"
    echo -e "  Windows: dist/windows/stardewl.exe"
    echo -e "  macOS:   dist/macos/stardewl"
    echo -e "  Linux:   dist/linux/stardewl"
}

# 主逻辑
if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

case "$1" in
    "list")
        list_platforms
        ;;
    "all")
        build_all
        ;;
    "windows"|"macos"|"linux")
        if [[ -v PLATFORMS[$1] ]]; then
            build_platform "$1"
        else
            echo -e "${RED}Unknown platform: $1${NC}"
            show_help
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        show_help
        exit 1
        ;;
esac

echo -e "\n${GREEN}Cross-compilation completed successfully!${NC}"