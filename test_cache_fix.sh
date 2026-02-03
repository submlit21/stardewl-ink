#!/bin/bash

echo "🔧 测试消息缓存修复..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | tee /tmp/server_cache.log &
SERVER_PID=$!
sleep 3

echo "服务器PID: $SERVER_PID"

# 启动主机（先启动，发送offer）
echo -e "\n2. 启动主机（先启动）..."
timeout 20 ./dist/stardewl --host 2>&1 | tee /tmp/host_cache.log &
HOST_PID=$!
sleep 5

echo "主机PID: $HOST_PID"

# 从主机输出提取连接码
HOST_ROOM_CODE=$(grep "连接码:" /tmp/host_cache.log | grep -o '[0-9]\{6\}' || echo "")
echo "主机连接码: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 等待主机发送offer和ICE候选..."
    sleep 5
    
    echo -e "\n4. 启动客户端（后启动，应该收到缓存的offer）..."
    timeout 15 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | tee /tmp/client_cache.log &
    CLIENT_PID=$!
    sleep 10
else
    echo -e "\n3. 跳过客户端测试（无效连接码）"
fi

# 显示关键日志
echo -e "\n5. 关键日志分析:"
echo "-------------------服务器日志（关键部分）-------------------"
grep -E "(Cached|Forwarding|Sending.*pending|Client connected|ICE candidate)" /tmp/server_cache.log | tail -30

echo -e "\n-------------------主机日志-------------------"
grep -E "(Creating|Offer|ICE|Answer|connected)" /tmp/host_cache.log | tail -20

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n-------------------客户端日志-------------------"
    grep -E "(received|answer|ICE|connected|failed)" /tmp/client_cache.log | tail -20
fi

# 检查是否有ICE连接建立
echo -e "\n6. 检查ICE连接状态:"
if grep -q "ICE connection established" /tmp/host_cache.log /tmp/client_cache.log 2>/dev/null; then
    echo "✅ ICE连接已建立！"
else
    echo "❌ ICE连接未建立"
    echo "检查日志中的错误信息..."
fi

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID $CLIENT_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 测试完成！"