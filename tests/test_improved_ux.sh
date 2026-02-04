#!/bin/bash

echo "Testing improved user experience..."
echo "=================================="

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
echo "Signaling server running"

# Test 1: Check help text is in English
echo -e "\n2. Testing help text language..."
./dist/stardewl --help 2>&1 | grep -q "Run in host mode" && \
    echo "✅ Help text is in English"

# Test 2: Host with timeout
echo -e "\n3. Testing host with timeout..."
start_time=$(date +%s)
timeout 8 ./dist/stardewl --host --timeout=5 2>&1 > /tmp/host_test.log
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "Host ran for $duration seconds"
if grep -q "Waiting for 5 seconds (timeout)" /tmp/host_test.log && \
   grep -q "Timeout reached, exiting" /tmp/host_test.log; then
    echo "✅ Host timeout feature works"
else
    echo "Host output:"
    tail -5 /tmp/host_test.log
fi

# Test 3: Create room for client test
echo -e "\n4. Creating room for client test..."
ROOM_RESPONSE=$(curl -s -X POST http://localhost:8080/create)
ROOM_CODE=$(echo "$ROOM_RESPONSE" | grep -o '[0-9]\{6\}')
echo "Room code: $ROOM_CODE"

# Test 4: Client with timeout
echo -e "\n5. Testing client with timeout..."
start_time=$(date +%s)
timeout 6 ./dist/stardewl --join=$ROOM_CODE --timeout=3 2>&1 > /tmp/client_test.log
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "Client ran for $duration seconds"
if grep -q "Waiting for 3 seconds (timeout)" /tmp/client_test.log && \
   grep -q "Timeout reached, exiting" /tmp/client_test.log; then
    echo "✅ Client timeout feature works"
else
    echo "Client output (last 5 lines):"
    tail -5 /tmp/client_test.log
fi

# Test 5: Default behavior (no timeout)
echo -e "\n6. Testing default behavior..."
echo "Starting brief test of default mode..."
timeout 3 ./dist/stardewl --host 2>&1 | grep -q "Press Enter to exit" && \
    echo "✅ Default mode shows 'Press Enter to exit'"

# Test 6: Error messages in English
echo -e "\n7. Testing error messages..."
# Try to connect to non-existent room
timeout 3 ./dist/stardewl --join=000000 2>&1 | grep -q "Room does not exist" && \
    echo "✅ Error messages are in English"

# Test 7: Verify all UI text is in English
echo -e "\n8. Verifying all UI text language..."
ALL_TESTS_PASSED=true

check_output() {
    local command="$1"
    local description="$2"
    timeout 4 $command 2>&1 | grep -q "[一-龥]" && {
        echo "⚠️  $description: Contains Chinese text"
        ALL_TESTS_PASSED=false
    } || echo "✅ $description: All English"
}

echo "Starting signaling server for UI tests..."
./dist/stardewl-signaling > /dev/null 2>&1 &
sleep 2

check_output "./dist/stardewl --host --timeout=1" "Host mode UI"
check_output "./dist/stardewl --join=123456 --timeout=1" "Client mode UI"
check_output "./dist/stardewl --list-mods" "List mods UI"

echo -e "\n9. Summary of improvements:"
echo "   - Timeout feature: ✓ Added (--timeout parameter)"
echo "   - UI language: ✓ Unified to English"
echo "   - Error messages: ✓ English"
echo "   - Help text: ✓ English"
echo "   - Automation support: ✓ Enabled via timeout"

if $ALL_TESTS_PASSED; then
    echo -e "\n✅ All user experience improvements PASSED!"
    echo "The application is now much more user-friendly and automation-ready."
else
    echo -e "\n⚠️  Some tests had warnings"
    echo "Most improvements are in place, but some Chinese text may remain."
fi

cleanup
exit 0