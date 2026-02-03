package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/submlit21/stardewl-ink/core"
)

var (
	version = "0.1.0"
)

func main() {
	// 解析命令行参数
	host := flag.Bool("host", false, "作为主机运行（生成连接码）")
	join := flag.String("join", "", "作为客户端运行，加入指定连接码")
	signaling := flag.String("signaling", "ws://localhost:8080/ws", "信令服务器地址")
	modsPath := flag.String("mods", "", "星露谷Mods路径（默认自动检测）")
	listMods := flag.Bool("list-mods", false, "列出本地Mods")
	checkOnly := flag.Bool("check-only", false, "只检查Mods，不建立连接")
	verbose := flag.Bool("verbose", false, "显示详细日志")
	versionFlag := flag.Bool("version", false, "显示版本信息")
	interactive := flag.Bool("interactive", false, "交互模式")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "星露谷联机工具 v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "使用方法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  # 交互模式\n")
		fmt.Fprintf(os.Stderr, "  %s --interactive\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 作为主机运行\n")
		fmt.Fprintf(os.Stderr, "  %s --host\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 作为客户端加入\n")
		fmt.Fprintf(os.Stderr, "  %s --join=123456\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 只检查Mods\n")
		fmt.Fprintf(os.Stderr, "  %s --list-mods\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 使用自定义信令服务器\n")
		fmt.Fprintf(os.Stderr, "  %s --host --signaling=ws://example.com:8080/ws\n", os.Args[0])
	}
	
	flag.Parse()
	
	// 显示版本信息
	if *versionFlag {
		fmt.Printf("stardewl v%s\n", version)
		os.Exit(0)
	}
	
	// 设置日志级别
	if !*verbose {
		log.SetOutput(io.Discard)
	}
	
	// 如果未指定Mods路径，尝试自动检测
	if *modsPath == "" {
		defaultPath := core.GetDefaultStardewValleyModsPath()
		if defaultPath != "" {
			*modsPath = defaultPath
			if *verbose {
				log.Printf("使用自动检测的Mods路径: %s\n", defaultPath)
			}
		} else {
			if *verbose {
				log.Println("警告: 未检测到星露谷Mods路径")
			}
		}
	}
	
	// 只列出Mods模式
	if *listMods {
		listLocalMods(*modsPath, *verbose)
		os.Exit(0)
	}
	
	// 检查参数
	if *host && *join != "" {
		fmt.Fprintf(os.Stderr, "错误: 不能同时指定 --host 和 --join\n")
		os.Exit(1)
	}
	
	// 交互模式不需要其他参数
	if *interactive {
		// 交互模式会处理所有逻辑
	} else if !*host && *join == "" && !*checkOnly && !*listMods {
		fmt.Fprintf(os.Stderr, "错误: 必须指定运行模式\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// 运行主逻辑
	if *interactive {
		runInteractive()
	} else if *checkOnly {
		runModsCheck(*modsPath, *verbose)
	} else if *host {
		runAsHost(*signaling, *modsPath, *verbose)
	} else if *join != "" {
		runAsClient(*signaling, *join, *modsPath, *verbose)
	} else {
		fmt.Fprintf(os.Stderr, "错误: 必须指定运行模式\n")
		flag.Usage()
		os.Exit(1)
	}
}

func listLocalMods(modsPath string, verbose bool) {
	fmt.Println("=== 本地Mods列表 ===")
	
	if modsPath == "" {
		fmt.Println("未指定Mods路径")
		return
	}
	
	mods, err := core.ScanMods(modsPath)
	if err != nil {
		fmt.Printf("扫描Mods失败: %v\n", err)
		return
	}
	
	if len(mods) == 0 {
		fmt.Println("未找到Mod文件")
		return
	}
	
	fmt.Printf("找到 %d 个Mod文件:\n", len(mods))
	for i, mod := range mods {
		hashDisplay := mod.Checksum
		if len(hashDisplay) > 8 {
			hashDisplay = hashDisplay[:8]
		}
		fmt.Printf("%3d. %-30s %8d bytes  %s\n", 
			i+1, mod.Name, mod.Size, hashDisplay)
	}
	
	// 显示路径信息
	fmt.Printf("\n扫描路径: %s\n", modsPath)
	if stat, err := os.Stat(modsPath); err == nil {
		fmt.Printf("路径类型: 目录\n")
		fmt.Printf("修改时间: %s\n", stat.ModTime().Format("2006-01-02 15:04:05"))
	}
}

func runModsCheck(modsPath string, verbose bool) {
	fmt.Println("=== Mods检查 ===")
	
	if modsPath == "" {
		fmt.Println("错误: 未指定Mods路径")
		os.Exit(1)
	}
	
	mods, err := core.ScanMods(modsPath)
	if err != nil {
		fmt.Printf("扫描Mods失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("扫描完成，找到 %d 个Mod文件\n", len(mods))
	
	// 显示统计信息
	var totalSize int64
	for _, mod := range mods {
		totalSize += mod.Size
	}
	
	fmt.Printf("总大小: %.2f MB\n", float64(totalSize)/1024/1024)
	
	// 如果有Mods，显示详细信息
	if len(mods) > 0 && verbose {
		fmt.Println("\n详细列表:")
		for _, mod := range mods {
			hashDisplay := mod.Checksum
			if len(hashDisplay) > 8 {
				hashDisplay = hashDisplay[:8]
			}
			fmt.Printf("  - %s (%s, %d bytes)\n", mod.Name, hashDisplay, mod.Size)
		}
	}
}

func runAsHost(signalingURL, modsPath string, verbose bool) {
	fmt.Println("=== 主机模式 ===")
	
	// 从服务器获取连接码
	connectionID, err := getConnectionCodeFromServer(signalingURL)
	if err != nil {
		fmt.Printf("获取连接码失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 连接码: %s\n", connectionID)
	fmt.Println("等待客户端连接...")
	fmt.Println("(按 Ctrl+C 退出)")
	
	// 创建P2P配置
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		RoomID:       connectionID,
		IsHost:       true,
		ModsPath:     modsPath,
		ICEServers:   core.GetDefaultICEServers(),
	}
	
	// 创建P2P连接器
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		fmt.Printf("创建P2P连接器失败: %v\n", err)
		os.Exit(1)
	}
	defer connector.Close()
	
	// 设置回调
	connector.SetCallbacks(
		func(comparison core.ModComparison) {
			fmt.Println("\n" + strings.Repeat("=", 50))
			fmt.Println("Mods对比结果:")
			fmt.Println(core.FormatComparisonResult(comparison))
			fmt.Println(strings.Repeat("=", 50))
			
			// 如果有差异，提示用户
			if len(comparison.OnlyInLocal) > 0 || 
			   len(comparison.OnlyInRemote) > 0 || 
			   len(comparison.Different) > 0 {
				fmt.Println("\n⚠️  发现Mod差异！")
				fmt.Println("请确保双方Mod一致后再开始游戏。")
			} else if len(comparison.Same) > 0 {
				fmt.Println("\n✅ 所有Mod一致，可以开始游戏！")
			}
		},
		func() {
			fmt.Println("\n✅ 客户端已连接")
			fmt.Println("正在交换Mod信息...")
			
			// 发送Mod列表
			if err := connector.SendModsList(); err != nil {
				fmt.Printf("发送Mod列表失败: %v\n", err)
			}
		},
		func() {
			fmt.Println("\n❌ 客户端断开连接")
		},
	)
	
	// 启动P2P连接
	if err := connector.Start(); err != nil {
		fmt.Printf("启动P2P连接失败: %v\n", err)
		os.Exit(1)
	}
	
	// 等待用户中断
	waitForInterrupt()
	
	fmt.Println("\n👋 程序退出")
}

func runAsClient(signalingURL, connectionID, modsPath string, verbose bool) {
	fmt.Println("=== 客户端模式 ===")
	
	if connectionID == "" {
		fmt.Println("错误: 必须指定连接码")
		os.Exit(1)
	}
	
	fmt.Printf("连接码: %s\n", connectionID)
	fmt.Println("正在连接到主机...")
	fmt.Println("(按 Ctrl+C 退出)")
	
	// 创建P2P配置
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		RoomID:       connectionID,
		IsHost:       false,
		ModsPath:     modsPath,
		ICEServers:   core.GetDefaultICEServers(),
	}
	
	// 创建P2P连接器
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		fmt.Printf("创建P2P连接器失败: %v\n", err)
		os.Exit(1)
	}
	defer connector.Close()
	
	// 设置回调
	connector.SetCallbacks(
		func(comparison core.ModComparison) {
			fmt.Println("\n" + strings.Repeat("=", 50))
			fmt.Println("Mods对比结果:")
			fmt.Println(core.FormatComparisonResult(comparison))
			fmt.Println(strings.Repeat("=", 50))
			
			// 如果有差异，提示用户
			if len(comparison.OnlyInLocal) > 0 || 
			   len(comparison.OnlyInRemote) > 0 || 
			   len(comparison.Different) > 0 {
				fmt.Println("\n⚠️  发现Mod差异！")
				fmt.Println("请确保双方Mod一致后再开始游戏。")
			} else if len(comparison.Same) > 0 {
				fmt.Println("\n✅ 所有Mod一致，可以开始游戏！")
			}
		},
		func() {
			fmt.Println("\n✅ 已连接到主机")
		},
		func() {
			fmt.Println("\n❌ 与主机断开连接")
		},
	)
	
	// 启动P2P连接
	if err := connector.Start(); err != nil {
		fmt.Printf("启动P2P连接失败: %v\n", err)
		os.Exit(1)
	}
	
	// 等待用户中断
	waitForInterrupt()
	
	fmt.Println("\n👋 程序退出")
}

func waitForInterrupt() {
	// 简单版本：等待用户输入
	fmt.Print("\n按 Enter 键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}