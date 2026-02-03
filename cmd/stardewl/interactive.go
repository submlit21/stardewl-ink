package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("🎮 星露谷联机工具 - 交互模式")
	fmt.Println(strings.Repeat("=", 40))
	
	for {
		fmt.Println("\n请选择模式:")
		fmt.Println("1. 作为主机运行（创建房间）")
		fmt.Println("2. 作为客户端运行（加入房间）")
		fmt.Println("3. 检查本地Mods")
		fmt.Println("4. 启动信令服务器")
		fmt.Println("5. 退出")
		fmt.Print("\n选择 (1-5): ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		
		switch choice {
		case "1":
			runHostInteractive(reader)
		case "2":
			runClientInteractive(reader)
		case "3":
			runModsCheckInteractive(reader)
		case "4":
			runSignalingInteractive()
		case "5":
			fmt.Println("👋 再见！")
			return
		default:
			fmt.Println("❌ 无效选择，请重试")
		}
	}
}

func runHostInteractive(reader *bufio.Reader) {
	fmt.Println("\n🎯 主机模式")
	fmt.Println(strings.Repeat("-", 30))
	
	// 获取Mods路径
	fmt.Print("Mods路径（留空使用默认）: ")
	modsPath, _ := reader.ReadString('\n')
	modsPath = strings.TrimSpace(modsPath)
	
	// 获取信令服务器地址
	fmt.Print("信令服务器地址（留空使用默认）: ")
	signalingURL, _ := reader.ReadString('\n')
	signalingURL = strings.TrimSpace(signalingURL)
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Println("\n正在启动主机...")
	
	// 这里可以调用实际的host逻辑
	// 暂时显示模拟信息
	fmt.Println("✅ 主机已启动")
	fmt.Println("📋 连接码: 123456")
	fmt.Println("⏳ 等待客户端连接...")
	
	fmt.Print("\n按 Enter 返回主菜单...")
	reader.ReadString('\n')
}

func runClientInteractive(reader *bufio.Reader) {
	fmt.Println("\n🎯 客户端模式")
	fmt.Println(strings.Repeat("-", 30))
	
	// 获取连接码
	fmt.Print("请输入连接码: ")
	connectionID, _ := reader.ReadString('\n')
	connectionID = strings.TrimSpace(connectionID)
	
	if connectionID == "" {
		fmt.Println("❌ 连接码不能为空")
		return
	}
	
	// 获取Mods路径
	fmt.Print("Mods路径（留空使用默认）: ")
	modsPath, _ := reader.ReadString('\n')
	modsPath = strings.TrimSpace(modsPath)
	
	// 获取信令服务器地址
	fmt.Print("信令服务器地址（留空使用默认）: ")
	signalingURL, _ := reader.ReadString('\n')
	signalingURL = strings.TrimSpace(signalingURL)
	if signalingURL == "" {
		signalingURL = "ws://localhost:8080/ws"
	}
	
	fmt.Printf("\n正在连接到主机 %s...\n", connectionID)
	
	// 这里可以调用实际的client逻辑
	// 暂时显示模拟信息
	fmt.Println("✅ 已连接到主机")
	fmt.Println("🔍 正在检查Mods...")
	
	// 模拟Mod检查结果
	fmt.Println("\n📊 Mods对比结果:")
	fmt.Println("   相同的Mod: 5个")
	fmt.Println("   不同的Mod: 2个")
	fmt.Println("   需要同步的Mod: 1个")
	
	fmt.Print("\n按 Enter 返回主菜单...")
	reader.ReadString('\n')
}

func runModsCheckInteractive(reader *bufio.Reader) {
	fmt.Println("\n🔍 Mods检查")
	fmt.Println(strings.Repeat("-", 30))
	
	// 获取Mods路径
	fmt.Print("Mods路径（留空使用默认）: ")
	modsPath, _ := reader.ReadString('\n')
	modsPath = strings.TrimSpace(modsPath)
	
	fmt.Printf("\n正在扫描 %s...\n", modsPath)
	
	// 这里可以调用实际的Mods检查逻辑
	// 暂时显示模拟信息
	fmt.Println("✅ 扫描完成")
	fmt.Println("📊 扫描结果:")
	fmt.Println("   找到Mod文件: 8个")
	fmt.Println("   总大小: 45.2 MB")
	fmt.Println("   最新修改: 2024-01-15 14:30:22")
	
	fmt.Print("\n按 Enter 返回主菜单...")
	reader.ReadString('\n')
}

func runSignalingInteractive() {
	fmt.Println("\n🌐 信令服务器")
	fmt.Println(strings.Repeat("-", 30))
	
	fmt.Println("正在启动信令服务器...")
	fmt.Println("服务器地址: ws://localhost:8080/ws")
	fmt.Println("HTTP接口: http://localhost:8080")
	fmt.Println("\n按 Ctrl+C 停止服务器")
	
	// 这里可以启动实际的信令服务器
	// 暂时显示信息
	fmt.Println("\n✅ 服务器已启动（模拟）")
	fmt.Println("📈 状态: 运行中")
	fmt.Println("👥 连接数: 0")
	
	fmt.Print("\n按 Enter 返回主菜单...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}