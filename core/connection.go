package core

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v3"
)

// Connection 表示一个WebRTC连接
type Connection struct {
	peerConnection *webrtc.PeerConnection
	dataChannel   *webrtc.DataChannel
	connectionID  string
	isHost        bool
	onMessage     func([]byte)
	onClose       func()
	mu            sync.RWMutex
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	ICEServers []webrtc.ICEServer
}

// NewConnection 创建新的WebRTC连接
func NewConnection(connectionID string, isHost bool, config ConnectionConfig) (*Connection, error) {
	// 创建PeerConnection配置
	peerConfig := webrtc.Configuration{
		ICEServers: config.ICEServers,
	}

	// 创建PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(peerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	conn := &Connection{
		peerConnection: peerConnection,
		connectionID:   connectionID,
		isHost:         isHost,
	}

	// 设置ICE连接状态回调
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", state.String())
		
		if state == webrtc.ICEConnectionStateDisconnected ||
			state == webrtc.ICEConnectionStateFailed ||
			state == webrtc.ICEConnectionStateClosed {
			conn.close()
		}
	})

	// 如果是主机，创建数据通道
	if isHost {
		dataChannel, err := peerConnection.CreateDataChannel("stardewl", nil)
		if err != nil {
			peerConnection.Close()
			return nil, fmt.Errorf("failed to create data channel: %w", err)
		}
		
		conn.setupDataChannel(dataChannel)
		conn.dataChannel = dataChannel
	} else {
		// 如果是客户端，监听数据通道
		peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
			conn.setupDataChannel(dc)
			conn.dataChannel = dc
			log.Printf("Data channel '%s' opened\n", dc.Label())
		})
	}

	return conn, nil
}

// setupDataChannel 设置数据通道的回调
func (c *Connection) setupDataChannel(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		log.Printf("Data channel '%s' opened\n", dc.Label())
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		c.mu.RLock()
		onMessage := c.onMessage
		c.mu.RUnlock()
		
		if onMessage != nil {
			onMessage(msg.Data)
		}
	})

	dc.OnClose(func() {
		log.Printf("Data channel '%s' closed\n", dc.Label())
		c.close()
	})
}

// CreateOffer 创建SDP Offer（主机调用）
func (c *Connection) CreateOffer() (string, error) {
	if !c.isHost {
		return "", fmt.Errorf("only host can create offer")
	}

	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create offer: %w", err)
	}

	err = c.peerConnection.SetLocalDescription(offer)
	if err != nil {
		return "", fmt.Errorf("failed to set local description: %w", err)
	}

	// 等待ICE收集完成
	gatherComplete := webrtc.GatheringCompletePromise(c.peerConnection)
	<-gatherComplete

	offerJSON, err := json.Marshal(c.peerConnection.LocalDescription())
	if err != nil {
		return "", fmt.Errorf("failed to marshal offer: %w", err)
	}

	return string(offerJSON), nil
}

// SetRemoteDescription 设置远程SDP描述
func (c *Connection) SetRemoteDescription(sdp string) error {
	var desc webrtc.SessionDescription
	if err := json.Unmarshal([]byte(sdp), &desc); err != nil {
		return fmt.Errorf("failed to unmarshal SDP: %w", err)
	}

	return c.peerConnection.SetRemoteDescription(desc)
}

// CreateAnswer 创建SDP Answer（客户端调用）
func (c *Connection) CreateAnswer() (string, error) {
	if c.isHost {
		return "", fmt.Errorf("only client can create answer")
	}

	answer, err := c.peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create answer: %w", err)
	}

	err = c.peerConnection.SetLocalDescription(answer)
	if err != nil {
		return "", fmt.Errorf("failed to set local description: %w", err)
	}

	// 等待ICE收集完成
	gatherComplete := webrtc.GatheringCompletePromise(c.peerConnection)
	<-gatherComplete

	answerJSON, err := json.Marshal(c.peerConnection.LocalDescription())
	if err != nil {
		return "", fmt.Errorf("failed to marshal answer: %w", err)
	}

	return string(answerJSON), nil
}

// AddICECandidate 添加ICE候选
func (c *Connection) AddICECandidate(candidate string) error {
	var iceCandidate webrtc.ICECandidateInit
	if err := json.Unmarshal([]byte(candidate), &iceCandidate); err != nil {
		return fmt.Errorf("failed to unmarshal ICE candidate: %w", err)
	}

	return c.peerConnection.AddICECandidate(iceCandidate)
}

// SendMessage 发送消息到对端
func (c *Connection) SendMessage(data []byte) error {
	c.mu.RLock()
	dc := c.dataChannel
	c.mu.RUnlock()

	if dc == nil {
		return fmt.Errorf("data channel not ready")
	}

	if dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel not open")
	}

	return dc.Send(data)
}

// SendJSON 发送JSON消息到对端
func (c *Connection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendMessage(data)
}

// SetMessageHandler 设置消息处理回调
func (c *Connection) SetMessageHandler(handler func([]byte)) {
	c.mu.Lock()
	c.onMessage = handler
	c.mu.Unlock()
}

// SetCloseHandler 设置关闭回调
func (c *Connection) SetCloseHandler(handler func()) {
	c.mu.Lock()
	c.onClose = handler
	c.mu.Unlock()
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.close()
	return nil
}

// close 内部关闭方法
func (c *Connection) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peerConnection != nil {
		c.peerConnection.Close()
		c.peerConnection = nil
	}

	if c.onClose != nil {
		c.onClose()
	}
}

// ConnectionID 获取连接ID
func (c *Connection) ConnectionID() string {
	return c.connectionID
}

// IsHost 检查是否是主机
func (c *Connection) IsHost() bool {
	return c.isHost
}

// IsConnected 检查是否已连接
func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.dataChannel == nil {
		return false
	}
	return c.dataChannel.ReadyState() == webrtc.DataChannelStateOpen
}