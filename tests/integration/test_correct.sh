#!/bin/bash

echo "🔧 正确测试P2P连接..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | grep -v "Forwarding ICE" &
SERVER_PID=$!
sleep 3

# 启动主机并捕获输出
echo -e "\n2. 启动主机..."
HOST_OUTPUT=$(timeout 15 ./dist/stardewl --host 2>&1 | tee /tmp/host.out)

# 从输出中提取连接码
ROOM_CODE=$(echo "$HOST_OUTPUT" | grep "连接码:" | grep -o '[0-9]\{6\}')
echo "实际连接码: $ROOM_CODE"

if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 启动客户端连接到正确的房间 $ROOM_CODE ..."
    ROOM_ID=$ROOM_CODE timeout 15 ./dist/test-client 2>&1 | grep -E "(Connecting|connected|received|answer|ICE|Processing|offer|error)"
else
    echo "❌ 无法获取有效连接码"
fi

# 显示主机ICE状态
echo -e "\n4. 主机ICE状态:"
grep -E "(ICE Connection State|answer|connected)" /tmp/host.out | tail -10

# 清理
kill $SERVER_PID 2>/dev/null
echo -e "\n✅ 测试完成！"
