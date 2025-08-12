# MCP Go Template 架构设计文档

本文档详细描述了 MCP Go Template 项目的架构设计，所有架构组件都与实际代码文件和函数一一对应。

## 系统整体架构

```mermaid
graph TB
    subgraph "Client Layer"
        C1["LangGraph Agent<br/>testAgent/langgraph_mcp_agent.py"]
        C2["JavaScript Client<br/>WebSocket Connection"]
        C3["Python Client<br/>websockets library"]
    end
    
    subgraph "Server Layer"
        S1["HTTP Server<br/>internal/server/server.go"]
        S2["WebSocket Handler<br/>handleWebSocket()"]
        S3["Health Check<br/>handleHealth()"]
        S4["Root Handler<br/>handleRoot()"]
    end
    
    subgraph "MCP Protocol Layer"
        M1["Message Handler<br/>pkg/mcp/handler.go"]
        M2["Protocol Types<br/>pkg/mcp/types.go"]
        M3["Message Validation<br/>pkg/mcp/validation.go"]
    end
    
    subgraph "Business Logic Layer"
        T1["Calculator Tool<br/>internal/tools/examples/calculator.go"]
        T2["Web Search Tool<br/>internal/tools/examples/web_search.go"]
        T3["Document Analyzer<br/>internal/tools/examples/document_analyzer.go"]
        T4["Knowledge Graph Tool<br/>internal/tools/examples/knowledge_graph.go"]
    end
    
    subgraph "Configuration Layer"
        CF1["Config Management<br/>internal/config/config.go"]
        CF2["Logger Utilities<br/>pkg/utils/logger.go"]
    end
    
    subgraph "Application Entry"
        APP["Main Application<br/>cmd/server/main.go"]
    end
    
    C1 -->|WebSocket JSON-RPC| S2
    C2 -->|WebSocket JSON-RPC| S2
    C3 -->|WebSocket JSON-RPC| S2
    
    S1 --> S2
    S1 --> S3
    S1 --> S4
    
    S2 --> M1
    M1 --> M2
    M1 --> M3
    
    M1 --> T1
    M1 --> T2
    M1 --> T3
    M1 --> T4
    
    APP --> CF1
    APP --> S1
    APP --> M1
    
    CF1 --> CF2
    
    style C1 fill:#e1f5fe
    style APP fill:#f3e5f5
    style M1 fill:#e8f5e8
    style T1 fill:#fff3e0
    style T2 fill:#fff3e0
    style T3 fill:#fff3e0
    style T4 fill:#fff3e0
```

## 1. 应用程序入口点

### 主程序 (`cmd/server/main.go`)

**核心函数：**
- `main()` - 程序入口点，负责整体初始化和生命周期管理
- `createServerCapabilities()` - 根据配置创建服务器能力
- `registerTools()` - 注册所有工具实例

**主要职责：**
1. 解析命令行参数
2. 加载配置文件 (`config.Load()`)
3. 设置日志系统 (`utils.SetLogLevel()`)
4. 创建MCP处理器 (`mcp.NewBaseHandler()`)
5. 注册工具集合 (`registerTools()`)
6. 启动服务器 (`srv.Start()`)
7. 优雅关闭处理

## 2. 服务器层架构

### HTTP 服务器 (`internal/server/server.go`)

```mermaid
graph LR
    subgraph "Server Struct"
        SS["Server<br/>internal/server/server.go"]
        SS --> Config["config *config.Config"]
        SS --> Handler["handler mcp.Handler"]
        SS --> Upgrader["upgrader websocket.Upgrader"]
        SS --> Logger["logger *logrus.Logger"]
    end
    
    subgraph "HTTP Endpoints"
        EP1["/mcp - WebSocket<br/>handleWebSocket()"]
        EP2["/health - Health Check<br/>handleHealth()"]
        EP3["/ - Root Info<br/>handleRoot()"]
    end
    
    subgraph "Connection Handling"
        WS["WebSocket Connection<br/>handleConnection()"]
        MSG["Message Processing<br/>sendMessage()"]
        IP["IP Filtering<br/>getClientIP()"]
    end
    
    SS --> EP1
    SS --> EP2
    SS --> EP3
    EP1 --> WS
    WS --> MSG
    EP1 --> IP
```

**关键函数说明：**

- `New()` - 创建服务器实例，初始化WebSocket升级器
- `Start()` - 启动HTTP服务器，支持TLS和优雅关闭
- `handleWebSocket()` - 处理WebSocket连接升级和IP过滤
- `handleConnection()` - 处理单个WebSocket连接的消息循环
- `sendMessage()` - 发送JSON-RPC消息到客户端
- `handleHealth()` - 健康检查端点，返回服务状态
- `getClientIP()` - 提取客户端真实IP地址

## 3. MCP协议层架构

### 消息处理流程

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    participant H as Handler
    participant T as Tool
    
    Note over C,T: MCP Protocol Initialization Flow
    C->>+S: WebSocket Connection
    C->>+S: initialize request<br/>pkg/mcp/handler.go
    S->>+H: HandleMessage()<br/>pkg/mcp/handler.go
    H->>H: handleRequest()<br/>pkg/mcp/handler.go
    H->>H: Initialize()<br/>pkg/mcp/handler.go
    H-->>-S: InitializeResult
    S-->>-C: initialize response
    C->>S: initialized notification<br/>pkg/mcp/handler.go
    
    Note over C,T: Tool Discovery and Execution
    C->>+S: tools/list request
    S->>+H: HandleMessage()
    H->>H: ListTools()<br/>pkg/mcp/handler.go
    H-->>-S: tools list
    S-->>-C: tools response
    
    C->>+S: tools/call request
    S->>+H: HandleMessage()
    H->>H: CallTool()<br/>pkg/mcp/handler.go
    H->>+T: Execute()<br/>internal/tools/examples/*.go
    T-->>-H: CallToolResult
    H-->>-S: tool result
    S-->>-C: call response
```

### MCP处理器 (`pkg/mcp/handler.go`)

**BaseHandler结构：**
```go
type BaseHandler struct {
    serverInfo   ServerInfo
    capabilities ServerCapabilities
    tools        map[string]ToolHandler
    resources    map[string]ResourceHandler
    prompts      map[string]PromptHandler
    initialized  bool
}
```

**关键函数：**
- `NewBaseHandler()` - 创建基础处理器实例
- `HandleMessage()` - 处理所有MCP消息的入口点
- `handleRequest()` - 处理请求消息的分发器
- `handleNotification()` - 处理通知消息
- `Initialize()` - 处理初始化请求
- `CallTool()` - 执行工具调用
- `RegisterTool()` - 注册工具实例

## 4. 工具系统架构

### 工具接口设计

```mermaid
classDiagram
    class ToolHandler {
        <<interface>>
        +Definition() *Tool
        +Execute(ctx Context, params map) *CallToolResult
    }
    
    class CalculatorTool {
        -definition *Tool
        +NewCalculatorTool() *CalculatorTool
        +Definition() *Tool
        +Execute() *CallToolResult
    }
    
    class WebSearchTool {
        -definition *Tool
        +NewWebSearchTool() *WebSearchTool
        +Definition() *Tool
        +Execute() *CallToolResult
    }
    
    class DocumentAnalyzerTool {
        -definition *Tool
        +NewDocumentAnalyzerTool() *DocumentAnalyzerTool
        +Definition() *Tool
        +Execute() *CallToolResult
    }
    
    class KnowledgeGraphTool {
        -definition *Tool
        +NewKnowledgeGraphTool() *KnowledgeGraphTool
        +Definition() *Tool
        +Execute() *CallToolResult
    }
    
    ToolHandler <|-- CalculatorTool
    ToolHandler <|-- WebSearchTool
    ToolHandler <|-- DocumentAnalyzerTool
    ToolHandler <|-- KnowledgeGraphTool
    
    note for CalculatorTool "internal/tools/examples/calculator.go"
    note for WebSearchTool "internal/tools/examples/web_search.go"
    note for DocumentAnalyzerTool "internal/tools/examples/document_analyzer.go"
    note for KnowledgeGraphTool "internal/tools/examples/knowledge_graph.go"
```

### 具体工具实现

#### 1. 计算器工具 (`internal/tools/examples/calculator.go`)

**结构定义：**
```go
type CalculatorTool struct {
    definition *mcp.Tool
}
```

**关键函数：**
- `NewCalculatorTool()` - 构造函数，定义工具规范
- `Definition()` - 返回工具定义
- `Execute()` - 执行数学运算

**支持操作：** add, subtract, multiply, divide, power

#### 2. 网络搜索工具 (`internal/tools/examples/web_search.go`)

**主要功能：**
- 多搜索引擎支持 (DuckDuckGo, Bing, Google)
- 搜索结果解析和结构化输出
- 安全搜索配置

#### 3. 文档分析工具 (`internal/tools/examples/document_analyzer.go`)

**主要功能：**
- 支持文件、URL、文本分析
- 关键词提取和频率分析
- 文档统计和摘要生成

#### 4. 知识图谱工具 (`internal/tools/examples/knowledge_graph.go`)

**主要功能：**
- 实体提取和关系构建
- 图谱可视化支持
- 智能查询处理

## 5. 测试架构 - LangGraph 集成

### 测试智能体架构

```mermaid
graph TB
    subgraph "LangGraph Agent Architecture"
        A1["MCPTestAgent<br/>testAgent/langgraph_mcp_agent.py:42"]
        
        subgraph "State Graph Nodes"
            N1["check_mcp_status<br/>_check_mcp_status()"]
            N2["discover_tools<br/>_discover_tools()"]
            N3["test_calculator<br/>_test_calculator()"]
            N4["test_web_search<br/>_test_web_search()"]
            N5["test_knowledge_graph<br/>_test_knowledge_graph()"]
            N6["generate_report<br/>_generate_report()"]
        end
        
        subgraph "State Management"
            S1["AgentState<br/>TypedDict"]
            S2["MCPServiceStatus<br/>Enum"]
            S3["MCPTool<br/>Dataclass"]
        end
        
        subgraph "Communication Layer"
            WS["WebSocket Client<br/>websockets library"]
            HTTP["HTTP Client<br/>httpx library"]
        end
    end
    
    A1 --> N1
    N1 --> N2
    N2 --> N3
    N3 --> N4
    N4 --> N5
    N5 --> N6
    
    N1 --> HTTP
    N2 --> WS
    N3 --> WS
    N4 --> WS
    N5 --> WS
    
    A1 --> S1
    S1 --> S2
    S1 --> S3
    
    style A1 fill:#e1f5fe
    style N1 fill:#e8f5e8
    style N2 fill:#e8f5e8
    style N3 fill:#fff3e0
    style N4 fill:#fff3e0
    style N5 fill:#fff3e0
    style N6 fill:#f3e5f5
```

**关键组件说明：**

#### MCPTestAgent类 (`testAgent/langgraph_mcp_agent.py`)

**核心属性：**
- `mcp_server_url` - MCP服务器地址
- `websocket` - WebSocket连接实例
- `http_client` - HTTP客户端实例
- `graph` - LangGraph状态图

**关键方法：**
- `_send_mcp_request()` - 发送JSON-RPC请求
- `_build_graph()` - 构建状态图工作流
- `run_tests()` - 执行完整测试流程

#### 测试运行器 (`testAgent/test_runner.py`)

**主要函数：**
- `quick_test()` - 快速连接和工具发现测试
- `main()` - 完整功能测试流程

## 6. 配置管理架构

### 配置系统 (`internal/config/config.go`)

```mermaid
graph LR
    subgraph "Configuration Structure"
        C1[Config Struct<br/>internal/config/config.go]
        
        subgraph "Server Config"
            SC1[Host]
            SC2[Port]
            SC3[Timeout]
        end
        
        subgraph "MCP Config"
            MC1[Name]
            MC2[Version]
            MC3[Capabilities]
        end
        
        subgraph "Security Config"
            SEC1[EnableTLS]
            SEC2[AllowedIPs]
        end
        
        subgraph "Logging Config"
            LC1[Level]
            LC2[Format]
        end
    end
    
    C1 --> SC1
    C1 --> SC2
    C1 --> SC3
    C1 --> MC1
    C1 --> MC2
    C1 --> MC3
    C1 --> SEC1
    C1 --> SEC2
    C1 --> LC1
    C1 --> LC2
```

**关键函数：**
- `Load()` - 加载YAML配置文件
- `GetAddress()` - 获取服务器监听地址
- `IsToolsEnabled()` - 检查工具功能是否启用
- `IsLoggingEnabled()` - 检查日志功能是否启用

## 7. 数据流向分析

### 完整请求处理流程

```mermaid
flowchart TD
    A["Client Request"] --> B["WebSocket Connection<br/>server.go:handleWebSocket()"]
    B --> C["JSON Message Parse<br/>server.go"]
    C --> D["MCP Message Validation<br/>mcp/types.go"]
    D --> E["Handler Dispatch<br/>handler.go:HandleMessage()"]
    
    E --> F{"Message Type"}
    F -->|Request| G["handleRequest()<br/>handler.go"]
    F -->|Notification| H["handleNotification()<br/>handler.go"]
    
    G --> I{"Method Type"}
    I -->|initialize| J["Initialize()<br/>handler.go"]
    I -->|tools/list| K["ListTools()<br/>handler.go"]
    I -->|tools/call| L["CallTool()<br/>handler.go"]
    
    L --> M["Tool Execute()<br/>examples/*.go"]
    M --> N["Result Processing"]
    N --> O["JSON Response<br/>server.go:sendMessage()"]
    O --> P["WebSocket Send"]
    
    J --> N
    K --> N
    H --> Q["Notification Processing"]
    Q --> R["State Update"]
    
    style A fill:#e1f5fe
    style M fill:#fff3e0
    style P fill:#e8f5e8
```

## 8. 错误处理架构

### 错误处理层次

```mermaid
graph TB
    subgraph "Error Handling Layers"
        E1[Application Level<br/>cmd/server/main.go]
        E2[Server Level<br/>internal/server/server.go]
        E3[Protocol Level<br/>pkg/mcp/handler.go]
        E4[Tool Level<br/>internal/tools/examples/*.go]
    end
    
    E1 --> E2
    E2 --> E3
    E3 --> E4
    
    subgraph "Error Codes"
        EC1[ParseError: -32700]
        EC2[InvalidRequest: -32600]
        EC3[MethodNotFound: -32601]
        EC4[InvalidParams: -32602]
        EC5[InternalError: -32603]
    end
    
    E3 --> EC1
    E3 --> EC2
    E3 --> EC3
    E3 --> EC4
    E3 --> EC5
```


## 总结

此架构设计确保了：

1. **模块化**: 每个组件职责明确，代码位置清晰
2. **可扩展性**: 工具系统支持轻松添加新功能
3. **可测试性**: LangGraph智能体提供完整的自动化测试
4. **可维护性**: 配置管理和错误处理层次分明
5. **标准兼容**: 完整实现MCP协议规范

所有架构组件都与实际代码文件一一对应，便于开发者理解和维护。