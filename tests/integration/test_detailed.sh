#!/bin/bash

echo "🔧 详细P2P连接诊断测试..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器（详细日志）
echo "1. 启动信令服务器（详细日志）..."
./dist/stardewl-signaling 2>&1 | tee /tmp/server_detailed.log &
SERVER_PID=$!
sleep 3

echo "服务器PID: $SERVER_PID"

# 测试主机模式
echo -e "\n2. 启动主机（详细日志）..."
timeout 15 ./dist/stardewl --host 2>&1 | tee /tmp/host_detailed.log &
HOST_PID=$!
sleep 5

echo "主机PID: $HOST_PID"

# 从主机输出提取连接码
HOST_ROOM_CODE=$(grep "连接码:" /tmp/host_detailed.log | grep -o '[0-9]\{6\}' || echo "")
echo "主机连接码: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 启动客户端（详细日志，连接码: $HOST_ROOM_CODE）..."
    timeout 10 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | tee /tmp/client_detailed.log &
    CLIENT_PID=$!
    sleep 8
else
    echo -e "\n3. 跳过客户端测试（无效连接码）"
fi

# 显示关键日志
echo -e "\n4. 关键日志分析:"
echo "-------------------服务器日志-------------------"
grep -E "(Room created|Host connected|Client connected|Forwarding|ICE candidate)" /tmp/server_detailed.log | tail -20

echo -e "\n-------------------主机日志-------------------"
grep -E "(Creating|Offer|ICE|connected|failed|error)" /tmp/host_detailed.log | tail -20

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n-------------------客户端日志-------------------"
    grep -E "(Waiting|offer|answer|ICE|connected|failed|error)" /tmp/client_detailed.log | tail -20
fi

# 检查WebSocket连接
echo -e "\n5. 检查WebSocket连接状态..."
if netstat -tuln 2>/dev/null | grep -q ":8080"; then
    echo "✅ 服务器端口8080监听中"
else
    echo "❌ 服务器端口未监听"
fi

# 检查进程
echo -e "\n6. 进程状态:"
ps -ef | grep -E "(stardewl-signaling|stardewl)" | grep -v grep

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID $CLIENT_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 诊断测试完成！"
echo -e "\n📋 建议:"
echo "1. 查看完整日志文件:"
echo "   - 服务器: /tmp/server_detailed.log"
echo "   - 主机: /tmp/host_detailed.log"
echo "   - 客户端: /tmp/client_detailed.log"
echo "2. 检查是否有'ICE connection established'日志"
echo "3. 检查是否有'Forwarding'相关的信令消息"