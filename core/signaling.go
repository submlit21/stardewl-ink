package core

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SignalingClient 信令客户端
type SignalingClient struct {
	conn          *websocket.Conn
	url           string
	roomID        string
	isHost        bool
	onMessage     func(msgType string, data []byte)
	onConnected   func()
	onError       func(err error)
	mu            sync.RWMutex
	closed        bool
}

// NewSignalingClient 创建新的信令客户端
func NewSignalingClient(url, roomID string, isHost bool) (*SignalingClient, error) {
	// 建立WebSocket连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to signaling server: %w", err)
	}

	client := &SignalingClient{
		conn:    conn,
		url:     url,
		roomID:  roomID,
		isHost:  isHost,
		closed:  false,
	}

	// 发送加入消息
	joinMsg := map[string]interface{}{
		"connection_id": roomID,
		"is_host":       isHost,
	}
	
	if err := conn.WriteJSON(joinMsg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send join message: %w", err)
	}

	// 启动消息处理协程
	go client.handleMessages()

	return client, nil
}

// handleMessages 处理来自信令服务器的消息
func (c *SignalingClient) handleMessages() {
	defer func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !c.isClosed() {
				log.Printf("Signaling connection closed: %v", err)
				if c.onError != nil {
					c.onError(err)
				}
			}
			return
		}

		// 解析消息
		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse signaling message: %v", err)
			continue
		}

		// 处理连接成功消息
		if msg.Type == "connected" {
			log.Printf("Connected to signaling server for room: %s", c.roomID)
			if c.onConnected != nil {
				c.onConnected()
			}
			continue
		}

		// 转发给消息处理器
		if c.onMessage != nil {
			c.onMessage(msg.Type, msg.Data)
		}
	}
}

// SendMessage 发送消息到信令服务器
func (c *SignalingClient) SendMessage(msgType string, data interface{}) error {
	if c.isClosed() {
		return fmt.Errorf("signaling client is closed")
	}

	msg := map[string]interface{}{
		"type": msgType,
		"data": data,
	}

	return c.conn.WriteJSON(msg)
}

// Close 关闭信令客户端
func (c *SignalingClient) Close() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	
	return c.conn.Close()
}

// isClosed 检查客户端是否已关闭
func (c *SignalingClient) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// SetCallbacks 设置回调函数
func (c *SignalingClient) SetCallbacks(
	onMessage func(msgType string, data []byte),
	onConnected func(),
	onError func(err error),
) {
	c.onMessage = onMessage
	c.onConnected = onConnected
	c.onError = onError
}

// WaitForConnection 等待连接建立
func (c *SignalingClient) WaitForConnection(timeout time.Duration) bool {
	connected := make(chan bool, 1)
	
	originalOnConnected := c.onConnected
	c.onConnected = func() {
		if originalOnConnected != nil {
			originalOnConnected()
		}
		connected <- true
	}
	
	select {
	case <-connected:
		return true
	case <-time.After(timeout):
		return false
	}
}