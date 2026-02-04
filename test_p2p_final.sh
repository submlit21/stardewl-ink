#!/bin/bash

echo "🚀 P2P连接最终测试..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling &
SERVER_PID=$!
sleep 3

# 启动主机
echo -e "\n2. 启动主机..."
./dist/stardewl --host 2>&1 | grep -E "(连接码|ICE Connection State|answer|connected)" &
HOST_PID=$!
sleep 8

# 获取连接码
ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "主机连接码: $ROOM_CODE"

if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 启动客户端连接到房间 $ROOM_CODE ..."
    ROOM_ID=$ROOM_CODE ./dist/test-client 2>&1 | grep -E "(Connecting|connected|received|answer|ICE|Processing)"
    
    echo -e "\n4. 等待连接建立..."
    sleep 10
else
    echo -e "\n3. 跳过客户端测试"
fi

# 清理
kill $SERVER_PID $HOST_PID 2>/dev/null
echo -e "\n✅ 测试完成！"
