#!/bin/bash

echo "🔧 测试房间创建修复..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | grep -E "(starting|Room created|Host connected)" &
SERVER_PID=$!
sleep 3

# 测试HTTP API
echo -e "\n2. 测试HTTP API..."
curl -s http://localhost:8080/create-room | jq . 2>/dev/null || curl -s http://localhost:8080/create-room

# 启动主机
echo -e "\n3. 启动主机（应该创建房间）..."
timeout 8 ./dist/stardewl --host 2>&1 | grep -E "(创建房间|连接码|Connecting|connected|error|failed)" &
HOST_PID=$!
sleep 6

# 获取连接码
echo -e "\n4. 获取连接码..."
ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "连接码: $ROOM_CODE"

# 清理
kill $SERVER_PID $HOST_PID 2>/dev/null
echo -e "\n✅ 测试完成！"