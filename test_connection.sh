#!/bin/bash

echo "🔧 测试信令服务器连接..."
echo "=================================="

# 清理之前的进程
pkill -f stardewl-signaling 2>/dev/null
sleep 1

# 启动信令服务器
echo "启动信令服务器..."
./dist/stardewl-signaling > /tmp/server.log 2>&1 &
SERVER_PID=$!
sleep 2

echo "服务器PID: $SERVER_PID"

# 测试服务器健康状态
echo -e "\n测试服务器健康状态..."
curl -s http://localhost:8080/health | python3 -m json.tool

# 测试创建房间
echo -e "\n测试创建房间..."
ROOM_CODE=$(curl -s -X POST http://localhost:8080/create | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
echo "创建的房间代码: $ROOM_CODE"

# 测试房间存在检查
echo -e "\n测试房间存在检查..."
curl -s http://localhost:8080/join/$ROOM_CODE | python3 -m json.tool

# 测试无效房间
echo -e "\n测试无效房间..."
curl -s http://localhost:8080/join/999999 | head -1

# 显示服务器日志
echo -e "\n服务器日志:"
tail -20 /tmp/server.log

# 清理
echo -e "\n清理..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\n✅ 测试完成！"