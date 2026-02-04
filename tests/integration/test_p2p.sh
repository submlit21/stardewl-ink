#!/bin/bash

echo "🔧 测试P2P连接功能..."
echo "=================================="

# 清理
pkill -f stardewl-signaling 2>/dev/null
pkill -f stardewl 2>/dev/null
sleep 1

# 启动信令服务器
echo "1. 启动信令服务器..."
./dist/stardewl-signaling > /tmp/server_test.log 2>&1 &
SERVER_PID=$!
sleep 3

echo "服务器PID: $SERVER_PID"

# 测试服务器API
echo -e "\n2. 测试服务器API..."
echo "健康检查:"
curl -s http://localhost:8080/health | python3 -m json.tool

echo -e "\n创建房间:"
ROOM_CODE=$(curl -s -X POST http://localhost:8080/create | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
echo "房间代码: $ROOM_CODE"

echo -e "\n检查房间:"
curl -s http://localhost:8080/join/$ROOM_CODE | python3 -m json.tool

# 测试主机模式
echo -e "\n3. 测试主机模式..."
timeout 8 ./dist/stardewl --host > /tmp/host_test.log 2>&1 &
HOST_PID=$!
sleep 3

echo "主机输出:"
cat /tmp/host_test.log | head -10

# 从主机输出提取连接码
HOST_ROOM_CODE=$(grep "连接码:" /tmp/host_test.log | grep -o '[0-9]\{6\}' || echo "未找到")
echo "提取的连接码: $HOST_ROOM_CODE"

# 测试客户端模式
if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n4. 测试客户端模式 (连接码: $HOST_ROOM_CODE)..."
    timeout 5 ./dist/stardewl --join=$HOST_ROOM_CODE > /tmp/client_test.log 2>&1 &
    CLIENT_PID=$!
    sleep 3
    
    echo "客户端输出:"
    cat /tmp/client_test.log | head -10
else
    echo -e "\n4. 跳过客户端测试 (无效连接码)"
fi

# 显示服务器日志
echo -e "\n5. 服务器日志:"
tail -30 /tmp/server_test.log

# 清理
echo -e "\n6. 清理..."
kill $SERVER_PID $HOST_PID $CLIENT_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null 2>/dev/null

echo -e "\n✅ 测试完成！"