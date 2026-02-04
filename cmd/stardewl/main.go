package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	var (
		hostMode      bool
		joinCode      string
		signalingURL  string
		modsPath      string
		verbose       bool
		listMods      bool
		interactive   bool
	)

	flag.BoolVar(&hostMode, "host", false, "以主机模式运行")
	flag.StringVar(&joinCode, "join", "", "以客户端模式运行，指定连接码")
	flag.StringVar(&signalingURL, "signaling", "", "信令服务器URL (默认: ws://localhost:8080/ws)")
	flag.StringVar(&modsPath, "mods", "", "Mods文件夹路径 (默认: 自动检测)")
	flag.BoolVar(&verbose, "verbose", false, "启用详细日志")
	flag.BoolVar(&listMods, "list-mods", false, "列出Mods文件夹中的mods")
	flag.BoolVar(&interactive, "interactive", false, "交互模式")
	flag.Parse()

	if interactive {
		runInteractive()
		return
	}

	if listMods {
		listModsInPath(modsPath)
		return
	}

	if hostMode {
		runAsHost(signalingURL, modsPath, verbose)
	} else if joinCode != "" {
		runAsClient(signalingURL, joinCode, modsPath, verbose)
	} else {
		fmt.Println("请指定模式:")
		fmt.Println("  --host                   以主机模式运行")
		fmt.Println("  --join=<code>            以客户端模式运行，指定连接码")
		fmt.Println("  --interactive            交互模式")
		fmt.Println("  --list-mods              列出Mods文件夹中的mods")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  --signaling=<url>        信令服务器URL")
		fmt.Println("  --mods=<path>            Mods文件夹路径")
		fmt.Println("  --verbose                启用详细日志")
	}
}

func runAsHost(signalingURL, modsPath string, verbose bool) {
	fmt.Println("=== 主机模式 ===")
	
	// 如果没有指定信令服务器URL，使用默认值
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Printf("信令服务器: %s\n", signalingURL)
	
	// 创建P2P连接器配置
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		IsHost:       true,
		ModsPath:     modsPath,
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}

	// 自动生成房间ID
	roomID := core.GenerateRoomID()
	config.RoomID = roomID

	// 先在信令服务器上创建房间（带重试）
	fmt.Println("在信令服务器上创建房间...")
	createRoomURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	createRoomURL = strings.Replace(createRoomURL, "/ws", "/create", 1)
	
	var resp *http.Response
	var err error
	
	// 重试3次，每次等待1秒
	for i := 0; i < 3; i++ {
		resp, err = http.Post(createRoomURL, "application/json", nil)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		
		if err != nil {
			fmt.Printf("⚠️  创建房间尝试 %d 失败: %v\n", i+1, err)
		} else {
			resp.Body.Close()
			fmt.Printf("⚠️  创建房间尝试 %d 失败，状态码: %d\n", i+1, resp.StatusCode)
		}
		
		if i < 2 {
			time.Sleep(1 * time.Second)
		}
	}
	
	if err != nil {
		fmt.Printf("❌ 创建房间失败（重试3次后）: %v\n", err)
		fmt.Println("请确保信令服务器正在运行: ./dist/stardewl-signaling")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 创建房间失败，状态码: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	var roomResponse struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roomResponse); err != nil {
		fmt.Printf("❌ 解析房间响应失败: %v\n", err)
		os.Exit(1)
	}
	
	// 使用服务器返回的房间ID
	config.RoomID = roomResponse.Code
	roomID = roomResponse.Code

	fmt.Printf("Connection code: %s\n", roomID)
	fmt.Println("Waiting for client connection...")
	fmt.Println("(Press Ctrl+C to exit)")

	// Create P2P connector
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		log.Printf("Failed to create P2P connector: %v", err)
		os.Exit(1)
	}
	defer connector.Close()

	// 启动连接
	if err := connector.Start(); err != nil {
		log.Printf(" 启动P2P连接失败: %v", err)
		os.Exit(1)
	}

	// 简单等待
	fmt.Print("\n按 Enter 键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func runAsClient(signalingURL, connectionID, modsPath string, verbose bool) {
	fmt.Println("=== 客户端模式 ===")
	
	if connectionID == "" {
		fmt.Println("错误: 必须指定连接码")
		os.Exit(1)
	}
	
	fmt.Printf("连接码: %s\n", connectionID)
	
	// 如果没有指定信令服务器URL，使用默认值
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Printf("信令服务器: %s\n", signalingURL)
	fmt.Println("正在连接到主机...")
	fmt.Println("(按 Ctrl+C 退出)")
	
	// 验证房间是否存在
	fmt.Println("验证房间是否存在...")
	checkRoomURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	checkRoomURL = strings.Replace(checkRoomURL, "/ws", "/join/"+connectionID, 1)
	
	resp, err := http.Get(checkRoomURL)
	if err != nil {
		fmt.Printf("❌ 无法连接到信令服务器: %v\n", err)
		fmt.Println("请确保信令服务器正在运行: ./dist/stardewl-signaling")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		fmt.Printf("❌ 房间不存在: %s\n", connectionID)
		fmt.Println("请检查连接码是否正确，或让主机先创建房间")
		os.Exit(1)
	} else if resp.StatusCode != 200 {
		fmt.Printf("❌ 验证房间失败，状态码: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	// 解析响应
	var roomResponse struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Ready   bool   `json:"ready"`
		Message string `json:"message,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&roomResponse); err != nil {
		fmt.Printf("❌ 解析房间响应失败: %v\n", err)
		os.Exit(1)
	}
	
	if roomResponse.Ready {
		fmt.Println("✅ 房间验证通过 (主机已连接)")
	} else {
		fmt.Println("⚠️  房间存在但主机未连接")
		fmt.Println("请等待主机连接，或检查主机是否正在运行")
		// 这里可以选择等待或退出
	}

	// 创建P2P连接器配置
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		RoomID:       connectionID,
		IsHost:       false,
		ModsPath:     modsPath,
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}

	// Create P2P connector
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		log.Printf("Failed to create P2P connector: %v", err)
		os.Exit(1)
	}
	defer connector.Close()

	// Start connection
	if err := connector.Start(); err != nil {
		log.Printf("Failed to start P2P connection: %v", err)
		os.Exit(1)
	}

	// Simple wait
	fmt.Print("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func listModsInPath(modsPath string) {
	fmt.Println("=== Listing Mods ===")
	
	mods, err := core.ScanMods(modsPath)
	if err != nil {
		log.Printf("Failed to scan Mods: %v", err)
		os.Exit(1)
	}

	if len(mods) == 0 {
		fmt.Println("未找到Mods")
		return
	}

	fmt.Printf("找到 %d 个Mods:\n", len(mods))
	for _, mod := range mods {
		fmt.Printf("  - %s (%s)\n", mod.Name, mod.Version)
	}
}
