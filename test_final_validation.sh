#!/bin/bash

echo "🎯 最终验证测试..."
echo "=================================="

# 清理
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling &
SERVER_PID=$!
sleep 3

# 测试主机创建房间
echo -e "\n2. 启动主机（应该成功创建房间）..."
timeout 15 ./dist/stardewl --host 2>&1 | tee /tmp/host_final.log &
HOST_PID=$!
sleep 8

# 检查主机是否成功
echo -e "\n3. 检查主机状态..."
if grep -q "✅ 连接码:" /tmp/host_final.log; then
    HOST_ROOM_CODE=$(grep "✅ 连接码:" /tmp/host_final.log | grep -o '[0-9]\{6\}')
    echo "主机创建的房间码: $HOST_ROOM_CODE"
    
    echo -e "\n4. 验证房间..."
    curl -s http://localhost:8080/join/$HOST_ROOM_CODE
    
    echo -e "\n5. 客户端连接测试..."
    timeout 8 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | tee /tmp/client_final.log | grep -E "(验证|房间|连接|错误)"
    
    echo -e "\n6. 检查客户端结果..."
    if grep -q "✅ 房间验证通过" /tmp/client_final.log; then
        echo "✅ 客户端验证成功"
    elif grep -q "⚠️  房间存在但主机未连接" /tmp/client_final.log; then
        echo "⚠️  房间存在但主机未连接（可能需要更多时间）"
    else
        echo "❌ 客户端验证失败"
        cat /tmp/client_final.log
    fi
else
    echo "❌ 主机创建房间失败"
    cat /tmp/host_final.log
fi

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID 2>/dev/null
echo "✅ 测试完成"
