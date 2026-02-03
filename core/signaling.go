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
	// 消息队列：在回调设置前缓存消息
	messageQueue  []queuedMessage
	queueMu       sync.RWMutex
}

// queuedMessage 队列中的消息
type queuedMessage struct {
	msgType string
	data    []byte
}

// NewSignalingClient 创建新的信令客户端
func NewSignalingClient(url, roomID string, isHost bool) (*SignalingClient, error) {
	log.Printf("🔗 Connecting to signaling server: %s (room: %s, host: %v)", url, roomID, isHost)
	
	// 建立WebSocket连接（带重试）
	var conn *websocket.Conn
	var err error
	
	for i := 0; i < 3; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			break
		}
		
		log.Printf("⚠️  Connection attempt %d failed: %v", i+1, err)
		if i < 2 {
			time.Sleep(1 * time.Second)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to signaling server after 3 attempts: %w", err)
	}

	log.Printf("✅ WebSocket connection established to signaling server")
	
	client := &SignalingClient{
		conn:         conn,
		url:          url,
		roomID:       roomID,
		isHost:       isHost,
		closed:       false,
		messageQueue: make([]queuedMessage, 0),
	}

	// 发送加入消息
	joinMsg := map[string]interface{}{
		"connection_id": roomID,
		"is_host":       isHost,
	}
	
	log.Printf("📤 Sending join message for room: %s", roomID)
	if err := conn.WriteJSON(joinMsg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send join message: %w", err)
	}

	log.Printf("✅ Join message sent successfully")
	
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
		
		// 防止消息处理崩溃
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in message handling: %v", r)
				}
			}()
			
			// 解析消息
			var msg struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to parse signaling message: %v", err)
				return
			}

			log.Printf("Signaling client received message type: %s, data length: %d", 
				msg.Type, len(msg.Data))

			// 处理连接成功消息
			if msg.Type == "connected" {
				log.Printf("Connected to signaling server for room: %s", c.roomID)
				if c.onConnected != nil {
					c.onConnected()
				}
				return
			}

			// 转发给消息处理器
			c.queueMu.Lock()
			if c.onMessage != nil {
				log.Printf("Calling onMessage callback for type: %s", msg.Type)
				c.onMessage(msg.Type, msg.Data)
				
				// 如果有队列中的消息，也处理它们
				if len(c.messageQueue) > 0 {
					log.Printf("Processing %d queued messages", len(c.messageQueue))
					for _, qm := range c.messageQueue {
						log.Printf("  -> Processing queued message: %s", qm.msgType)
						c.onMessage(qm.msgType, qm.data)
					}
					// 清空队列
					c.messageQueue = make([]queuedMessage, 0)
				}
			} else {
				// 回调还没有设置，将消息加入队列
				log.Printf("📦 Queueing message (callback not set yet): %s", msg.Type)
				c.messageQueue = append(c.messageQueue, queuedMessage{
					msgType: msg.Type,
					data:    msg.Data,
				})
			}
			c.queueMu.Unlock()
		}()
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
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	
	c.onMessage = onMessage
	c.onConnected = onConnected
	c.onError = onError
	
	// 如果有队列中的消息，立即处理它们
	if onMessage != nil && len(c.messageQueue) > 0 {
		log.Printf("🔄 Processing %d queued messages after setting callbacks", len(c.messageQueue))
		for _, qm := range c.messageQueue {
			log.Printf("  -> Processing queued: %s", qm.msgType)
			onMessage(qm.msgType, qm.data)
		}
		// 清空队列
		c.messageQueue = make([]queuedMessage, 0)
	}
	
	log.Printf("✅ Callbacks set successfully for room: %s", c.roomID)
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
