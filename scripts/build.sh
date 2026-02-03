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

# 构建核心库
echo -e "${YELLOW}Building core library...${NC}"
cd core
go build -o ../dist/libstardewl-core.a -buildmode=c-archive
cd ..

# 构建信令服务器
echo -e "${YELLOW}Building signaling server...${NC}"
cd signaling
go build -o ../dist/stardewl-signaling
cd ..

# 构建示例客户端
echo -e "${YELLOW}Building example client...${NC}"
cat > examples/simple_client.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	config := core.ClientConfig{
		SignalingURL: "ws://localhost:8080/ws",
		ConnectionID: "demo-connection",
		IsHost:       true,
		ICEServers:   core.GetDefaultICEServers(),
	}

	client, err := core.NewStardewlClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	client.SetModsCheckedHandler(func(comparison core.ModComparison) {
		fmt.Println("\n=== Mods Comparison ===")
		fmt.Println(core.FormatComparisonResult(comparison))
	})

	client.SetConnectedHandler(func() {
		fmt.Println("✓ Connected to peer")
	})

	client.SetDisconnectedHandler(func() {
		fmt.Println("✗ Disconnected from peer")
	})

	fmt.Println("Starting Stardewl-Ink client...")
	fmt.Println("Press Ctrl+C to exit")

	// 保持运行
	for {
		time.Sleep(1 * time.Second)
	}
}
EOF

go build -o dist/stardewl-example examples/simple_client.go

echo -e "${GREEN}Build completed!${NC}"
echo -e "Output files in ${YELLOW}dist/${NC}:"
ls -la dist/

echo -e "\n${YELLOW}To start the signaling server:${NC}"
echo -e "  ./dist/stardewl-signaling"
echo -e "\n${YELLOW}To run the example client:${NC}"
echo -e "  ./dist/stardewl-example"