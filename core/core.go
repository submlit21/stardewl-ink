package core

import (
	"fmt"
	"log"
	"time"

	"github.com/pion/webrtc/v3"
)

// StardewlClient 星露谷联机客户端
type StardewlClient struct {
	connection    *Connection
	signalingURL  string
	connectionID  string
	isHost        bool
	modsPath      string
	onModsChecked func(ModComparison)
	onConnected   func()
	onDisconnected func()
}

// ClientConfig 客户端配置
type ClientConfig struct {
	SignalingURL string
	ConnectionID string
	IsHost       bool
	ModsPath     string
	ICEServers   []webrtc.ICEServer
}

// NewStardewlClient 创建新的客户端
func NewStardewlClient(config ClientConfig) (*StardewlClient, error) {
	// 如果没有指定Mods路径，使用默认路径
	modsPath := config.ModsPath
	if modsPath == "" {
		modsPath = GetDefaultStardewValleyModsPath()
		if modsPath == "" {
			log.Println("Warning: Could not find default Stardew Valley Mods path")
		}
	}

	// 创建连接配置
	connConfig := ConnectionConfig{
		ICEServers: config.ICEServers,
	}

	// 创建WebRTC连接
	connection, err := NewConnection(config.ConnectionID, config.IsHost, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	client := &StardewlClient{
		connection:   connection,
		signalingURL: config.SignalingURL,
		connectionID: config.ConnectionID,
		isHost:       config.IsHost,
		modsPath:     modsPath,
	}

	// 设置消息处理器
	connection.SetMessageHandler(client.handleMessage)

	// 设置关闭处理器
	connection.SetCloseHandler(func() {
		if client.onDisconnected != nil {
			client.onDisconnected()
		}
	})

	return client, nil
}

// handleMessage 处理接收到的消息
func (c *StardewlClient) handleMessage(data []byte) {
	msg, err := ParseMessage(data)
	if err != nil {
		log.Printf("Failed to parse message: %v\n", err)
		return
	}

	switch msg.Type {
	case MessageTypeModsList:
		c.handleModsList(msg.Payload)
	case MessageTypeModsComparison:
		c.handleModsComparison(msg.Payload)
	case MessageTypePing:
		c.handlePing()
	case MessageTypeGameReady:
		c.handleGameReady()
	case MessageTypeError:
		c.handleError(msg.Payload)
	default:
		log.Printf("Unknown message type: %s\n", msg.Type)
	}
}

// handleModsList 处理Mod列表消息
func (c *StardewlClient) handleModsList(payload json.RawMessage) {
	modsMsg, err := ParseModsList(payload)
	if err != nil {
		log.Printf("Failed to parse mods list: %v\n", err)
		return
	}

	// 扫描本地Mods
	localMods, err := ScanMods(c.modsPath)
	if err != nil {
		log.Printf("Failed to scan local mods: %v\n", err)
		return
	}

	// 比较Mods
	comparison := CompareMods(localMods, modsMsg.Mods)

	// 发送对比结果给对端
	comparisonMsg := ModsComparisonMessage{
		Comparison: comparison,
	}
	
	if err := c.SendModsComparison(comparisonMsg); err != nil {
		log.Printf("Failed to send mods comparison: %v\n", err)
	}

	// 通知UI
	if c.onModsChecked != nil {
		c.onModsChecked(comparison)
	}
}

// handleModsComparison 处理Mod对比消息
func (c *StardewlClient) handleModsComparison(payload json.RawMessage) {
	comparisonMsg, err := ParseModsComparison(payload)
	if err != nil {
		log.Printf("Failed to parse mods comparison: %v\n", err)
		return
	}

	// 通知UI
	if c.onModsChecked != nil {
		c.onModsChecked(comparisonMsg.Comparison)
	}
}

// handlePing 处理心跳消息
func (c *StardewlClient) handlePing() {
	// 发送pong响应
	pongMsg, err := NewMessage(MessageTypePong, nil)
	if err != nil {
		log.Printf("Failed to create pong message: %v\n", err)
		return
	}
	
	if err := c.connection.SendMessage(pongMsg); err != nil {
		log.Printf("Failed to send pong: %v\n", err)
	}
}

// handleGameReady 处理游戏准备就绪消息
func (c *StardewlClient) handleGameReady() {
	log.Println("Remote peer is ready to play")
	// 可以在这里通知UI游戏可以开始了
}

// handleError 处理错误消息
func (c *StardewlClient) handleError(payload json.RawMessage) {
	errorMsg, err := ParseError(payload)
	if err != nil {
		log.Printf("Failed to parse error message: %v\n", err)
		return
	}
	
	log.Printf("Received error from peer: %s - %s\n", errorMsg.Code, errorMsg.Message)
}

// StartAsHost 作为主机启动
func (c *StardewlClient) StartAsHost() error {
	if !c.isHost {
		return fmt.Errorf("client is not configured as host")
	}

	// 创建offer
	offer, err := c.connection.CreateOffer()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	log.Printf("Created offer for connection %s\n", c.connectionID)
	
	// 在实际应用中，这里应该通过信令服务器发送offer
	// 为了简化，我们假设offer已经通过其他方式交换
	
	return nil
}

// ConnectAsClient 作为客户端连接
func (c *StardewlClient) ConnectAsClient(offer string) error {
	if c.isHost {
		return fmt.Errorf("client is configured as host")
	}

	// 设置远程offer
	if err := c.connection.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	// 创建answer
	answer, err := c.connection.CreateAnswer()
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	log.Printf("Created answer for connection %s\n", c.connectionID)
	
	// 在实际应用中，这里应该通过信令服务器发送answer
	// 为了简化，我们假设answer已经通过其他方式交换
	
	return nil
}

// SendModsList 发送Mod列表
func (c *StardewlClient) SendModsList() error {
	// 扫描本地Mods
	mods, err := ScanMods(c.modsPath)
	if err != nil {
		return fmt.Errorf("failed to scan mods: %w", err)
	}

	// 创建消息
	modsMsg := ModsListMessage{
		Mods: mods,
	}
	
	msg, err := NewMessage(MessageTypeModsList, modsMsg)
	if err != nil {
		return fmt.Errorf("failed to create mods list message: %w", err)
	}

	// 发送消息
	return c.connection.SendMessage(msg)
}

// SendModsComparison 发送Mod对比结果
func (c *StardewlClient) SendModsComparison(comparison ModsComparisonMessage) error {
	msg, err := NewMessage(MessageTypeModsComparison, comparison)
	if err != nil {
		return fmt.Errorf("failed to create mods comparison message: %w", err)
	}

	return c.connection.SendMessage(msg)
}

// SendGameReady 发送游戏准备就绪消息
func (c *StardewlClient) SendGameReady() error {
	msg, err := NewMessage(MessageTypeGameReady, nil)
	if err != nil {
		return fmt.Errorf("failed to create game ready message: %w", err)
	}

	return c.connection.SendMessage(msg)
}

// SendPing 发送心跳消息
func (c *StardewlClient) SendPing() error {
	msg, err := NewMessage(MessageTypePing, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping message: %w", err)
	}

	return c.connection.SendMessage(msg)
}

// StartHeartbeat 开始心跳检测
func (c *StardewlClient) StartHeartbeat(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			if c.connection.IsConnected() {
				if err := c.SendPing(); err != nil {
					log.Printf("Failed to send heartbeat: %v\n", err)
				}
			}
		}
	}()
}

// SetModsCheckedHandler 设置Mod检查回调
func (c *StardewlClient) SetModsCheckedHandler(handler func(ModComparison)) {
	c.onModsChecked = handler
}

// SetConnectedHandler 设置连接成功回调
func (c *StardewlClient) SetConnectedHandler(handler func()) {
	c.onConnected = handler
}

// SetDisconnectedHandler 设置断开连接回调
func (c *StardewlClient) SetDisconnectedHandler(handler func()) {
	c.onDisconnected = handler
}

// ConnectionID 获取连接ID
func (c *StardewlClient) ConnectionID() string {
	return c.connectionID
}

// IsHost 检查是否是主机
func (c *StardewlClient) IsHost() bool {
	return c.isHost
}

// IsConnected 检查是否已连接
func (c *StardewlClient) IsConnected() bool {
	return c.connection.IsConnected()
}

// Close 关闭客户端
func (c *StardewlClient) Close() error {
	return c.connection.Close()
}

// GetDefaultICEServers 获取默认的ICE服务器
func GetDefaultICEServers() []webrtc.ICEServer {
	return []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
		{
			URLs: []string{"stun:stun1.l.google.com:19302"},
		},
		{
			URLs: []string{"stun:stun2.l.google.com:19302"},
		},
		{
			URLs: []string{"stun:stun3.l.google.com:19302"},
		},
		{
			URLs: []string{"stun:stun4.l.google.com:19302"},
		},
	}
}