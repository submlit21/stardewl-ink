package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
	}

	// 房间管理
	rooms = make(map[string]*RoomInfo)
	
	// 连接管理
	connections = make(map[string]*Connection)
	mu          sync.RWMutex
)

// Connection 表示一个WebSocket连接
type Connection struct {
	conn     *websocket.Conn
	roomID   string
	isHost   bool
	lastSeen time.Time
}

// RoomInfo 表示一个房间的信息
type RoomInfo struct {
	ID        string
	CreatedAt time.Time
	Host      *Connection
	Clients   map[string]*Connection
}

// Message 信令消息
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// OfferMessage Offer消息
type OfferMessage struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// AnswerMessage Answer消息
type AnswerMessage struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// ICECandidateMessage ICE候选消息
type ICECandidateMessage struct {
	ConnectionID string `json:"connection_id"`
	Candidate    string `json:"candidate"`
}

// JoinRoomMessage 加入房间消息
type JoinRoomMessage struct {
	ConnectionID string `json:"connection_id"`
	IsHost       bool   `json:"is_host"`
}

// ConnectionCodeMessage 连接码消息
type ConnectionCodeMessage struct {
	Code string `json:"code"`
}

// ErrorMessage 错误消息
type ErrorMessage struct {
	Error string `json:"error"`
}

func main() {
	// 启动清理goroutine
	go cleanupConnections()

	// 设置路由
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/create", handleCreateRoom)
	http.HandleFunc("/join/", handleJoinRoom)
	http.HandleFunc("/health", handleHealth)

	// 启动服务器
	port := ":8080"
	log.Printf("Signaling server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v\n", err)
		return
	}
	defer conn.Close()

	// 读取连接ID
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read connection ID: %v\n", err)
		return
	}

	var joinMsg JoinRoomMessage
	if err := json.Unmarshal(message, &joinMsg); err != nil {
		sendError(conn, "Invalid join message")
		return
	}

	connectionID := joinMsg.ConnectionID
	if connectionID == "" {
		sendError(conn, "Connection ID is required")
		return
	}

	// 注册连接
	connection := &Connection{
		conn:     conn,
		roomID:   connectionID,
		isHost:   joinMsg.IsHost,
		lastSeen: time.Now(),
	}

	mu.Lock()
	connections[connectionID] = connection
	mu.Unlock()

	log.Printf("Connection established: %s (host: %v)\n", connectionID, joinMsg.IsHost)

	// 发送连接成功消息
	successMsg := Message{
		Type: "connected",
		Data: json.RawMessage(`{"status": "connected"}`),
	}
	if err := conn.WriteJSON(successMsg); err != nil {
		log.Printf("Failed to send connected message: %v\n", err)
		return
	}

	// 处理消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection %s closed: %v\n", connectionID, err)
			break
		}

		connection.lastSeen = time.Now()

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message: %v\n", err)
			continue
		}

		handleMessage(connection, msg)
	}

	// 清理连接
	mu.Lock()
	delete(connections, connectionID)
	mu.Unlock()
	log.Printf("Connection removed: %s\n", connectionID)
}

func handleMessage(conn *Connection, msg Message) {
	switch msg.Type {
	case "offer":
		handleOffer(conn, msg.Data)
	case "answer":
		handleAnswer(conn, msg.Data)
	case "ice_candidate":
		handleICECandidate(conn, msg.Data)
	case "ping":
		handlePing(conn)
	default:
		log.Printf("Unknown message type: %s\n", msg.Type)
	}
}

func handleOffer(conn *Connection, data json.RawMessage) {
	var offerMsg OfferMessage
	if err := json.Unmarshal(data, &offerMsg); err != nil {
		log.Printf("Failed to parse offer: %v\n", err)
		return
	}

	// 转发给房间内的其他客户端
	forwardToRoom(conn.roomID, conn, Message{
		Type: "offer",
		Data: data,
	})
}

func handleAnswer(conn *Connection, data json.RawMessage) {
	var answerMsg AnswerMessage
	if err := json.Unmarshal(data, &answerMsg); err != nil {
		log.Printf("Failed to parse answer: %v\n", err)
		return
	}

	// 转发给主机
	forwardToHost(conn.roomID, conn, Message{
		Type: "answer",
		Data: data,
	})
}

func handleICECandidate(conn *Connection, data json.RawMessage) {
	var iceMsg ICECandidateMessage
	if err := json.Unmarshal(data, &iceMsg); err != nil {
		log.Printf("Failed to parse ICE candidate: %v\n", err)
		return
	}

	// 转发给房间内的其他客户端
	forwardToRoom(conn.roomID, conn, Message{
		Type: "ice_candidate",
		Data: data,
	})
}

func handlePing(conn *Connection) {
	// 更新最后活跃时间
	conn.lastSeen = time.Now()
	
	// 发送pong响应
	pongMsg := Message{
		Type: "pong",
		Data: json.RawMessage(`{}`),
	}
	
	if err := conn.conn.WriteJSON(pongMsg); err != nil {
		log.Printf("Failed to send pong: %v\n", err)
	}
}

func forwardToRoom(roomID string, sender *Connection, msg Message) {
	mu.RLock()
	defer mu.RUnlock()

	for id, conn := range connections {
		if id == roomID && conn != sender {
			if err := conn.conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to forward message to %s: %v\n", id, err)
			}
		}
	}
}

func forwardToHost(roomID string, sender *Connection, msg Message) {
	mu.RLock()
	defer mu.RUnlock()

	for id, conn := range connections {
		if id == roomID && conn != sender && conn.isHost {
			if err := conn.conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to forward message to host %s: %v\n", id, err)
			}
			break
		}
	}
}

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 生成唯一的连接码
	connectionID := generateUniqueConnectionCode()
	
	// 预创建房间（等待主机连接）
	mu.Lock()
	// 只创建房间记录，不创建连接
	rooms[connectionID] = &RoomInfo{
		ID:        connectionID,
		CreatedAt: time.Now(),
		Host:      nil,
		Clients:   make(map[string]*Connection),
	}
	mu.Unlock()
	
	response := ConnectionCodeMessage{
		Code: connectionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	log.Printf("Room created: %s (waiting for host)\n", connectionID)
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	connectionID := r.URL.Path[len("/join/"):]
	if connectionID == "" {
		http.Error(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	mu.RLock()
	_, exists := connections[connectionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "room_exists",
		"code":   connectionID,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"connections": len(connections),
	})
}

func generateUniqueConnectionCode() string {
	// 生成6位数字连接码
	rand.Seed(time.Now().UnixNano())
	
	for {
		code := fmt.Sprintf("%06d", rand.Intn(1000000))
		
		// 检查是否已存在
		mu.RLock()
		_, exists := rooms[code]
		mu.RUnlock()
		
		if !exists {
			return code
		}
		
		// 如果代码已存在，重试
		time.Sleep(time.Millisecond)
	}
}

func sendError(conn *websocket.Conn, errorMsg string) {
	msg := Message{
		Type: "error",
		Data: json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, errorMsg)),
	}
	
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send error message: %v\n", err)
	}
}

func cleanupConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for id, conn := range connections {
			if now.Sub(conn.lastSeen) > 10*time.Minute {
				conn.conn.Close()
				delete(connections, id)
				log.Printf("Cleaned up stale connection: %s\n", id)
			}
		}
		mu.Unlock()
	}
}