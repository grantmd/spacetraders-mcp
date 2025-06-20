# SpaceTraders MCP Server

A Model Context Protocol (MCP) server for interacting with the SpaceTraders API. This server provides tools to interact with your SpaceTraders agent and game data through MCP-compatible clients like Claude Desktop.

## Features

- **get_agent_info**: Retrieves your agent's current information including credits, headquarters, faction, and ship count

## Prerequisites

- Go 1.24.4 or later
- A SpaceTraders API token (get one by registering at [SpaceTraders](https://spacetraders.io/))
- An MCP-compatible client (like Claude Desktop)

## Setup

1. **Clone and build the project:**
   ```bash
   git clone <your-repo-url>
   cd spacetraders-mcp
   go mod tidy
   go build -o spacetraders-mcp .
   ```

2. **Configure your API token:**
   Create a `.env` file in the project root:
   ```
   SPACETRADERS_API_TOKEN="your_api_token_here"
   ```

   You can get your API token by:
   - Registering a new agent at https://spacetraders.io/
   - Or using an existing token if you already have an agent

3. **Test the server:**
   ```bash
   ./test_mcp.sh
   ```

## Usage

### Running the Server

The MCP server communicates via stdin/stdout:

```bash
./spacetraders-mcp
```

### Available Tools

#### `get_agent_info`

Retrieves comprehensive information about your SpaceTraders agent.

**Parameters:** None

**Returns:**
```json
{
  "agent": {
    "accountId": "your_account_id",
    "credits": 175000,
    "headquarters": "X1-FM66-A1",
    "shipCount": 2,
    "startingFaction": "ASTRO",
    "symbol": "YOUR_CALLSIGN"
  }
}
```

### Manual Testing

You can test the server manually using JSON-RPC 2.0 messages:

1. **List available tools:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | ./spacetraders-mcp
   ```

2. **Call get_agent_info:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "get_agent_info", "arguments": {}}}' | ./spacetraders-mcp
   ```

## Integration with Claude Desktop

To use this MCP server with Claude Desktop, add it to your Claude Desktop configuration:

### macOS
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "spacetraders": {
      "command": "/path/to/your/spacetraders-mcp/spacetraders-mcp",
      "args": []
    }
  }
}
```

### Windows
Edit `%APPDATA%\Claude\claude_desktop_config.json` with the same structure, using Windows paths.

## Development

### Project Structure

```
spacetraders-mcp/
├── main.go              # Main server implementation
├── .env                 # API token configuration (not in git)
├── test_mcp.sh         # Test script
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── README.md           # This file
```

### Adding New Tools

To add new SpaceTraders API tools:

1. Add the API endpoint method to `SpaceTradersClient`
2. Define the response struct if needed
3. Create a new `mcp.Tool` definition in `main()`
4. Add the tool handler function
5. Test your new tool

### Error Handling

The server uses proper MCP error handling:
- API errors are returned as tool results with `isError: true`
- Network errors are handled gracefully
- Invalid requests return appropriate error responses

## API Reference

This server uses the SpaceTraders v2 API. For full API documentation, visit:
- [SpaceTraders API Docs](https://spacetraders.stoplight.io/docs/spacetraders)
- [SpaceTraders Getting Started](https://docs.spacetraders.io/)

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

## Troubleshooting

### Common Issues

1. **"SPACETRADERS_API_TOKEN not found"**
   - Make sure your `.env` file exists and contains the token
   - Check that the token is properly quoted in the `.env` file

2. **"API request failed with status 401"**
   - Your API token is invalid or expired
   - Register a new agent or check your existing token

3. **"Failed to make request"**
   - Check your internet connection
   - Verify the SpaceTraders API is accessible

4. **Build errors**
   - Ensure you have Go 1.24.4 or later
   - Run `go mod tidy` to resolve dependencies

### Getting Help

- Join the [SpaceTraders Discord](https://discord.gg/UpEfRRjsCT)
- Check the SpaceTraders documentation
- Open an issue in this repository