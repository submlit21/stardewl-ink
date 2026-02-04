#!/bin/bash

echo "🚀 最终P2P连接测试..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | tee /tmp/server_final.log &
SERVER_PID=$!
sleep 3

echo "服务器PID: $SERVER_PID"

# 启动主机
echo -e "\n2. 启动主机..."
./dist/stardewl --host 2>&1 | tee /tmp/host_final.log &
HOST_PID=$!
sleep 8

echo "主机PID: $HOST_PID"

# 从主机输出提取连接码
HOST_ROOM_CODE=$(grep "连接码:" /tmp/host_final.log | grep -o '[0-9]\{6\}' || echo "")
echo "主机连接码: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n3. 启动简化客户端..."
    # 使用简化客户端
    ROOM_ID=$HOST_ROOM_CODE ./dist/test-client 2>&1 | tee /tmp/client_final.log
    
    echo -e "\n4. 等待5秒..."
    sleep 5
else
    echo -e "\n3. 跳过客户端测试（无效连接码）"
fi

# 显示关键日志
echo -e "\n5. 关键日志分析:"
echo "-------------------服务器日志-------------------"
grep -E "(Client connected|Sending.*pending|Forwarding.*answer|ICE candidate)" /tmp/server_final.log | tail -20

echo -e "\n-------------------主机日志-------------------"
grep -E "(ICE Connection State|answer|connected|failed)" /tmp/host_final.log | tail -20

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n-------------------客户端日志-------------------"
    grep -E "(Connecting|connected|received|answer|ICE)" /tmp/client_final.log | tail -20
fi

# 检查ICE连接
echo -e "\n6. ICE连接状态:"
if grep -q "ICE Connection State has changed: connected" /tmp/host_final.log; then
    echo "✅ ICE连接已建立！"
elif grep -q "ICE Connection State has changed: failed" /tmp/host_final.log; then
    echo "❌ ICE连接失败"
elif grep -q "ICE Connection State has changed: checking" /tmp/host_final.log; then
    echo "🔄 ICE连接检查中..."
else
    echo "⚠️  未找到ICE连接状态"
fi

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 测试完成！"