# LangGraph MCP 智能体测试

*[English](README.md) | 中文*

这个目录包含了基于 LangGraph 构建的智能体，用于全面测试 MCP 服务的功能和性能。

## 概述

LangGraph MCP 智能体是一个自动化测试系统，能够：
- 建立 WebSocket 连接并完成 MCP 协议握手
- 自动发现所有可用的 MCP 工具
- 系统性地测试每个工具的功能
- 生成详细的测试报告
- 提供智能化的工作流管理

## 文件结构

```
testAgent/
├── langgraph_mcp_agent.py  # 主要的 LangGraph 智能体实现
├── test_runner.py          # 简化的测试运行器
├── requirements.txt        # Python 依赖包
└── README.md              # 本文档
```

## 核心组件

### LangGraph 智能体 (`langgraph_mcp_agent.py`)

这是一个基于状态图的智能体，包含以下节点：

1. **check_mcp_status** - 检查 MCP 服务状态
   - 验证 HTTP 健康端点
   - 建立 WebSocket 连接
   - 执行 MCP 初始化握手

2. **discover_tools** - 发现可用工具
   - 调用 `tools/list` 获取工具列表
   - 解析工具定义和参数

3. **test_calculator** - 测试计算器工具
   - 执行数学表达式计算测试

4. **test_web_search** - 测试网络搜索工具
   - 执行搜索查询测试

5. **test_knowledge_graph** - 测试知识图谱工具
   - 执行知识图谱查询测试

6. **generate_report** - 生成测试报告
   - 汇总所有测试结果
   - 生成详细的统计信息

### 测试运行器 (`test_runner.py`)

提供两种测试模式：
- **快速测试** (`quick`): 仅测试连接和工具发现
- **完整测试** (默认): 执行所有工具的功能测试

## 使用方法

### 环境准备

确保已安装所需的 Python 依赖：

```bash
# 安装依赖
pip install -r requirements.txt

# 或者在 conda/venv 环境中安装
conda activate your-env
pip install -r requirements.txt
```

### 启动 MCP 服务

在测试前，确保 MCP 服务正在运行：

```bash
# 在项目根目录
go run cmd/server/main.go
```

服务应该在 `localhost:8030` 运行。

### 运行测试

```bash
# 快速连接测试
python test_runner.py quick

# 完整功能测试
python test_runner.py
```

## 测试输出示例

### 快速测试输出
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

### 完整测试输出
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

## 架构设计

### MCP 协议实现

智能体实现了完整的 MCP 协议流程：

1. **初始化握手**
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

2. **初始化完成通知**
   ```json
   {
     "jsonrpc": "2.0",
     "method": "initialized"
   }
   ```

3. **工具调用**
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

### 状态管理

智能体使用 TypedDict 定义状态结构：

```python
class AgentState(TypedDict):
    messages: List[Dict[str, Any]]      # 执行日志
    mcp_status: MCPServiceStatus        # MCP 服务状态
    available_tools: List[MCPTool]      # 发现的工具
    current_task: Optional[str]         # 当前任务
    result: Optional[Dict[str, Any]]    # 测试结果
    error: Optional[str]                # 错误信息
```

### 错误处理

智能体具有完善的错误处理机制：
- 连接超时处理
- WebSocket 异常捕获
- JSON-RPC 错误响应解析
- 工具执行失败处理

## 扩展开发

### 添加新的工具测试

1. 在 `MCPTestAgent` 类中添加新的测试方法：
   ```python
   async def _test_new_tool(self, state: AgentState) -> AgentState:
       state["current_task"] = "Testing new tool"
       # 测试逻辑
       return state
   ```

2. 在 `_build_graph` 方法中添加节点和边：
   ```python
   workflow.add_node("test_new_tool", self._test_new_tool)
   workflow.add_edge("previous_node", "test_new_tool")
   ```

### 自定义测试用例

可以通过修改各个测试方法中的参数来自定义测试用例：

```python
# 自定义计算器测试
test_data = {
    "name": "calculator",
    "arguments": {
        "expression": "sqrt(16) + log(10)"  # 更复杂的表达式
    }
}
```

## 依赖说明

- **httpx**: HTTP 客户端，用于健康检查
- **websockets**: WebSocket 客户端，用于 MCP 通信
- **langgraph**: 状态图工作流引擎
- **langchain**: LangChain 核心库
- **typing-extensions**: 类型扩展支持

## 故障排除

### 常见问题

1. **连接失败**
   - 检查 MCP 服务是否启动
   - 确认端口号是否正确（默认 8030）
   - 查看防火墙设置

2. **工具调用失败**
   - 检查 MCP 服务日志
   - 确认工具参数格式是否正确
   - 验证 `initialized` 通知是否发送

3. **依赖安装失败**
   - 使用正确的 Python 环境
   - 检查网络连接
   - 考虑使用镜像源安装

### 调试模式

可以在代码中添加更详细的日志输出：

```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

## 未来改进

- [ ] 支持并发工具测试
- [ ] 添加性能基准测试
- [ ] 实现自动化回归测试
- [ ] 支持自定义测试配置文件
- [ ] 添加可视化测试报告