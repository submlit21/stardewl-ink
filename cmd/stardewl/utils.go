package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// getConnectionCodeFromServer 从信令服务器获取连接码
func getConnectionCodeFromServer(signalingURL string) (string, error) {
	// 从WebSocket URL提取HTTP URL
	baseURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	baseURL = strings.Replace(baseURL, "wss://", "https://", 1)
	baseURL = strings.Split(baseURL, "/ws")[0]
	
	createURL := baseURL + "/create"
	
	// 发送POST请求创建房间
	resp, err := http.Post(createURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("无法连接到信令服务器: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务器返回错误: %s", resp.Status)
	}
	
	// 解析响应
	var result struct {
		Code string `json:"code"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析服务器响应失败: %w", err)
	}
	
	if result.Code == "" {
		return "", fmt.Errorf("服务器返回空的连接码")
	}
	
	return result.Code, nil
}

// checkRoomExists 检查房间是否存在
func checkRoomExists(signalingURL, connectionID string) (bool, error) {
	// 从WebSocket URL提取HTTP URL
	baseURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	baseURL = strings.Replace(baseURL, "wss://", "https://", 1)
	baseURL = strings.Split(baseURL, "/ws")[0]
	
	checkURL := fmt.Sprintf("%s/join/%s", baseURL, connectionID)
	
	// 发送GET请求检查房间
	resp, err := http.Get(checkURL)
	if err != nil {
		return false, fmt.Errorf("无法连接到信令服务器: %w", err)
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK, nil
}