#!/bin/bash

echo "🔧 完整连接流程测试..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling > /tmp/server_full.log 2>&1 &
SERVER_PID=$!
sleep 3

echo "服务器运行中 (PID: $SERVER_PID)"

# 测试主机获取连接码
echo -e "\n2. 测试主机获取连接码..."
HOST_OUTPUT=$(timeout 5 ./dist/stardewl --host 2>&1 | head -10)
echo "$HOST_OUTPUT"

# 从输出中提取连接码
ROOM_CODE=$(echo "$HOST_OUTPUT" | grep "连接码:" | awk '{print $2}')
if [ -z "$ROOM_CODE" ]; then
    ROOM_CODE="未获取到"
fi
echo "提取的连接码: $ROOM_CODE"

# 如果获取到连接码，测试客户端加入
if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 测试客户端加入房间 $ROOM_CODE ..."
    timeout 5 ./dist/stardewl --join=$ROOM_CODE 2>&1 | head -10
else
    echo -e "\n3. 跳过客户端测试（未获取到有效连接码）"
fi

# 显示服务器日志
echo -e "\n4. 服务器日志:"
tail -30 /tmp/server_full.log

# 测试交互模式
echo -e "\n5. 测试交互模式（快速测试）..."
timeout 3 ./dist/stardewl --interactive 2>&1 | head -20

# 清理
echo -e "\n6. 清理..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 完整测试完成！"