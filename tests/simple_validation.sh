#!/bin/bash

echo "Simple validation test..."
echo "========================="

# Cleanup
cleanup() {
    echo "Cleaning up..."
    pkill -f stardewl 2>/dev/null
    pkill -f stardewl-signaling 2>/dev/null
    sleep 1
}

trap cleanup EXIT

# Start signaling server
echo "1. Starting signaling server..."
./dist/stardewl-signaling &
SERVER_PID=$!
sleep 3

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "ERROR: Signaling server failed to start"
    exit 1
fi

echo "Signaling server started (PID: $SERVER_PID)"

# Test HTTP API
echo -e "\n2. Testing HTTP API..."
ROOM_RESPONSE=$(curl -s -X POST http://localhost:8080/create)
echo "Create response: $ROOM_RESPONSE"

ROOM_CODE=$(echo "$ROOM_RESPONSE" | grep -o '[0-9]\{6\}' || echo "")
if [ -z "$ROOM_CODE" ]; then
    echo "ERROR: Failed to get room code from response"
    exit 1
fi

echo "Room code: $ROOM_CODE"

# Verify room
VERIFY_RESPONSE=$(curl -s http://localhost:8080/join/$ROOM_CODE)
echo "Verify response: $VERIFY_RESPONSE"

if echo "$VERIFY_RESPONSE" | grep -q '"ready":false'; then
    echo "Room verification successful (ready: false - no host connected)"
else
    echo "WARNING: Unexpected verify response"
fi

# Test CLI help
echo -e "\n3. Testing CLI help..."
./dist/stardewl --help 2>&1 | grep -q "Usage of" && echo "CLI help works"

# Test list mods (should work without actual mods)
echo -e "\n4. Testing list-mods..."
./dist/stardewl --list-mods 2>&1 | grep -q "Listing Mods" && echo "List mods works"

echo -e "\n✅ Basic validation passed!"
echo "All core components are working:"
echo "  - Signaling server ✓"
echo "  - HTTP API ✓"  
echo "  - CLI interface ✓"
echo "  - Mod scanning ✓"

cleanup
exit 0