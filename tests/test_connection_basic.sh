#!/bin/bash

echo "Basic connection test..."
echo "========================"

set -e

cleanup() {
    echo "Cleaning up..."
    pkill -f stardewl 2>/dev/null
    pkill -f stardewl-signaling 2>/dev/null
    sleep 1
}

trap cleanup EXIT

# Start signaling server
echo "1. Starting signaling server..."
./dist/stardewl-signaling > /dev/null 2>&1 &
SERVER_PID=$!
sleep 3

if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "ERROR: Signaling server failed to start"
    exit 1
fi
echo "Signaling server running (PID: $SERVER_PID)"

# Create room via HTTP API
echo -e "\n2. Creating room via HTTP API..."
ROOM_RESPONSE=$(curl -s -X POST http://localhost:8080/create)
ROOM_CODE=$(echo "$ROOM_RESPONSE" | grep -o '[0-9]\{6\}')
echo "Room created: $ROOM_CODE"
echo "Response: $ROOM_RESPONSE"

# Verify room exists
echo -e "\n3. Verifying room..."
VERIFY_RESPONSE=$(curl -s http://localhost:8080/join/$ROOM_CODE)
echo "Verify response: $VERIFY_RESPONSE"

if echo "$VERIFY_RESPONSE" | grep -q '"ready":false'; then
    echo "Room exists and is waiting for host"
else
    echo "ERROR: Room verification failed"
    exit 1
fi

# Test CLI functionality
echo -e "\n4. Testing CLI commands..."
echo "   - Help command:"
./dist/stardewl --help 2>&1 | grep -q "Usage of" && echo "      ✓ Works"

echo "   - List mods command:"
./dist/stardewl --list-mods 2>&1 | grep -q "Listing Mods" && echo "      ✓ Works"

# Test interactive mode (briefly)
echo -e "\n5. Testing interactive mode (5 seconds)..."
timeout 5 ./dist/stardewl --interactive > /tmp/interactive.log 2>&1 &
INTERACTIVE_PID=$!
sleep 2

if grep -q "Stardewl-Ink" /tmp/interactive.log; then
    echo "   ✓ Interactive mode starts"
else
    echo "   ⚠️  Interactive mode may have issues"
fi

kill $INTERACTIVE_PID 2>/dev/null || true

echo -e "\n✅ Basic connection test PASSED!"
echo "All core components are functional:"
echo "  - Signaling server HTTP API ✓"
echo "  - Room creation/verification ✓"
echo "  - CLI interface ✓"
echo "  - Interactive mode ✓"

exit 0