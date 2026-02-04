#!/bin/bash

echo "End-to-end simple test..."
echo "=================================="

# Cleanup
pkill -f stardewl 2>/dev/null
sleep 1

# Start signaling server
echo "1. Starting signaling server..."
./dist/stardewl-signaling 2>&1 | grep -v "Forwarding ICE" &
SERVER_PID=$!
sleep 3

# Test 1: Direct HTTP create and verify
echo -e "\n2. Testing HTTP API..."
ROOM_CODE=$(curl -s -X POST http://localhost:8080/create | grep -o '[0-9]\{6\}')
echo "Created room code: $ROOM_CODE"
curl -s http://localhost:8080/join/$ROOM_CODE | grep -o '"ready":[^,]*'

# Test 2: Host creating room
echo -e "\n3. Starting host..."
timeout 15 ./dist/stardewl --host 2>&1 | grep -E "(Connection code|Creating room)" &
HOST_PID=$!
sleep 8

# Get host created room code
HOST_ROOM_CODE=$(ps aux | grep "stardewl --host" | grep -o "[0-9]\{6\}" | head -1)
echo "Host created room code: $HOST_ROOM_CODE"

if [[ "$HOST_ROOM_CODE" =~ ^[0-9]{6}$ ]]; then
    echo -e "\n4. Verifying host created room..."
    curl -s http://localhost:8080/join/$HOST_ROOM_CODE | grep -o '"ready":[^,]*'
    
    echo -e "\n5. Client connection test..."
    timeout 8 ./dist/stardewl --join=$HOST_ROOM_CODE 2>&1 | grep -E "(Verifying|Room|Connection)"
fi

# Cleanup
echo -e "\n6. Cleaning up..."
kill $SERVER_PID $HOST_PID 2>/dev/null
echo "Test completed"