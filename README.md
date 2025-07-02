# SpaceTraders MCP Server

A Model Context Protocol (MCP) server that provides seamless integration between Claude Desktop and the SpaceTraders API, enabling intelligent assistance for your space trading adventures.

## ✨ Features

- **🤖 Intelligent Game Assistant**: Get contextual advice and strategic recommendations
- **📊 Real-time Data Access**: Automatic access to your ships, contracts, and systems
- **🛠️ Automated Actions**: Execute complex operations through natural language
- **🗺️ Smart Exploration**: Intelligent system mapping and facility discovery
- **📈 Strategic Analysis**: Contract optimization and fleet management recommendations

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Claude Desktop
- SpaceTraders API token ([Get one here](https://spacetraders.io))

### Installation

1. **Clone and build:**
   ```bash
   git clone https://github.com/your-username/spacetraders-mcp.git
   cd spacetraders-mcp
   go build -o spacetraders-mcp
   ```

2. **Set up your SpaceTraders credentials:**
   ```bash
   export SPACETRADERS_TOKEN=your_token_here
   export SPACETRADERS_AGENT_SYMBOL=your_agent_symbol
   ```

3. **Configure Claude Desktop:**
   
   Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):
   ```json
   {
     "mcpServers": {
       "spacetraders": {
         "command": "/path/to/spacetraders-mcp/spacetraders-mcp",
         "args": []
       }
     }
   }
   ```

4. **Restart Claude Desktop** and start exploring!

### First Steps

Once configured, try these commands in Claude Desktop:

- **"Show me my current SpaceTraders status"** - Get a complete overview
- **"What systems should I explore next?"** - Get strategic recommendations  
- **"Help me optimize my fleet"** - Analyze and improve your ships
- **"Find the best trading opportunities"** - Discover profitable routes

## 📖 Documentation

Detailed documentation is available in the `docs/` folder:

- **[Resources](docs/resources.md)** - Available data resources and how to use them
- **[Tools](docs/tools.md)** - Actions you can perform through Claude
- **[Prompts](docs/prompts.md)** - Intelligent assistance for common tasks
- **[Integration](docs/integration.md)** - Claude Desktop setup and configuration
- **[MCP Resources Guide](docs/mcp-resources.md)** - Understanding how MCP resources work
- **[Development](docs/development.md)** - Contributing and extending the server
- **[Testing](docs/testing.md)** - Running tests and quality assurance
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions
- **[API Reference](docs/api-reference.md)** - Complete API documentation
- **[CI/CD](docs/cicd.md)** - Automated workflows and deployment

## 🎮 What You Can Do

### Game Management
- Check agent status, credits, and fleet overview
- Monitor ship locations, cargo, and fuel levels
- Review and manage contracts
- Track system exploration progress

### Strategic Operations
- Navigate ships between waypoints and systems
- Extract resources and manage cargo
- Find optimal trading routes and opportunities
- Purchase and configure new ships

### Intelligent Assistance
- Get recommendations based on your current situation
- Analyze systems for facilities and resources
- Optimize contract selection and completion strategies
- Plan fleet composition and expansion

## 🔗 Example Workflows

**System Exploration:**
```
"Explore system X1-DF55 and tell me what facilities are available"
```

**Fleet Management:**
```
"Where are all my ships and what should I do with them?"
```

**Trading Strategy:**
```
"Analyze my current cargo and find the best places to sell"
```

**Contract Planning:**
```
"Review my contracts and suggest the most profitable completion order"
```

## 🆘 Getting Help

- **First time setup?** Check the [Integration Guide](docs/integration.md)
- **Having issues?** See the [Troubleshooting Guide](docs/troubleshooting.md)
- **Want to contribute?** Read the [Development Guide](docs/development.md)
- **Found a bug?** [Open an issue](https://github.com/your-username/spacetraders-mcp/issues)

## 📜 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please see our [Development Guide](docs/development.md) for details on how to get started.

---

**Ready to enhance your SpaceTraders experience?** Follow the Quick Start guide above and begin your intelligent space trading journey! 🚀