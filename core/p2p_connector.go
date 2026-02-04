package core

import (
	"encoding/json"
	"fmt"
	"time"
	"log"
	"sync"

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
	// ICE候选队列：当远程描述未设置时缓存ICE候选
	pendingICECandidates []webrtc.ICECandidateInit
	pendingICEMu         sync.RWMutex
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
	// 先创建P2P连接器（但不立即创建信令客户端）
	connector := &P2PConnector{
		roomID:    config.RoomID,
		isHost:    config.IsHost,
		modsPath:  config.ModsPath,
		connected: false,
	}

	// 创建WebRTC连接配置
	connConfig := ConnectionConfig{
		ICEServers: config.ICEServers,
	}

	// 创建WebRTC连接
	connection, err := NewConnection(config.RoomID, config.IsHost, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebRTC connection: %w", err)
	}

	connector.connection = connection

	// 现在创建信令客户端（确保回调已经设置）
	signalingClient, err := NewSignalingClient(config.SignalingURL, config.RoomID, config.IsHost)
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("failed to create signaling client: %w", err)
	}

	connector.signalingClient = signalingClient

	// 设置ICE候选回调
	connection.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			log.Printf("ICE candidate gathering complete for %s", config.RoomID)
			return
		}

		candidateJSON, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Printf("Failed to marshal ICE candidate: %v", err)
			return
		}

		// 发送ICE候选到信令服务器
		if err := signalingClient.SendMessage("ice_candidate", map[string]string{
			"candidate": string(candidateJSON),
		}); err != nil {
			log.Printf("Failed to send ICE candidate: %v", err)
		} else {
			log.Printf("ICE candidate sent: %s:%d", candidate.Address, candidate.Port)
		}
	})

	// 设置信令客户端回调
	log.Printf("Setting signaling client callbacks for room: %s", config.RoomID)
	signalingClient.SetCallbacks(
		connector.handleSignalingMessage,
		connector.handleSignalingConnected,
		connector.handleSignalingError,
	)
	log.Printf("Signaling client callbacks set successfully")

	// 设置WebRTC连接回调
	connection.SetMessageHandler(connector.handleDataChannelMessage)
	connection.SetCloseHandler(connector.handleConnectionClose)

	return connector, nil
}

// Start starts the P2P connection
func (p *P2PConnector) Start() error {
	log.Printf("Starting P2P connection for room: %s (host: %v)", p.roomID, p.isHost)

	// Give signaling connection time to establish
	time.Sleep(2 * time.Second)

	log.Printf("Signaling connection established for room: %s", p.roomID)

	// 如果是主机，创建并发送offer
	if p.isHost {
		return p.startAsHost()
	}

	// 客户端等待offer
	return p.startAsClient()
}

// startAsHost starts as host
func (p *P2PConnector) startAsHost() error {
	log.Printf("Creating WebRTC Offer as host...")
	
	// Create offer
	offer, err := p.connection.CreateOffer()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}
	
	log.Printf("Offer created successfully, length: %d bytes", len(offer))

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
	log.Printf("Waiting for host offer...")
	return nil
}

// handleSignalingMessage 处理信令消息
func (p *P2PConnector) handleSignalingMessage(msgType string, data []byte) {
	// log.Printf("P2PConnector.handleSignalingMessage called! Type: %s, Data length: %d", msgType, len(data))

	switch msgType {
	case "offer":
		// log.Printf("Processing offer message")
		p.handleOffer(data)
	case "answer":
		// log.Printf("Processing answer message")
		p.handleAnswer(data)
	case "ice_candidate":
		// log.Printf("Processing ICE candidate message")
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

	log.Printf("Client received offer from host (data length: %d bytes)", len(data))

	var offerData struct {
		Offer string `json:"offer"`
	}
	if err := json.Unmarshal(data, &offerData); err != nil {
		log.Printf("Failed to parse offer: %v", err)
		log.Printf("Offer data (first 200 chars): %s", string(data)[:min(200, len(data))])
		return
	}

	if offerData.Offer == "" {
		log.Printf("Empty offer received")
		return
	}

	log.Printf("Setting remote description (offer length: %d chars)", len(offerData.Offer))

	// 设置远程描述
	if err := p.connection.SetRemoteDescription(offerData.Offer); err != nil {
		log.Printf(" Failed to set remote description: %v", err)
		return
	}

	log.Printf("Creating answer...")

	// 创建answer
	answer, err := p.connection.CreateAnswer()
	if err != nil {
		log.Printf(" Failed to create answer: %v", err)
		return
	}

	log.Printf("Sending answer (length: %d chars)", len(answer))

	// 发送answer到信令服务器
	if err := p.signalingClient.SendMessage("answer", map[string]string{
		"answer": answer,
	}); err != nil {
		log.Printf(" Failed to send answer: %v", err)
	} else {
		log.Printf(" Answer sent to signaling server")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

	// 处理缓存的ICE候选
	p.pendingICEMu.Lock()
	if len(p.pendingICECandidates) > 0 {
		log.Printf("处理 %d 个缓存的ICE候选", len(p.pendingICECandidates))
		for _, candidate := range p.pendingICECandidates {
			// 将ICECandidateInit转换为JSON字符串
			candidateJSON, err := json.Marshal(candidate)
			if err != nil {
				log.Printf("Failed to serialize ICE candidate: %v", err)
				continue
			}
			if err := p.connection.AddICECandidate(string(candidateJSON)); err != nil {
				log.Printf("Failed to add cached ICE candidate: %v", err)
			}
		}
		// 清空缓存
		p.pendingICECandidates = nil
	}
	p.pendingICEMu.Unlock()

	// 连接建立
	p.mu.Lock()
	p.connected = true
	p.mu.Unlock()
	log.Printf("P2P connection established")
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

	// 解析ICE候选
	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal([]byte(iceData.Candidate), &candidate); err != nil {
		log.Printf("Failed to unmarshal ICE candidate: %v", err)
		return
	}

	// 尝试添加ICE候选
	// 将ICECandidateInit转换为JSON字符串
	candidateJSON, err := json.Marshal(candidate)
	if err != nil {
		log.Printf("Failed to serialize ICE candidate: %v", err)
		return
	}

	if err := p.connection.AddICECandidate(string(candidateJSON)); err != nil {
		// 如果失败（可能是远程描述未设置），缓存起来
		log.Printf("ICE候选添加失败，缓存起来等待远程描述设置: %v", err)
		p.pendingICEMu.Lock()
		p.pendingICECandidates = append(p.pendingICECandidates, candidate)
		p.pendingICEMu.Unlock()
	} else {
		log.Printf("ICE candidate added successfully")
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