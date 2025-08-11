# mcp-go-template

*English | [中文](README.zh.md)*

# Golang MCP (Model Context Protocol) Server Template

A complete MCP server template implemented in Go, providing comprehensive project structure and fundamental functionality.

## Project Structure

```
mcp-go-template/
├── api/                        # API definitions and specifications
│   ├── mcp/
│   │   ├── v1/
│   │   │   └── schema.json     # MCP protocol JSON Schema
│   │   └── openapi.yaml        # OpenAPI specification
│   └── README.md
├── cmd/                        # Application entry points
│   └── server/
│       └── main.go
├── docs/                       # Project documentation
│   ├── architecture.md         # Architecture design
│   ├── deployment.md           # Deployment guide
│   └── examples.md             # Usage examples
├── internal/                   # Private application code
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── server/
│   │   └── server.go          # Main server logic
│   ├── tools/                 # MCP tools implementation
│   │   ├── registry.go        # Tool registry
│   │   └── examples/
│   │       ├── calculator.go  # Calculator tool example
│   │       ├── web_search.go  # Web search tool
│   │       ├── document_analyzer.go # Document analysis tool
│   │       └── knowledge_graph.go   # Knowledge graph tool
│   ├── resources/             # MCP resource management
│   │   ├── registry.go        # Resource registry
│   │   └── examples/
│   │       └── memory.go      # Memory resource example
│   └── prompts/               # MCP prompt management
│       ├── registry.go        # Prompt registry
│       └── examples/
│           └── templates.go   # Prompt templates
├── pkg/                       # Public library code
│   ├── mcp/
│   │   ├── types.go          # MCP protocol type definitions
│   │   ├── handler.go        # MCP handler
│   │   └── validation.go     # Protocol validation
│   └── utils/
│       └── logger.go         # Logging utilities
├── test/                      # Test code
│   ├── integration/           # Integration tests
│   └── testdata/             # Test data
├── testAgent/                 # LangGraph agent testing
│   ├── langgraph_mcp_agent.py # LangGraph MCP test agent
│   ├── test_runner.py         # Test runner
│   └── requirements.txt       # Python dependencies
├── go.mod
├── go.sum
├── README.md
├── Dockerfile
└── docker-compose.yml
```

## Features

- 🚀 Complete MCP protocol implementation
- 🔧 Extensible tool system
- 📦 Resource management support
- 🎯 Prompt template system
- 🔒 Secure middleware support
- 📝 Comprehensive documentation and examples
- 🐳 Docker support
- ✅ Complete test coverage
- 🤖 LangGraph agent integration testing

## Quick Start

```bash
# Clone the project
git clone <your-repo-url>
cd mcp-go-template

# Initialize Go module
go mod init github.com/chongliujia/mcp-go-template

# Install dependencies
go mod tidy

# Run the service
go run cmd/server/main.go

# Or use Docker
docker-compose up
```

## Implementation Status

- ✅ Project structure design
- ✅ Basic MCP protocol implementation
- ✅ Server core functionality
- ✅ Advanced research tool system
- ✅ Configuration management system
- ✅ Docker support

## Advanced Research Tool Suite

This project is designed specifically for advanced research scenarios, providing the following sophisticated tools:

### 🔍 Web Search Tool (web_search)
- Multi-search engine support (DuckDuckGo, Bing, Google)
- Configurable result count and safe search
- Structured search result output

### 📄 Document Analysis Tool (document_analyzer)
- Support for files, URLs, and direct text analysis
- Keyword extraction and frequency analysis
- Document statistics (word count, sentence count, reading time, etc.)
- Automatic summarization
- Entity recognition

### 🕸️ Knowledge Graph Tool (knowledge_graph)
- Build knowledge graphs from text
- Entity extraction (people, organizations, places, concepts, etc.)
- Relationship inference and weight calculation
- Graph visualization and querying

### 🧮 Calculator Tool (calculator)
- Basic mathematical operations
- Floating-point arithmetic support

## Development Guide

### Adding New Tools

1. Create a new tool file under `internal/tools/examples/`
2. Register the new tool in `internal/tools/registry.go`
3. Implement the MCP tool interface

### Adding New Resources

1. Create a new resource file under `internal/resources/examples/`
2. Register the new resource in `internal/resources/registry.go`
3. Implement the MCP resource interface

### Configuration Management

The project uses Viper for configuration management, supporting multiple configuration formats. Configuration files are located at `internal/config/config.go`.

## Testing

### Go Unit Tests

```bash
# Run unit tests
go test ./...

# Run integration tests
go test ./test/integration/...

# Test coverage
go test -cover ./...
```

### LangGraph Agent Testing

Use the LangGraph-built agent to test the complete functionality of the MCP service:

```bash
# Enter the test directory
cd testAgent

# Install Python dependencies (if needed)
pip install -r requirements.txt

# Quick connection test
python test_runner.py quick

# Complete functionality test
python test_runner.py
```

Agent testing features:
- 🔌 WebSocket connection and MCP protocol handshake
- 🛠️ Automatic discovery and testing of all tools
- 📊 Detailed test report generation
- 🤖 LangGraph-based intelligent workflows

## Deployment

See `docs/deployment.md` for details

## Contributing

Issues and Pull Requests are welcome!

## License

MIT License