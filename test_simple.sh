#!/bin/bash

echo "🔧 简单测试当前状态..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 检查文件
echo "1. 检查文件..."
ls -la dist/

# 启动信令服务器
echo -e "\n2. 启动信令服务器..."
./dist/stardewl-signaling &
SERVER_PID=$!
sleep 3

# 测试服务器是否运行
echo "3. 测试服务器..."
curl -s http://localhost:8080/ || echo "服务器可能没有运行"

# 启动主机（简化）
echo -e "\n4. 启动主机（简化）..."
timeout 10 ./dist/stardewl --host 2>&1 | head -20 &
HOST_PID=$!
sleep 5

# 获取连接码
echo -e "\n5. 获取连接码..."
ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "连接码: $ROOM_CODE"

if [[ "$ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n6. 启动客户端..."
    timeout 8 ./dist/stardewl --join=$ROOM_CODE 2>&1 | head -20
    sleep 2
fi

# 清理
echo -e "\n7. 清理..."
kill $SERVER_PID $HOST_PID 2>/dev/null

echo -e "\n✅ 测试完成！"