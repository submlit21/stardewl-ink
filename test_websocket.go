package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	// 先创建房间
	resp, err := http.Post("http://localhost:8080/create", "application/json", nil)
	if err != nil {
		log.Fatal("创建房间失败:", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("解析响应失败:", err)
	}

	roomCode := result.Code
	fmt.Printf("创建的房间代码: %s\n", roomCode)

	// 尝试WebSocket连接
	origin := "http://localhost/"
	url := "ws://localhost:8080/ws"
	
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal("WebSocket连接失败:", err)
	}
	defer ws.Close()

	// 发送加入消息
	joinMsg := map[string]interface{}{
		"connection_id": roomCode,
		"is_host":       false,
	}
	
	msgBytes, _ := json.Marshal(joinMsg)
	if _, err := ws.Write(msgBytes); err != nil {
		log.Fatal("发送消息失败:", err)
	}
	fmt.Println("已发送加入消息")

	// 尝试接收响应
	var response []byte
	if err := websocket.Message.Receive(ws, &response); err != nil {
		log.Fatal("接收响应失败:", err)
	}
	
	fmt.Printf("收到响应: %s\n", response)
	
	// 等待一下看是否有更多消息
	time.Sleep(2 * time.Second)
}