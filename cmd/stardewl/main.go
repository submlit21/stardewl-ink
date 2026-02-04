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
		hostMode       bool
		joinCode       string
		signalingURL   string
		modsPath       string
		verbose        bool
		listMods       bool
		interactive    bool
		timeoutSeconds int
	)

	flag.BoolVar(&hostMode, "host", false, "Run in host mode")
	flag.StringVar(&joinCode, "join", "", "Run in client mode, specify connection code")
	flag.StringVar(&signalingURL, "signaling", "", "Signaling server URL (default: ws://localhost:8080/ws)")
	flag.StringVar(&modsPath, "mods", "", "Mods folder path (default: auto-detect)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&listMods, "list-mods", false, "List mods in Mods folder")
	flag.BoolVar(&interactive, "interactive", false, "Interactive mode")
	flag.IntVar(&timeoutSeconds, "timeout", 0, "Timeout in seconds, 0 means wait indefinitely")
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
		runAsHost(signalingURL, modsPath, verbose, timeoutSeconds)
	} else if joinCode != "" {
		runAsClient(signalingURL, joinCode, modsPath, verbose, timeoutSeconds)
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

func runAsHost(signalingURL, modsPath string, verbose bool, timeoutSeconds int) {
	fmt.Println("=== Host Mode ===")
	
	// 如果没有指定信令服务器URL，使用默认值
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Printf("Signaling server: %s\n", signalingURL)
	
	fmt.Println("Creating room on signaling server...")
	
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
	fmt.Println("Creating room on signaling server...")
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
			fmt.Printf("⚠️  Create room attempt %d failed: %v\n", i+1, err)
		} else {
			resp.Body.Close()
			fmt.Printf("⚠️  Create room attempt %d failed, status code: %d\n", i+1, resp.StatusCode)
		}
		
		if i < 2 {
			time.Sleep(1 * time.Second)
		}
	}
	
	if err != nil {
		fmt.Printf("❌ Failed to create room (after 3 attempts): %v\n", err)
		fmt.Println("Please ensure signaling server is running: ./dist/stardewl-signaling")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("❌ Failed to create room, status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	var roomResponse struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roomResponse); err != nil {
		fmt.Printf("❌ Failed to parse room response: %v\n", err)
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
		log.Printf("Failed to start P2P connection: %v", err)
		os.Exit(1)
	}

	// 根据超时设置等待
	if timeoutSeconds > 0 {
		fmt.Printf("\nWaiting for %d seconds (timeout)...\n", timeoutSeconds)
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
		fmt.Println("Timeout reached, exiting...")
	} else {
		fmt.Print("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func runAsClient(signalingURL, connectionID, modsPath string, verbose bool, timeoutSeconds int) {
	fmt.Println("=== 客户端模式 ===")
	
	if connectionID == "" {
		fmt.Println("Error: Must specify connection code")
		os.Exit(1)
	}
	
	fmt.Printf("Connection code: %s\n", connectionID)
	
	// 如果没有指定信令服务器URL，使用默认值
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Printf("Signaling server: %s\n", signalingURL)
	fmt.Println("Connecting to host...")
	fmt.Println("(Press Ctrl+C to exit)")
	
	// 验证房间是否存在
	fmt.Println("Verifying room exists...")
	checkRoomURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	checkRoomURL = strings.Replace(checkRoomURL, "/ws", "/join/"+connectionID, 1)
	
	resp, err := http.Get(checkRoomURL)
	if err != nil {
		fmt.Printf("❌ 无法连接到信令服务器: %v\n", err)
		fmt.Println("Please ensure signaling server is running: ./dist/stardewl-signaling")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		fmt.Printf("❌ Room does not exist: %s\n", connectionID)
		fmt.Println("Please check connection code, or wait for host to create room")
		os.Exit(1)
	} else if resp.StatusCode != 200 {
		fmt.Printf("❌ Failed to verify room, status code: %d\n", resp.StatusCode)
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
		fmt.Printf("❌ Failed to parse room response: %v\n", err)
		os.Exit(1)
	}
	
	if roomResponse.Ready {
		fmt.Println("✅ Room verified (host connected)")
	} else {
		fmt.Println("⚠️  Room exists but host not connected")
		fmt.Println("Please wait for host to connect, or check if host is running")
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

	// 根据超时设置等待
	if timeoutSeconds > 0 {
		fmt.Printf("\nWaiting for %d seconds (timeout)...\n", timeoutSeconds)
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
		fmt.Println("Timeout reached, exiting...")
	} else {
		fmt.Print("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
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
