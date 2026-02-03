package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ModInfo 表示一个Mod的信息
type ModInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Checksum string `json:"checksum"`
	Size     int64  `json:"size"`
	Path     string `json:"path,omitempty"`
}

// ModComparison 表示Mod对比结果
type ModComparison struct {
	OnlyInLocal  []ModInfo `json:"only_in_local"`
	OnlyInRemote []ModInfo `json:"only_in_remote"`
	Different    []ModDiff `json:"different"`
	Same         []ModInfo `json:"same"`
}

// ModDiff 表示不同的Mod信息
type ModDiff struct {
	Name    string `json:"name"`
	Local   ModInfo `json:"local"`
	Remote  ModInfo `json:"remote"`
}

// ScanMods 扫描指定路径下的Mods文件夹
func ScanMods(modsPath string) ([]ModInfo, error) {
	var mods []ModInfo
	
	// 检查路径是否存在
	if _, err := os.Stat(modsPath); os.IsNotExist(err) {
		return mods, nil // 路径不存在，返回空列表
	}

	// 遍历Mods文件夹
	err := filepath.Walk(modsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身和隐藏文件
		if info.IsDir() && path == modsPath {
			return nil
		}
		
		// 只处理文件和.mod文件
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".mod") || 
			strings.HasSuffix(strings.ToLower(info.Name()), ".dll") ||
			strings.HasSuffix(strings.ToLower(info.Name()), ".zip")) {
			
			mod, err := getModInfo(path)
			if err != nil {
				// 如果无法读取文件，跳过它
				return nil
			}
			
			mods = append(mods, mod)
		}
		
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan mods: %w", err)
	}

	// 按名称排序
	sort.Slice(mods, func(i, j int) bool {
		return mods[i].Name < mods[j].Name
	})

	return mods, nil
}

// getModInfo 获取单个Mod文件的信息
func getModInfo(filePath string) (ModInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ModInfo{}, err
	}
	defer file.Close()

	// 计算文件哈希
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ModInfo{}, err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return ModInfo{}, err
	}

	// 从文件名提取Mod名称（去掉扩展名）
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	return ModInfo{
		Name:     name,
		Checksum: checksum,
		Size:     info.Size(),
		Path:     filePath,
	}, nil
}

// CompareMods 比较本地和远程的Mod列表
func CompareMods(localMods, remoteMods []ModInfo) ModComparison {
	var result ModComparison
	
	// 创建映射以便快速查找
	localMap := make(map[string]ModInfo)
	remoteMap := make(map[string]ModInfo)
	
	for _, mod := range localMods {
		localMap[mod.Name] = mod
	}
	
	for _, mod := range remoteMods {
		remoteMap[mod.Name] = mod
	}
	
	// 找出只在本地存在的Mod
	for name, mod := range localMap {
		if _, exists := remoteMap[name]; !exists {
			result.OnlyInLocal = append(result.OnlyInLocal, mod)
		}
	}
	
	// 找出只在远程存在的Mod
	for name, mod := range remoteMap {
		if _, exists := localMap[name]; !exists {
			result.OnlyInRemote = append(result.OnlyInRemote, mod)
		}
	}
	
	// 找出两边都存在的Mod，比较是否相同
	for name, localMod := range localMap {
		if remoteMod, exists := remoteMap[name]; exists {
			if localMod.Checksum == remoteMod.Checksum && localMod.Size == remoteMod.Size {
				result.Same = append(result.Same, localMod)
			} else {
				result.Different = append(result.Different, ModDiff{
					Name:   name,
					Local:  localMod,
					Remote: remoteMod,
				})
			}
		}
	}
	
	// 排序结果
	sort.Slice(result.OnlyInLocal, func(i, j int) bool {
		return result.OnlyInLocal[i].Name < result.OnlyInLocal[j].Name
	})
	
	sort.Slice(result.OnlyInRemote, func(i, j int) bool {
		return result.OnlyInRemote[i].Name < result.OnlyInRemote[j].Name
	})
	
	sort.Slice(result.Different, func(i, j int) bool {
		return result.Different[i].Name < result.Different[j].Name
	})
	
	sort.Slice(result.Same, func(i, j int) bool {
		return result.Same[i].Name < result.Same[j].Name
	})
	
	return result
}

// GetDefaultStardewValleyModsPath 获取默认的星露谷物语Mods路径
func GetDefaultStardewValleyModsPath() string {
	// Windows 默认路径
	if home, err := os.UserHomeDir(); err == nil {
		// Windows
		windowsPath := filepath.Join(home, "AppData", "Roaming", "StardewValley", "Mods")
		if _, err := os.Stat(windowsPath); err == nil {
			return windowsPath
		}
		
		// macOS
		macPath := filepath.Join(home, "Library", "Application Support", "StardewValley", "Mods")
		if _, err := os.Stat(macPath); err == nil {
			return macPath
		}
		
		// Linux
		linuxPath := filepath.Join(home, ".local", "share", "StardewValley", "Mods")
		if _, err := os.Stat(linuxPath); err == nil {
			return linuxPath
		}
		
		// Steam Deck/Linux Flatpak
		flatpakPath := filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam", "steamapps", "compatdata", "413150", "pfx", "drive_c", "users", "steamuser", "AppData", "Roaming", "StardewValley", "Mods")
		if _, err := os.Stat(flatpakPath); err == nil {
			return flatpakPath
		}
	}
	
	return ""
}

// FormatComparisonResult 格式化Mod对比结果，用于显示给用户
func FormatComparisonResult(comparison ModComparison) string {
	var sb strings.Builder
	
	if len(comparison.OnlyInLocal) > 0 {
		sb.WriteString("只在本地存在的Mod:\n")
		for _, mod := range comparison.OnlyInLocal {
			sb.WriteString(fmt.Sprintf("  - %s (%s, %d bytes)\n", mod.Name, mod.Checksum[:8], mod.Size))
		}
		sb.WriteString("\n")
	}
	
	if len(comparison.OnlyInRemote) > 0 {
		sb.WriteString("只在远程存在的Mod:\n")
		for _, mod := range comparison.OnlyInRemote {
			sb.WriteString(fmt.Sprintf("  - %s (%s, %d bytes)\n", mod.Name, mod.Checksum[:8], mod.Size))
		}
		sb.WriteString("\n")
	}
	
	if len(comparison.Different) > 0 {
		sb.WriteString("版本不同的Mod:\n")
		for _, diff := range comparison.Different {
			sb.WriteString(fmt.Sprintf("  - %s:\n", diff.Name))
			sb.WriteString(fmt.Sprintf("    本地: %s (%d bytes)\n", diff.Local.Checksum[:8], diff.Local.Size))
			sb.WriteString(fmt.Sprintf("    远程: %s (%d bytes)\n", diff.Remote.Checksum[:8], diff.Remote.Size))
		}
		sb.WriteString("\n")
	}
	
	if len(comparison.Same) > 0 {
		sb.WriteString(fmt.Sprintf("相同的Mod (%d个):\n", len(comparison.Same)))
		for _, mod := range comparison.Same {
			sb.WriteString(fmt.Sprintf("  - %s\n", mod.Name))
		}
	}
	
	if sb.Len() == 0 {
		return "没有找到Mod文件"
	}
	
	return sb.String()
}