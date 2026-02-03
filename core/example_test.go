package core_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/submlit21/stardewl-ink/core"
)

func TestModScanning(t *testing.T) {
	// 创建一个临时目录来测试Mod扫描
	tempDir := t.TempDir()
	
	// 测试空的Mods目录
	mods, err := core.ScanMods(tempDir)
	if err != nil {
		t.Fatalf("Failed to scan empty mods directory: %v", err)
	}
	
	if len(mods) != 0 {
		t.Errorf("Expected 0 mods in empty directory, got %d", len(mods))
	}
	
	// 测试默认路径函数
	defaultPath := core.GetDefaultStardewValleyModsPath()
	fmt.Printf("Default Stardew Valley Mods path: %s\n", defaultPath)
}

func TestModComparison(t *testing.T) {
	// 创建测试Mod数据
	localMods := []core.ModInfo{
		{Name: "TestMod1", Checksum: "abc123", Size: 100},
		{Name: "TestMod2", Checksum: "def456", Size: 200},
		{Name: "TestMod3", Checksum: "same", Size: 300},
	}
	
	remoteMods := []core.ModInfo{
		{Name: "TestMod2", Checksum: "different", Size: 250}, // 不同的版本
		{Name: "TestMod3", Checksum: "same", Size: 300},      // 相同的
		{Name: "TestMod4", Checksum: "ghi789", Size: 400},    // 只在远程存在
	}
	
	comparison := core.CompareMods(localMods, remoteMods)
	
	// 验证结果
	if len(comparison.OnlyInLocal) != 1 || comparison.OnlyInLocal[0].Name != "TestMod1" {
		t.Errorf("Expected TestMod1 only in local, got %v", comparison.OnlyInLocal)
	}
	
	if len(comparison.OnlyInRemote) != 1 || comparison.OnlyInRemote[0].Name != "TestMod4" {
		t.Errorf("Expected TestMod4 only in remote, got %v", comparison.OnlyInRemote)
	}
	
	if len(comparison.Different) != 1 || comparison.Different[0].Name != "TestMod2" {
		t.Errorf("Expected TestMod2 different, got %v", comparison.Different)
	}
	
	if len(comparison.Same) != 1 || comparison.Same[0].Name != "TestMod3" {
		t.Errorf("Expected TestMod3 same, got %v", comparison.Same)
	}
	
	// 测试格式化输出
	formatted := core.FormatComparisonResult(comparison)
	fmt.Println("Formatted comparison result:")
	fmt.Println(formatted)
}

func ExampleNewStardewlClient() {
	// 这是一个使用示例
	
	// 创建客户端配置
	config := core.ClientConfig{
		SignalingURL: "ws://localhost:8080/ws",
		ConnectionID: "test-123",
		IsHost:       true,
		ICEServers:   core.GetDefaultICEServers(),
	}
	
	// 创建客户端
	client, err := core.NewStardewlClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	
	// 设置回调
	client.SetModsCheckedHandler(func(comparison core.ModComparison) {
		fmt.Println("Mods comparison completed:")
		fmt.Println(core.FormatComparisonResult(comparison))
	})
	
	client.SetConnectedHandler(func() {
		fmt.Println("Connected to peer")
		
		// 连接成功后发送Mod列表
		if err := client.SendModsList(); err != nil {
			log.Printf("Failed to send mods list: %v", err)
		}
	})
	
	client.SetDisconnectedHandler(func() {
		fmt.Println("Disconnected from peer")
	})
	
	// 作为主机启动
	if client.IsHost() {
		if err := client.StartAsHost(); err != nil {
			log.Fatalf("Failed to start as host: %v", err)
		}
	}
	
	// 启动心跳
	client.StartHeartbeat(30 * time.Second)
	
	// 保持运行
	time.Sleep(5 * time.Second)
	
	// Output: (示例输出，实际运行会有不同)
}