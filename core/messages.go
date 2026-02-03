package core

import "encoding/json"

// MessageType 消息类型
type MessageType string

const (
	// ModsList 发送Mod列表
	MessageTypeModsList MessageType = "mods_list"
	// ModsComparison 发送Mod对比结果
	MessageTypeModsComparison MessageType = "mods_comparison"
	// GameReady 游戏准备就绪
	MessageTypeGameReady MessageType = "game_ready"
	// Ping 心跳检测
	MessageTypePing MessageType = "ping"
	// Pong 心跳响应
	MessageTypePong MessageType = "pong"
	// Error 错误消息
	MessageTypeError MessageType = "error"
)

// Message 通用消息结构
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ModsListMessage Mod列表消息
type ModsListMessage struct {
	Mods []ModInfo `json:"mods"`
}

// ModsComparisonMessage Mod对比消息
type ModsComparisonMessage struct {
	Comparison ModComparison `json:"comparison"`
}

// ErrorMessage 错误消息
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, payload interface{}) ([]byte, error) {
	var rawPayload json.RawMessage
	var err error
	
	if payload != nil {
		rawPayload, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	
	msg := Message{
		Type:    msgType,
		Payload: rawPayload,
	}
	
	return json.Marshal(msg)
}

// ParseMessage 解析消息
func ParseMessage(data []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

// ParseModsList 解析Mod列表消息
func ParseModsList(data []byte) (ModsListMessage, error) {
	var msg ModsListMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

// ParseModsComparison 解析Mod对比消息
func ParseModsComparison(data []byte) (ModsComparisonMessage, error) {
	var msg ModsComparisonMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

// ParseError 解析错误消息
func ParseError(data []byte) (ErrorMessage, error) {
	var msg ErrorMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}