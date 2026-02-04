package core

import (
	"log"
	"time"
)

// 在P2PConnector结构体添加字段
// heartbeatTicker *time.Ticker
// stopHeartbeat   chan bool

// 添加startHeartbeat方法
func (p *P2PConnector) startHeartbeat() {
	p.heartbeatTicker = time.NewTicker(30 * time.Second)
	p.stopHeartbeat = make(chan bool)
	
	go func() {
		for {
			select {
			case <-p.heartbeatTicker.C:
				if p.IsConnected() {
					if err := p.signalingClient.SendMessage("ping", map[string]string{
						"timestamp": time.Now().Format(time.RFC3339),
					}); err != nil {
						log.Printf("⚠️ 发送心跳失败: %v", err)
					} else {
						log.Printf("💓 发送心跳 (room: %s)", p.roomID)
					}
				}
			case <-p.stopHeartbeat:
				return
			}
		}
	}()
	
	log.Printf("✅ 心跳机制已启动 (room: %s)", p.roomID)
}

// 添加stopHeartbeat方法
func (p *P2PConnector) stopHeartbeat() {
	if p.heartbeatTicker != nil {
		p.heartbeatTicker.Stop()
	}
	if p.stopHeartbeat != nil {
		close(p.stopHeartbeat)
	}
	log.Printf("🛑 心跳机制已停止 (room: %s)", p.roomID)
}
