#!/bin/bash

# Test script for SpaceTraders MCP Server
# This script makes it easy to test the MCP server functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}SpaceTraders MCP Server Test Script${NC}"
echo "======================================"

# Build the server
echo -e "\n${YELLOW}Building server...${NC}"
go build -o spacetraders-mcp .

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Test 1: List available resources
echo -e "\n${YELLOW}Test 1: Listing available resources${NC}"
echo '{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}' | ./spacetraders-mcp | jq .

# Test 2: Read agent info resource
echo -e "\n${YELLOW}Test 2: Reading agent info resource${NC}"
RESULT=$(echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "spacetraders://agent/info"}}' | ./spacetraders-mcp)

# Check if the result contains an error
if echo "$RESULT" | jq -e '.error' > /dev/null 2>&1; then
    echo -e "${RED}Resource read failed:${NC}"
    echo "$RESULT" | jq .
else
    echo -e "${GREEN}Resource read successful:${NC}"
    echo "$RESULT" | jq .
fi

# Test 3: Read ships list resource
echo -e "\n${YELLOW}Test 3: Reading ships list resource${NC}"
RESULT=$(echo '{"jsonrpc": "2.0", "id": 3, "method": "resources/read", "params": {"uri": "spacetraders://ships/list"}}' | ./spacetraders-mcp)

# Check if the result contains an error
if echo "$RESULT" | jq -e '.error' > /dev/null 2>&1; then
    echo -e "${RED}Ships resource read failed:${NC}"
    echo "$RESULT" | jq .
else
    echo -e "${GREEN}Ships resource read successful:${NC}"
    # Show just the ship count and first ship symbol for brevity
    SHIP_COUNT=$(echo "$RESULT" | jq -r '.result.contents[0].text' | jq -r '.meta.count')
    FIRST_SHIP=$(echo "$RESULT" | jq -r '.result.contents[0].text' | jq -r '.ships[0].symbol // "none"')
    echo "Ship count: $SHIP_COUNT, First ship: $FIRST_SHIP"
fi

# Test 4: Test with invalid resource URI
echo -e "\n${YELLOW}Test 4: Testing invalid resource URI (should fail gracefully)${NC}"
echo '{"jsonrpc": "2.0", "id": 4, "method": "resources/read", "params": {"uri": "spacetraders://invalid/resource"}}' | ./spacetraders-mcp | jq .

echo -e "\n${GREEN}All tests completed!${NC}"

# Clean up
rm -f spacetraders-mcp

echo -e "\n${YELLOW}Tips:${NC}"
echo "- Make sure your .env file contains SPACETRADERS_API_TOKEN"
echo "- The API token should be valid and not expired"
echo "- If you get authentication errors, check your token in .env"
echo "- Agent info is available as a resource at 'spacetraders://agent/info'"
echo "- Ships list is available as a resource at 'spacetraders://ships/list'"
echo "- Use 'jq' for better JSON formatting (install with: brew install jq or apt install jq)"
