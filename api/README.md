# MCP Go Template API Documentation

这个目录包含了MCP Go Template服务的API规范和文档。

## 文件说明

### `mcp/v1/schema.json`
包含完整的MCP协议JSON Schema定义，用于：
- 验证消息格式
- 生成客户端代码
- API文档生成
- 开发工具支持

### `openapi.yaml`
OpenAPI 3.0规范文档，包含：
- HTTP端点定义（健康检查等）
- WebSocket端点说明
- 消息格式示例
- 错误代码说明

## MCP协议概述

Model Context Protocol (MCP) 是一个用于AI代理与外部系统通信的标准协议。

### 核心概念

1. **Tools（工具）**: 可执行的功能，如计算器、搜索等
2. **Resources（资源）**: 可读取的数据源，如文件、数据库等  
3. **Prompts（提示）**: 可重用的提示模板

### 通信方式

- 基于WebSocket的JSON-RPC 2.0协议
- 支持请求/响应和通知消息
- 异步消息处理

### 消息类型

#### 请求消息
```json
{
  "jsonrpc": "2.0",
  "id": "unique-id",
  "method": "method-name",
  "params": { ... }
}
```

#### 响应消息
```json
{
  "jsonrpc": "2.0", 
  "id": "unique-id",
  "result": { ... }
}
```

#### 错误响应
```json
{
  "jsonrpc": "2.0",
  "id": "unique-id", 
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "additional error info"
  }
}
```

## 使用示例

### 连接到服务器

```javascript
const ws = new WebSocket('ws://localhost:8030/mcp');

ws.onopen = function() {
  // 发送初始化请求
  ws.send(JSON.stringify({
    jsonrpc: "2.0",
    id: "1",
    method: "initialize", 
    params: {
      protocolVersion: "2024-11-05",
      capabilities: {},
      clientInfo: {
        name: "my-client",
        version: "1.0.0"
      }
    }
  }));
};
```

### 调用工具

```javascript
ws.send(JSON.stringify({
  jsonrpc: "2.0",
  id: "2",
  method: "tools/call",
  params: {
    name: "calculator",
    arguments: {
      operation: "add",
      a: 10,
      b: 5
    }
  }
}));
```

## 错误代码

| 代码 | 名称 | 描述 |
|------|------|------|
| -32700 | Parse error | JSON解析错误 |
| -32600 | Invalid Request | 无效请求 |
| -32601 | Method not found | 方法未找到 |
| -32602 | Invalid params | 无效参数 |
| -32603 | Internal error | 内部错误 |
| -32000 | Invalid MCP version | 无效MCP版本 |
| -32001 | Unknown capability | 未知能力 |
| -32002 | Resource not found | 资源未找到 |
| -32003 | Tool not found | 工具未找到 |
| -32004 | Prompt not found | 提示未找到 |

## 开发工具

### 验证消息格式

使用JSON Schema验证器验证消息：

```bash
# 安装ajv-cli
npm install -g ajv-cli

# 验证消息
echo '{"jsonrpc":"2.0","id":"1","method":"initialize","params":{...}}' | ajv validate -s mcp/v1/schema.json
```

### 生成客户端代码

可以使用OpenAPI生成器生成各种语言的客户端代码：

```bash
# 生成Python客户端
openapi-generator generate -i openapi.yaml -g python -o ./clients/python

# 生成JavaScript客户端  
openapi-generator generate -i openapi.yaml -g javascript -o ./clients/javascript
```