#!/bin/bash

echo "🔍 测试房间验证功能..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling 2>&1 | grep -v "Forwarding ICE" &
SERVER_PID=$!
sleep 3

# 测试不存在的房间
echo -e "\n2. 测试连接不存在的房间 (999999)..."
timeout 5 ./dist/stardewl --join=999999 2>&1 | tee /tmp/client_test.log

echo -e "\n3. 检查结果..."
if grep -q "❌ 房间不存在" /tmp/client_test.log; then
    echo "✅ 验证成功：正确检测到房间不存在"
elif grep -q "✅ 房间验证通过" /tmp/client_test.log; then
    echo "❌ 验证失败：不应该显示验证通过"
else
    echo "⚠️  未知结果"
    cat /tmp/client_test.log
fi

# 创建真实房间测试
echo -e "\n4. 创建真实房间测试..."
timeout 8 ./dist/stardewl --host 2>&1 | grep "连接码" &
HOST_PID=$!
sleep 5

# 获取连接码
ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "真实房间连接码: $ROOM_CODE"

if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n5. 测试连接真实房间..."
    timeout 5 ./dist/stardewl --join=$ROOM_CODE 2>&1 | tee /tmp/client_real.log
    
    if grep -q "✅ 房间验证通过" /tmp/client_real.log; then
        echo "✅ 验证成功：正确检测到房间存在"
    else
        echo "❌ 验证失败"
        cat /tmp/client_real.log
    fi
fi

# 清理
echo -e "\n6. 清理..."
kill $SERVER_PID $HOST_PID 2>/dev/null
echo "✅ 测试完成"