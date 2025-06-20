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

# Test 1: List available tools
echo -e "\n${YELLOW}Test 1: Listing available tools${NC}"
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | ./spacetraders-mcp | jq .

# Test 2: Call get_agent_info tool
echo -e "\n${YELLOW}Test 2: Getting agent info${NC}"
RESULT=$(echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "get_agent_info", "arguments": {}}}' | ./spacetraders-mcp)

# Check if the result contains an error
if echo "$RESULT" | jq -e '.result.isError' > /dev/null 2>&1; then
    echo -e "${RED}Tool call failed:${NC}"
    echo "$RESULT" | jq .
else
    echo -e "${GREEN}Tool call successful:${NC}"
    echo "$RESULT" | jq .
fi

# Test 3: Test with invalid tool name
echo -e "\n${YELLOW}Test 3: Testing invalid tool name (should fail gracefully)${NC}"
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "invalid_tool", "arguments": {}}}' | ./spacetraders-mcp | jq .

echo -e "\n${GREEN}All tests completed!${NC}"

# Clean up
rm -f spacetraders-mcp

echo -e "\n${YELLOW}Tips:${NC}"
echo "- Make sure your .env file contains SPACETRADERS_API_TOKEN"
echo "- The API token should be valid and not expired"
echo "- If you get authentication errors, check your token in .env"
echo "- Use 'jq' for better JSON formatting (install with: brew install jq or apt install jq)"
