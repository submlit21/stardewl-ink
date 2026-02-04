#!/bin/bash
echo "快速修复编译错误..."

# 1. 修复ICE候选日志
sed -i 's/log.Printf("ICE candidate sent: %s:%d", candidate.Address, candidate.Port)/log.Printf("ICE candidate sent")/' core/p2p_connector.go

# 2. 修复stopHeartbeat变量名冲突
sed -i 's/p.stopHeartbeat = make(chan bool)/p.stopHeartbeatChan = make(chan bool)/' core/p2p_connector.go
sed -i 's/<-p.stopHeartbeat/<-p.stopHeartbeatChan/' core/p2p_connector.go
sed -i 's/close(p.stopHeartbeat)/close(p.stopHeartbeatChan)/' core/p2p_connector.go
sed -i 's/if p.stopHeartbeat != nil/if p.stopHeartbeatChan != nil/' core/p2p_connector.go

# 3. 修复结构体字段名
sed -i 's/stopHeartbeat   chan bool/stopHeartbeatChan chan bool/' core/p2p_connector.go

# 4. 修复AddICECandidate参数
sed -i 's/p.connection.AddICECandidate(candidate)/p.connection.AddICECandidate(candidate.Candidate)/' core/p2p_connector.go

# 5. 删除未使用的变量
sed -i '/msgData, err := NewMessage/d' core/p2p_connector.go
sed -i '/if err != nil {/,/return fmt.Errorf("failed to create mods message: %w", err)/d' core/p2p_connector.go

echo "修复完成"
