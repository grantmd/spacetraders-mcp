#!/bin/bash

# Test script to verify graceful shutdown behavior
# This script tests that Ctrl+C doesn't produce error messages

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Graceful Shutdown${NC}"
echo "================================"

# Build the server
echo -e "\n${YELLOW}Building server...${NC}"
go build -o spacetraders-mcp .

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Test graceful shutdown with a simple approach
echo -e "\n${YELLOW}Testing graceful shutdown behavior...${NC}"

# Start the server and send it a signal after a brief moment
(
    ./spacetraders-mcp &
    SERVER_PID=$!

    # Wait briefly for server to start
    sleep 1

    # Send SIGINT (Ctrl+C equivalent)
    kill -INT $SERVER_PID 2>/dev/null || true

    # Wait for it to finish
    wait $SERVER_PID 2>/dev/null || true

) 2>/tmp/shutdown_test_errors.txt

# Check for error messages
if [ -s /tmp/shutdown_test_errors.txt ]; then
    echo -e "${YELLOW}Messages during shutdown:${NC}"
    cat /tmp/shutdown_test_errors.txt

    # Check if it's just the expected context canceled error
    if grep -q "context canceled" /tmp/shutdown_test_errors.txt; then
        echo -e "${RED}✗ Still showing context canceled error${NC}"
        echo -e "${YELLOW}The fix may need adjustment${NC}"
    else
        echo -e "${GREEN}✓ No context canceled error (other messages may be normal)${NC}"
    fi
else
    echo -e "${GREEN}✓ Clean shutdown with no error messages${NC}"
fi

# Clean up
rm -f /tmp/shutdown_test_errors.txt spacetraders-mcp

echo -e "\n${GREEN}Shutdown test completed!${NC}"
echo -e "\n${YELLOW}Manual verification:${NC}"
echo "1. Run: ./spacetraders-mcp"
echo "2. Press Ctrl+C"
echo "3. Verify no 'Server error: context canceled' message appears"
echo "4. The server should exit cleanly"
