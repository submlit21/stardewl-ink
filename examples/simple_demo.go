package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/submlit21/stardewl-ink/core"
)

func main() {
	fmt.Println("=== Stardewl-Ink 演示程序 ===")
	
	// 演示Mod扫描功能
	fmt.Println("\n1. 测试Mod扫描功能:")
	
	// 创建一个测试Mods目录
	testDir := filepath.Join(os.TempDir(), "stardewl-test-mods")
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	
	// 创建一些测试文件
	testFiles := []struct {
		name string
		content string
	}{
		{"TestMod1.mod", "This is test mod 1"},
		{"TestMod2.dll", "This is test mod 2 DLL"},
		{"AnotherMod.zip", "This is another mod"},
	}
	
	for _, file := range testFiles {
		path := filepath.Join(testDir, file.name)
		if err := os.WriteFile(path, []byte(file.content), 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  创建测试文件: %s\n", file.name)
	}
	
	// 扫描Mods
	mods, err := core.ScanMods(testDir)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("  扫描到 %d 个Mod文件:\n", len(mods))
	for _, mod := range mods {
		fmt.Printf("    - %s (大小: %d bytes, 哈希: %s...)\n", 
			mod.Name, mod.Size, mod.Checksum[:8])
	}
	
	// 演示Mod对比功能
	fmt.Println("\n2. 测试Mod对比功能:")
	
	localMods := []core.ModInfo{
		{Name: "CommonMod", Checksum: "abc123", Size: 100},
		{Name: "LocalOnly", Checksum: "def456", Size: 200},
		{Name: "DifferentVersion", Checksum: "ver1", Size: 300},
	}
	
	remoteMods := []core.ModInfo{
		{Name: "CommonMod", Checksum: "abc123", Size: 100},
		{Name: "RemoteOnly", Checksum: "ghi789", Size: 400},
		{Name: "DifferentVersion", Checksum: "ver2", Size: 350},
	}
	
	comparison := core.CompareMods(localMods, remoteMods)
	
	fmt.Println("  对比结果:")
	fmt.Printf("    相同的Mod: %d 个\n", len(comparison.Same))
	fmt.Printf("    只在本地存在的Mod: %d 个\n", len(comparison.OnlyInLocal))
	fmt.Printf("    只在远程存在的Mod: %d 个\n", len(comparison.OnlyInRemote))
	fmt.Printf("    版本不同的Mod: %d 个\n", len(comparison.Different))
	
	// 演示客户端创建
	fmt.Println("\n3. 测试客户端创建:")
	
	config := core.ClientConfig{
		SignalingURL: "ws://localhost:8080/ws",
		ConnectionID: "demo-connection-123",
		IsHost:       true,
		ICEServers:   core.GetDefaultICEServers(),
		ModsPath:     testDir,
	}
	
	client, err := core.NewStardewlClient(config)
	if err != nil {
		log.Printf("  客户端创建失败: %v\n", err)
	} else {
		fmt.Println("  客户端创建成功!")
		fmt.Printf("  连接ID: %s\n", client.ConnectionID())
		fmt.Printf("  是否是主机: %v\n", client.IsHost())
		
		// 设置回调
		client.SetModsCheckedHandler(func(comparison core.ModComparison) {
			fmt.Println("\n  Mod检查完成:")
			fmt.Println(core.FormatComparisonResult(comparison))
		})
		
		client.SetConnectedHandler(func() {
			fmt.Println("  已连接到对端")
		})
		
		client.SetDisconnectedHandler(func() {
			fmt.Println("  与对端断开连接")
		})
		
		// 清理
		client.Close()
		fmt.Println("  客户端已关闭")
	}
	
	// 演示默认路径检测
	fmt.Println("\n4. 测试星露谷默认路径检测:")
	defaultPath := core.GetDefaultStardewValleyModsPath()
	if defaultPath != "" {
		fmt.Printf("  检测到默认Mods路径: %s\n", defaultPath)
		
		// 检查路径是否存在
		if _, err := os.Stat(defaultPath); err == nil {
			fmt.Println("  路径存在")
			
			// 尝试扫描（如果路径存在且有文件）
			mods, err := core.ScanMods(defaultPath)
			if err != nil {
				fmt.Printf("  扫描失败: %v\n", err)
			} else {
				fmt.Printf("  扫描到 %d 个Mod文件\n", len(mods))
			}
		} else {
			fmt.Println("  路径不存在或无法访问")
		}
	} else {
		fmt.Println("  未检测到默认Mods路径")
	}
	
	// 清理测试文件
	os.RemoveAll(testDir)
	
	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("\n下一步:")
	fmt.Println("1. 启动信令服务器: ./dist/stardewl-signaling")
	fmt.Println("2. 运行完整示例: go run examples/simple_demo.go")
	fmt.Println("3. 查看架构文档: cat docs/ARCHITECTURE.md")
	
	// 保持程序运行一段时间
	fmt.Println("\n程序将在5秒后退出...")
	time.Sleep(5 * time.Second)
}