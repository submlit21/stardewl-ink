#!/bin/bash

echo "🔧 诊断P2P连接问题..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | tee /tmp/server_diag.log &
SERVER_PID=$!
sleep 3

echo "服务器PID: $SERVER_PID"

# 启动主机
echo -e "\n2. 启动主机..."
timeout 30 ./dist/stardewl --host 2>&1 | tee /tmp/host_diag.log &
HOST_PID=$!
sleep 8

echo "主机PID: $HOST_PID"

# 从主机输出提取连接码
HOST_ROOM_CODE=$(grep "连接码:" /tmp/host_diag.log | grep -o '[0-9]\{6\}' || echo "")
echo "主机连接码: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 启动客户端..."
    echo "注意：观察客户端是否收到和处理offer"
    timeout 20 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | tee /tmp/client_diag.log &
    CLIENT_PID=$!
    sleep 15
else
    echo -e "\n3. 跳过客户端测试（无效连接码）"
fi

# 显示关键诊断信息
echo -e "\n4. 诊断信息:"
echo "-------------------服务器关键日志-------------------"
grep -E "(Cached|Sending.*pending|Client connected|Forwarding.*offer|Forwarding.*answer)" /tmp/server_diag.log | tail -20

echo -e "\n-------------------主机关键日志-------------------"
grep -E "(Creating|Offer|Answer|ICE|connected|failed|error)" /tmp/host_diag.log | tail -20

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n-------------------客户端关键日志-------------------"
    grep -E "(received|offer|answer|ICE|connected|failed|error|Warning)" /tmp/client_diag.log | tail -30
fi

# 检查WebSocket错误
echo -e "\n5. 检查WebSocket错误:"
if grep -q "websocket: close 1006" /tmp/server_diag.log; then
    echo "❌ 发现WebSocket 1006错误（异常关闭）"
    echo "可能原因：客户端处理消息时崩溃"
else
    echo "✅ 未发现WebSocket 1006错误"
fi

# 检查消息处理
echo -e "\n6. 消息处理状态:"
SERVER_SENT=$(grep -c "Sending pending message" /tmp/server_diag.log)
CLIENT_RECEIVED=$(grep -c "Signaling client received message" /tmp/client_diag.log 2>/dev/null || echo "0")

echo "服务器发送消息数: $SERVER_SENT"
echo "客户端收到消息数: $CLIENT_RECEIVED"

if [ "$SERVER_SENT" -gt 0 ] && [ "$CLIENT_RECEIVED" -eq 0 ]; then
    echo "❌ 问题：服务器发送了消息但客户端没有收到"
    echo "可能原因：WebSocket连接问题或消息格式错误"
elif [ "$CLIENT_RECEIVED" -gt 0 ]; then
    echo "✅ 客户端收到了消息"
    
    # 检查是否处理了offer
    if grep -q "Client received offer" /tmp/client_diag.log 2>/dev/null; then
        echo "✅ 客户端收到了offer"
    else
        echo "❌ 客户端没有收到或没有处理offer"
    fi
    
    # 检查是否发送了answer
    if grep -q "Answer sent" /tmp/client_diag.log 2>/dev/null; then
        echo "✅ 客户端发送了answer"
    else
        echo "❌ 客户端没有发送answer"
    fi
fi

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID $CLIENT_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 诊断完成！"
echo -e "\n📋 建议下一步:"
echo "1. 查看完整日志了解详细信息"
echo "2. 如果客户端没有收到消息，检查WebSocket连接"
echo "3. 如果收到消息但没有处理，检查消息格式和处理器"