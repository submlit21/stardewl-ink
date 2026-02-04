package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	fmt.Println("=== 简化客户端测试 ===")
	
	// 从环境变量获取房间ID，或使用默认值
	roomID := os.Getenv("ROOM_ID")
	if roomID == "" {
		fmt.Println("错误: 请设置ROOM_ID环境变量")
		fmt.Println("例如: ROOM_ID=053684 ./dist/test-client")
		os.Exit(1)
	}
	
	signalingURL := "ws://localhost:8080/ws"
	
	fmt.Printf("房间ID: %s\n", roomID)
	fmt.Printf("信令服务器: %s\n", signalingURL)
	
	// 创建配置
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		RoomID:       roomID,
		IsHost:       false,
		ModsPath:     "",
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	
	fmt.Println("创建P2P连接器...")
	
	// 创建P2P连接器
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		log.Fatalf("❌ 创建P2P连接器失败: %v", err)
	}
	defer connector.Close()
	
	fmt.Println("✅ P2P连接器创建成功")
	
	// 启动连接
	fmt.Println("启动连接...")
	if err := connector.Start(); err != nil {
		log.Fatalf("❌ 启动连接失败: %v", err)
	}
	
	fmt.Println("✅ 连接启动成功")
	fmt.Println("等待10秒...")
	
	// 简单等待
	time.Sleep(10 * time.Second)
	
	fmt.Println("✅ 测试完成")
}