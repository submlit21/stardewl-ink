#!/bin/bash
echo "🚀 快速测试P2P连接回调..."
pkill -f stardewl 2>/dev/null
sleep 1

# 启动服务器
./dist/stardewl-signaling 2>&1 | grep -E "(starting|Room created|Sending.*pending|Forwarding.*offer)" &
SERVER_PID=$!
sleep 3

# 启动主机
timeout 10 ./dist/stardewl --host 2>&1 | grep -E "(连接码|Creating|Offer)" &
HOST_PID=$!
sleep 5

# 获取连接码
ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "连接码: $ROOM_CODE"

if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo "启动客户端，观察回调日志..."
    timeout 8 ./dist/stardewl --join=$ROOM_CODE 2>&1 | grep -E "(received|handleSignalingMessage|Processing|Answer|ICE)"
    sleep 2
fi

kill $SERVER_PID $HOST_PID 2>/dev/null
echo "测试完成"
