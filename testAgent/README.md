# LangGraph MCP Agent Testing

*English | [中文](README.zh.md)*

This directory contains a LangGraph-based agent for comprehensive testing of MCP service functionality and performance.

## Overview

The LangGraph MCP Agent is an automated testing system that can:
- Establish WebSocket connections and complete MCP protocol handshakes
- Automatically discover all available MCP tools
- Systematically test each tool's functionality
- Generate detailed test reports
- Provide intelligent workflow management

## File Structure

```
testAgent/
├── langgraph_mcp_agent.py  # Main LangGraph agent implementation
├── test_runner.py          # Simplified test runner
├── requirements.txt        # Python dependencies
└── README.md              # This documentation
```

## Core Components

### LangGraph Agent (`langgraph_mcp_agent.py`)

This is a state graph-based agent containing the following nodes:

1. **check_mcp_status** - Check MCP service status
   - Verify HTTP health endpoint
   - Establish WebSocket connection
   - Execute MCP initialization handshake

2. **discover_tools** - Discover available tools
   - Call `tools/list` to get tool list
   - Parse tool definitions and parameters

3. **test_calculator** - Test calculator tool
   - Execute mathematical expression calculation tests

4. **test_web_search** - Test web search tool
   - Execute search query tests

5. **test_knowledge_graph** - Test knowledge graph tool
   - Execute knowledge graph query tests

6. **generate_report** - Generate test report
   - Summarize all test results
   - Generate detailed statistics

### Test Runner (`test_runner.py`)

Provides two test modes:
- **Quick Test** (`quick`): Only tests connection and tool discovery
- **Full Test** (default): Executes functional tests for all tools

## Usage

### Environment Setup

Ensure required Python dependencies are installed:

```bash
# Install dependencies
pip install -r requirements.txt

# Or install in conda/venv environment
conda activate your-env
pip install -r requirements.txt
```

### Start MCP Service

Before testing, ensure the MCP service is running:

```bash
# In project root directory
go run cmd/server/main.go
```

The service should run on `localhost:8030`.

### Run Tests

```bash
# Quick connection test
python test_runner.py quick

# Complete functionality test
python test_runner.py
```

## Test Output Examples

### Quick Test Output
```
🔍 Quick MCP Service Test
------------------------------
✅ MCP service is running and accessible
✅ Found 4 tools:
   • calculator: Basic arithmetic operations
   • web_search: Search the web for information
   • document_analyzer: Analyze documents and extract insights
   • knowledge_graph: Build and query knowledge graphs
```

### Complete Test Output
```
🚀 Starting LangGraph MCP Agent Test...

============================================================
📊 MCP SERVICE TEST REPORT
============================================================
MCP Service Status: available
Tools Discovered: 4
Tests Conducted: 3
Successful Tests: 3
Failed Tests: 0

📝 Detailed Results:
✅ calculator: success
✅ web_search: success
✅ knowledge_graph: success

💬 Execution Log:
ℹ️  Starting MCP service tests...
🔄 MCP service is available ✓
ℹ️  Discovered 4 tools: ['calculator', 'web_search', 'document_analyzer', 'knowledge_graph']
```

## Architecture Design

### MCP Protocol Implementation

The agent implements the complete MCP protocol flow:

1. **Initialization Handshake**
   ```json
   {
     "jsonrpc": "2.0",
     "id": "1",
     "method": "initialize",
     "params": {
       "protocolVersion": "2024-11-05",
       "capabilities": {},
       "clientInfo": {
         "name": "langgraph-test-agent",
         "version": "1.0.0"
       }
     }
   }
   ```

2. **Initialization Complete Notification**
   ```json
   {
     "jsonrpc": "2.0",
     "method": "initialized"
   }
   ```

3. **Tool Invocation**
   ```json
   {
     "jsonrpc": "2.0",
     "id": "2",
     "method": "tools/call",
     "params": {
       "name": "calculator",
       "arguments": {
         "expression": "2 + 3 * 4"
       }
     }
   }
   ```

### State Management

The agent uses TypedDict to define state structure:

```python
class AgentState(TypedDict):
    messages: List[Dict[str, Any]]      # Execution log
    mcp_status: MCPServiceStatus        # MCP service status
    available_tools: List[MCPTool]      # Discovered tools
    current_task: Optional[str]         # Current task
    result: Optional[Dict[str, Any]]    # Test results
    error: Optional[str]                # Error information
```

### Error Handling

The agent has comprehensive error handling mechanisms:
- Connection timeout handling
- WebSocket exception capturing
- JSON-RPC error response parsing
- Tool execution failure handling

## Extension Development

### Adding New Tool Tests

1. Add a new test method to the `MCPTestAgent` class:
   ```python
   async def _test_new_tool(self, state: AgentState) -> AgentState:
       state["current_task"] = "Testing new tool"
       # Test logic
       return state
   ```

2. Add nodes and edges in the `_build_graph` method:
   ```python
   workflow.add_node("test_new_tool", self._test_new_tool)
   workflow.add_edge("previous_node", "test_new_tool")
   ```

### Custom Test Cases

You can customize test cases by modifying parameters in each test method:

```python
# Custom calculator test
test_data = {
    "name": "calculator",
    "arguments": {
        "expression": "sqrt(16) + log(10)"  # More complex expression
    }
}
```

## Dependencies

- **httpx**: HTTP client for health checks
- **websockets**: WebSocket client for MCP communication
- **langgraph**: State graph workflow engine
- **langchain**: LangChain core library
- **typing-extensions**: Type extension support

## Troubleshooting

### Common Issues

1. **Connection Failure**
   - Check if MCP service is started
   - Confirm correct port number (default 8030)
   - Check firewall settings

2. **Tool Call Failure**
   - Check MCP service logs
   - Confirm tool parameter format is correct
   - Verify `initialized` notification was sent

3. **Dependency Installation Failure**
   - Use correct Python environment
   - Check network connection
   - Consider using mirror sources for installation

### Debug Mode

You can add more detailed logging output in the code:

```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

## Future Improvements

- [ ] Support concurrent tool testing
- [ ] Add performance benchmark testing
- [ ] Implement automated regression testing
- [ ] Support custom test configuration files
- [ ] Add visualized test reports