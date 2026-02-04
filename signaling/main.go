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
	clientID string
	isHost   bool
	lastSeen time.Time
}

// RoomInfo 表示一个房间的信息
type RoomInfo struct {
	ID        string
	CreatedAt time.Time
	Host      *Connection
	Clients   map[string]*Connection
	// 缓存主机发送的消息，以便新连接的客户端能立即收到
	PendingMessages []Message
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

	// 检查房间是否存在
	mu.RLock()
	room, roomExists := rooms[connectionID]
	mu.RUnlock()
	
	if !roomExists {
		sendError(conn, "Room not found")
		return
	}

	// 如果是主机，检查是否已有主机
	if joinMsg.IsHost {
		mu.Lock()
		if room.Host != nil {
			mu.Unlock()
			sendError(conn, "Room already has a host")
			return
		}
		mu.Unlock()
	}

	// 生成唯一的客户端ID
	clientID := fmt.Sprintf("%s-%d", connectionID, time.Now().UnixNano())

	// 注册连接
	connection := &Connection{
		conn:     conn,
		roomID:   connectionID,
		clientID: clientID,
		isHost:   joinMsg.IsHost,
		lastSeen: time.Now(),
	}

	mu.Lock()
	// 添加到全局连接映射
	connections[clientID] = connection
	
	// 添加到房间
	if joinMsg.IsHost {
		room.Host = connection
		log.Printf("Host connected to room %s (clientID: %s)\n", connectionID, clientID)
	} else {
		room.Clients[clientID] = connection
		log.Printf("Client connected to room %s (clientID: %s)\n", connectionID, clientID)
		
		// 如果有主机，通知主机有新客户端
		if room.Host != nil {
			notifyHostNewClient(room.Host, clientID)
		}
		
		// 发送缓存的pending消息给新客户端（分批发送，避免WebSocket过载）
		log.Printf("Sending %d pending messages to new client", len(room.PendingMessages))
		
		// 分批发送，每条消息之间有点延迟
		for i, msg := range room.PendingMessages {
			log.Printf("  -> Sending pending message %d/%d: %s", i+1, len(room.PendingMessages), msg.Type)
			
			if err := connection.conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send pending message %d to client: %v", i+1, err)
				break
			}
			
			// 小延迟，避免WebSocket过载
			time.Sleep(50 * time.Millisecond)
		}
	}
	mu.Unlock()

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
			log.Printf("Connection %s closed: %v\n", clientID, err)
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
	delete(connections, clientID)
	
	// 从房间中移除
	if room, exists := rooms[connectionID]; exists {
		if joinMsg.IsHost {
			room.Host = nil
			log.Printf("Host disconnected from room %s\n", connectionID)
			
			// 通知所有客户端主机已断开
			for _, client := range room.Clients {
				msg := Message{
					Type: "host_disconnected",
					Data: json.RawMessage(`{}`),
				}
				if err := client.conn.WriteJSON(msg); err != nil {
					log.Printf("Failed to notify client about host disconnect: %v\n", err)
				}
			}
		} else {
			delete(room.Clients, clientID)
			log.Printf("Client disconnected from room %s\n", connectionID)
			
			// 通知主机客户端已断开
			if room.Host != nil {
				msg := Message{
					Type: "client_disconnected",
					Data: json.RawMessage(fmt.Sprintf(`{"client_id": "%s"}`, clientID)),
				}
				if err := room.Host.conn.WriteJSON(msg); err != nil {
					log.Printf("Failed to notify host about client disconnect: %v\n", err)
				}
			}
		}
		
		// 如果房间为空，清理房间
		if room.Host == nil && len(room.Clients) == 0 {
			delete(rooms, connectionID)
			log.Printf("Room %s cleaned up (empty)\n", connectionID)
		}
	}
	
	mu.Unlock()
	log.Printf("Connection removed: %s\n", clientID)
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

	log.Printf("Forwarding offer from %s (host: %v) to room %s", 
		conn.clientID, conn.isHost, conn.roomID)
	
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

	log.Printf("Forwarding answer from %s (host: %v) to host in room %s", 
		conn.clientID, conn.isHost, conn.roomID)
	
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

	log.Printf("Forwarding ICE candidate from %s (host: %v) in room %s", 
		conn.clientID, conn.isHost, conn.roomID)
	
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

	room, exists := rooms[roomID]
	if !exists {
		return
	}

	// 转发给房间内的所有其他连接
	if sender.isHost {
		// 如果是主机发送的，缓存消息以便新客户端连接时能收到
		room.PendingMessages = append(room.PendingMessages, msg)
		log.Printf("Cached %s message from host in room %s (total cached: %d)", 
			msg.Type, roomID, len(room.PendingMessages))
		
		// 转发给所有已连接的客户端
		log.Printf("Forwarding %s from host to %d clients in room %s", 
			msg.Type, len(room.Clients), roomID)
		
		for clientID, client := range room.Clients {
			if client.clientID != sender.clientID {
				log.Printf("  -> Sending to client %s", clientID)
				if err := client.conn.WriteJSON(msg); err != nil {
					log.Printf("Failed to forward message to client %s: %v\n", client.clientID, err)
				}
			}
		}
	} else {
		// 如果是客户端发送的，转发给主机
		log.Printf("Forwarding %s from client to host in room %s", msg.Type, roomID)
		
		if room.Host != nil && room.Host.clientID != sender.clientID {
			log.Printf("  -> Sending to host %s", room.Host.clientID)
			if err := room.Host.conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to forward message to host %s: %v\n", room.Host.clientID, err)
			}
		}
		// 也可以选择转发给其他客户端（如果需要P2P）
	}
}

func forwardToHost(roomID string, sender *Connection, msg Message) {
	mu.RLock()
	defer mu.RUnlock()

	room, exists := rooms[roomID]
	if !exists {
		return
	}

	// 转发给主机
	if room.Host != nil && room.Host.clientID != sender.clientID {
		if err := room.Host.conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to forward message to host %s: %v\n", room.Host.clientID, err)
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
		ID:              connectionID,
		CreatedAt:       time.Now(),
		Host:            nil,
		Clients:         make(map[string]*Connection),
		PendingMessages: make([]Message, 0),
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
	room, exists := rooms[connectionID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 检查房间是否有主机连接（房间可能已创建但主机还未连接）
	mu.RLock()
	hasHost := room.Host != nil
	mu.RUnlock()
	
	response := map[string]interface{}{
		"status": "room_exists",
		"code":   connectionID,
		"ready":  hasHost,  // 房间是否就绪（有主机连接）
	}
	
	if !hasHost {
		response["message"] = "Room exists but host not connected yet"
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// notifyHostNewClient 通知主机有新客户端连接
func notifyHostNewClient(host *Connection, clientID string) {
	msg := Message{
		Type: "client_connected",
		Data: json.RawMessage(fmt.Sprintf(`{"client_id": "%s"}`, clientID)),
	}
	
	if err := host.conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to notify host about new client: %v\n", err)
	}
}

func cleanupConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		now := time.Now()
		
		// 清理过期的连接
		for clientID, conn := range connections {
			if now.Sub(conn.lastSeen) > 10*time.Minute {
				conn.conn.Close()
				delete(connections, clientID)
				
				// 从房间中移除
				if room, exists := rooms[conn.roomID]; exists {
					if conn.isHost {
						room.Host = nil
						log.Printf("Removed host from room %s: %s\n", conn.roomID, clientID)
					} else {
						delete(room.Clients, clientID)
						log.Printf("Removed client from room %s: %s\n", conn.roomID, clientID)
					}
					
					// 如果房间为空，清理房间
					if room.Host == nil && len(room.Clients) == 0 {
						delete(rooms, conn.roomID)
						log.Printf("Cleaned up empty room: %s\n", conn.roomID)
					}
				}
				
				log.Printf("Cleaned up stale connection: %s\n", clientID)
			}
		}
		
		// 清理过期的空房间
		for roomID, room := range rooms {
			if now.Sub(room.CreatedAt) > 30*time.Minute && room.Host == nil && len(room.Clients) == 0 {
				delete(rooms, roomID)
				log.Printf("Cleaned up expired empty room: %s\n", roomID)
			}
		}
		
		mu.Unlock()
	}
}