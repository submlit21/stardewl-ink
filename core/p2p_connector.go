package core

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

// P2PConnector P2P连接器
type P2PConnector struct {
	signalingClient *SignalingClient
	connection      *Connection
	roomID          string
	isHost          bool
	modsPath        string
	onModsChecked   func(ModComparison)
	onConnected     func()
	onDisconnected  func()
	mu              sync.RWMutex
	connected       bool
}

// P2PConfig P2P配置
type P2PConfig struct {
	SignalingURL string
	RoomID       string
	IsHost       bool
	ModsPath     string
	ICEServers   []webrtc.ICEServer
}

// NewP2PConnector 创建新的P2P连接器
func NewP2PConnector(config P2PConfig) (*P2PConnector, error) {
	// 创建信令客户端
	signalingClient, err := NewSignalingClient(config.SignalingURL, config.RoomID, config.IsHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create signaling client: %w", err)
	}

	// 创建WebRTC连接配置
	connConfig := ConnectionConfig{
		ICEServers: config.ICEServers,
	}

	// 创建WebRTC连接
	connection, err := NewConnection(config.RoomID, config.IsHost, connConfig)
	if err != nil {
		signalingClient.Close()
		return nil, fmt.Errorf("failed to create WebRTC connection: %w", err)
	}

	connector := &P2PConnector{
		signalingClient: signalingClient,
		connection:      connection,
		roomID:          config.RoomID,
		isHost:          config.IsHost,
		modsPath:        config.ModsPath,
		connected:       false,
	}

	// 设置信令客户端回调
	signalingClient.SetCallbacks(
		connector.handleSignalingMessage,
		connector.handleSignalingConnected,
		connector.handleSignalingError,
	)

	// 设置WebRTC连接回调
	connection.SetMessageHandler(connector.handleDataChannelMessage)
	connection.SetCloseHandler(connector.handleConnectionClose)

	return connector, nil
}

// Start 启动P2P连接
func (p *P2PConnector) Start() error {
	// 等待信令连接建立
	if !p.signalingClient.WaitForConnection(10 * time.Second) {
		return fmt.Errorf("failed to establish signaling connection")
	}

	log.Printf("Signaling connection established for room: %s", p.roomID)

	// 如果是主机，创建并发送offer
	if p.isHost {
		return p.startAsHost()
	}

	// 客户端等待offer
	return p.startAsClient()
}

// startAsHost 作为主机启动
func (p *P2PConnector) startAsHost() error {
	log.Printf("Creating WebRTC offer as host...")
	
	// 创建offer
	offer, err := p.connection.CreateOffer()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	// 发送offer到信令服务器
	if err := p.signalingClient.SendMessage("offer", map[string]string{
		"offer": offer,
	}); err != nil {
		return fmt.Errorf("failed to send offer: %w", err)
	}

	log.Printf("Offer sent to signaling server")
	return nil
}

// startAsClient 作为客户端启动
func (p *P2PConnector) startAsClient() error {
	log.Printf("Waiting for offer from host...")
	return nil
}

// handleSignalingMessage 处理信令消息
func (p *P2PConnector) handleSignalingMessage(msgType string, data []byte) {
	switch msgType {
	case "offer":
		p.handleOffer(data)
	case "answer":
		p.handleAnswer(data)
	case "ice_candidate":
		p.handleICECandidate(data)
	case "client_connected":
		log.Printf("New client connected to room")
	case "host_disconnected":
		log.Printf("Host disconnected from room")
		p.handleDisconnection()
	case "client_disconnected":
		log.Printf("Client disconnected from room")
		p.handleDisconnection()
	case "error":
		var errorData struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(data, &errorData); err == nil {
			log.Printf("Signaling error: %s", errorData.Error)
		}
	default:
		log.Printf("Unknown signaling message type: %s", msgType)
	}
}

// handleOffer 处理收到的Offer
func (p *P2PConnector) handleOffer(data []byte) {
	if p.isHost {
		log.Printf("Host received offer, ignoring")
		return
	}

	log.Printf("Client received offer from host")
	
	var offerData struct {
		Offer string `json:"offer"`
	}
	if err := json.Unmarshal(data, &offerData); err != nil {
		log.Printf("Failed to parse offer: %v", err)
		return
	}

	// 设置远程描述
	if err := p.connection.SetRemoteDescription(offerData.Offer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	// 创建answer
	answer, err := p.connection.CreateAnswer()
	if err != nil {
		log.Printf("Failed to create answer: %v", err)
		return
	}

	// 发送answer到信令服务器
	if err := p.signalingClient.SendMessage("answer", map[string]string{
		"answer": answer,
	}); err != nil {
		log.Printf("Failed to send answer: %v", err)
	}

	log.Printf("Answer sent to signaling server")
}

// handleAnswer 处理收到的Answer
func (p *P2PConnector) handleAnswer(data []byte) {
	if !p.isHost {
		log.Printf("Client received answer, ignoring")
		return
	}

	log.Printf("Host received answer from client")
	
	var answerData struct {
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(data, &answerData); err != nil {
		log.Printf("Failed to parse answer: %v", err)
		return
	}

	// 设置远程描述
	if err := p.connection.SetRemoteDescription(answerData.Answer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	log.Printf("Remote description set successfully")
}

// handleICECandidate 处理ICE候选
func (p *P2PConnector) handleICECandidate(data []byte) {
	var iceData struct {
		Candidate string `json:"candidate"`
	}
	if err := json.Unmarshal(data, &iceData); err != nil {
		log.Printf("Failed to parse ICE candidate: %v", err)
		return
	}

	if err := p.connection.AddICECandidate(iceData.Candidate); err != nil {
		log.Printf("Failed to add ICE candidate: %v", err)
	}
}

// handleSignalingConnected 处理信令连接建立
func (p *P2PConnector) handleSignalingConnected() {
	log.Printf("Signaling connection fully established")
}

// handleSignalingError 处理信令错误
func (p *P2PConnector) handleSignalingError(err error) {
	log.Printf("Signaling error: %v", err)
	p.handleDisconnection()
}

// handleDataChannelMessage 处理数据通道消息
func (p *P2PConnector) handleDataChannelMessage(data []byte) {
	msg, err := ParseMessage(data)
	if err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case MessageTypeModsList:
		p.handleModsList(msg.Payload)
	case MessageTypeModsComparison:
		p.handleModsComparison(msg.Payload)
	case MessageTypePing:
		p.handlePing()
	case MessageTypeGameReady:
		p.handleGameReady()
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleModsList 处理Mod列表
func (p *P2PConnector) handleModsList(payload json.RawMessage) {
	modsMsg, err := ParseModsList(payload)
	if err != nil {
		log.Printf("Failed to parse mods list: %v", err)
		return
	}

	// 扫描本地Mods
	localMods, err := ScanMods(p.modsPath)
	if err != nil {
		log.Printf("Failed to scan local mods: %v", err)
		return
	}

	// 比较Mods
	comparison := CompareMods(localMods, modsMsg.Mods)

	// 发送比较结果
	comparisonMsg := ModsComparisonMessage{
		Comparison: comparison,
	}
	comparisonData, _ := json.Marshal(comparisonMsg)
	
	msg := Message{
		Type:    MessageTypeModsComparison,
		Payload: comparisonData,
	}
	
	msgData, _ := json.Marshal(msg)
	p.connection.SendMessage(msgData)

	// 调用回调
	if p.onModsChecked != nil {
		p.onModsChecked(comparison)
	}
}

// handleModsComparison 处理Mod比较结果
func (p *P2PConnector) handleModsComparison(payload json.RawMessage) {
	var comparisonMsg ModsComparisonMessage
	if err := json.Unmarshal(payload, &comparisonMsg); err != nil {
		log.Printf("Failed to parse mods comparison: %v", err)
		return
	}

	comparison := comparisonMsg.Comparison
	
	log.Printf("Mods comparison received:")
	log.Printf("  Only in local: %d", len(comparison.OnlyInLocal))
	log.Printf("  Only in remote: %d", len(comparison.OnlyInRemote))
	log.Printf("  Different: %d", len(comparison.Different))
	log.Printf("  Same: %d", len(comparison.Same))

	// 调用回调
	if p.onModsChecked != nil {
		p.onModsChecked(comparison)
	}
}

// handlePing 处理心跳
func (p *P2PConnector) handlePing() {
	// 发送pong响应
	pongMsg := Message{
		Type: MessageTypePong,
	}
	pongData, _ := json.Marshal(pongMsg)
	p.connection.SendMessage(pongData)
}

// handleGameReady 处理游戏就绪
func (p *P2PConnector) handleGameReady() {
	log.Printf("Remote peer is ready to play")
}

// handleConnectionClose 处理连接关闭
func (p *P2PConnector) handleConnectionClose() {
	log.Printf("WebRTC connection closed")
	p.handleDisconnection()
}

// handleDisconnection 处理断开连接
func (p *P2PConnector) handleDisconnection() {
	p.mu.Lock()
	if p.connected {
		p.connected = false
		if p.onDisconnected != nil {
			p.onDisconnected()
		}
	}
	p.mu.Unlock()
}

// SendModsList 发送Mod列表
func (p *P2PConnector) SendModsList() error {
	if !p.connection.IsConnected() {
		return fmt.Errorf("not connected")
	}

	mods, err := ScanMods(p.modsPath)
	if err != nil {
		return fmt.Errorf("failed to scan mods: %w", err)
	}

	modsMsg := ModsListMessage{
		Mods: mods,
	}
	modsData, _ := json.Marshal(modsMsg)
	
	msg := Message{
		Type:    MessageTypeModsList,
		Payload: modsData,
	}
	
	msgData, _ := json.Marshal(msg)
	return p.connection.SendMessage(msgData)
}

// SetCallbacks 设置回调函数
func (p *P2PConnector) SetCallbacks(
	onModsChecked func(ModComparison),
	onConnected func(),
	onDisconnected func(),
) {
	p.onModsChecked = onModsChecked
	p.onConnected = onConnected
	p.onDisconnected = onDisconnected
}

// Close 关闭P2P连接器
func (p *P2PConnector) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.signalingClient != nil {
		p.signalingClient.Close()
	}
	
	if p.connection != nil {
		p.connection.Close()
	}
	
	p.connected = false
}

// IsConnected 检查是否已连接
func (p *P2PConnector) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected && p.connection.IsConnected()
}