#!/bin/bash

# Debug wrapper for SpaceTraders MCP Server
# This script helps monitor MCP communication by logging all stdin/stdout

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔍 SpaceTraders MCP Debug Monitor${NC}"
echo "This will log all MCP communication to help debug client interactions"
echo "Press Ctrl+C to stop"
echo ""

# Create debug log directory
mkdir -p debug-logs
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
STDIN_LOG="debug-logs/mcp_stdin_${TIMESTAMP}.log"
STDOUT_LOG="debug-logs/mcp_stdout_${TIMESTAMP}.log"
STDERR_LOG="debug-logs/mcp_stderr_${TIMESTAMP}.log"

echo -e "${BLUE}📝 Logging to:${NC}"
echo "  STDIN:  $STDIN_LOG"
echo "  STDOUT: $STDOUT_LOG"
echo "  STDERR: $STDERR_LOG"
echo ""

# Function to log and forward stdin
log_stdin() {
    while IFS= read -r line; do
        echo "$(date '+%H:%M:%S') STDIN: $line" >> "$STDIN_LOG"
        echo -e "${YELLOW}→ IN:${NC} $line" >&2
        echo "$line"
    done
}

# Function to log and forward stdout
log_stdout() {
    while IFS= read -r line; do
        echo "$(date '+%H:%M:%S') STDOUT: $line" >> "$STDOUT_LOG"
        echo -e "${GREEN}← OUT:${NC} $line" >&2
        echo "$line"
    done
}

# Function to log and forward stderr
log_stderr() {
    while IFS= read -r line; do
        echo "$(date '+%H:%M:%S') STDERR: $line" >> "$STDERR_LOG"
        echo -e "${RED}⚠ ERR:${NC} $line" >&2
    done
}

# Cleanup function
cleanup() {
    echo ""
    echo -e "${GREEN}🔍 Debug session ended${NC}"
    echo -e "${BLUE}📊 Summary:${NC}"

    if [ -f "$STDIN_LOG" ]; then
        STDIN_COUNT=$(wc -l < "$STDIN_LOG" 2>/dev/null || echo "0")
        echo "  Messages received: $STDIN_COUNT"
    fi

    if [ -f "$STDOUT_LOG" ]; then
        STDOUT_COUNT=$(wc -l < "$STDOUT_LOG" 2>/dev/null || echo "0")
        echo "  Messages sent: $STDOUT_COUNT"
    fi

    if [ -f "$STDERR_LOG" ]; then
        STDERR_COUNT=$(wc -l < "$STDERR_LOG" 2>/dev/null || echo "0")
        echo "  Log messages: $STDERR_COUNT"
    fi

    echo ""
    echo -e "${BLUE}🔍 Common MCP calls to look for:${NC}"
    echo "  - initialize: Client connecting"
    echo "  - resources/list: Client discovering resources"
    echo "  - tools/list: Client discovering tools"
    echo "  - resources/read: Client reading a resource"
    echo "  - tools/call: Client calling a tool"
    echo ""

    if [ -f "$STDIN_LOG" ]; then
        echo -e "${BLUE}📋 Received calls:${NC}"
        grep -o '"method":"[^"]*"' "$STDIN_LOG" 2>/dev/null | sort | uniq -c | sed 's/^/  /' || echo "  No method calls found"
    fi

    exit 0
}

# Set up signal handling
trap cleanup SIGINT SIGTERM

echo -e "${GREEN}🚀 Starting MCP server with debugging...${NC}"
echo ""

# Build the server first
if ! go build -o spacetraders-mcp-debug; then
    echo -e "${RED}❌ Failed to build server${NC}"
    exit 1
fi

# Start the server with logging
log_stdin | ./spacetraders-mcp-debug 2> >(log_stderr) | log_stdout

# Cleanup on normal exit
cleanup
