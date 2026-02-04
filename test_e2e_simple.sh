#!/bin/bash

echo "🔍 端到端简单测试..."
echo "=================================="

# 清理
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | grep -v "Forwarding ICE" &
SERVER_PID=$!
sleep 3

# 测试1: 直接HTTP创建和验证
echo -e "\n2. 测试HTTP API..."
ROOM_CODE=$(curl -s -X POST http://localhost:8080/create | grep -o '[0-9]\{6\}')
echo "创建的房间码: $ROOM_CODE"
curl -s http://localhost:8080/join/$ROOM_CODE | grep -o '"ready":[^,]*'

# 测试2: 主机创建房间
echo -e "\n3. 启动主机..."
timeout 10 ./dist/stardewl --host 2>&1 | grep -E "(连接码|创建房间)" &
HOST_PID=$!
sleep 6

# 获取主机创建的房间码
HOST_ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "主机创建的房间码: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n4. 验证主机创建的房间..."
    curl -s http://localhost:8080/join/$HOST_ROOM_CODE | grep -o '"ready":[^,]*'
    
    echo -e "\n5. 客户端连接测试..."
    timeout 5 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | grep -E "(验证|房间|连接)"
fi

# 清理
echo -e "\n6. 清理..."
kill $SERVER_PID $HOST_PID 2>/dev/null
echo "✅ 测试完成"