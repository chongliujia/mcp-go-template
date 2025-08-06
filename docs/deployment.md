# MCP Go Template 服务部署指南

本文档详细介绍了如何将MCP Go Template服务部署到不同环境，以便外部应用（如LangGraph、Claude Desktop等）可以访问。

## 目录

- [本地开发部署](#本地开发部署)
- [使用ngrok快速公开服务](#使用ngrok快速公开服务)
- [Docker容器部署](#docker容器部署)
- [云服务器部署](#云服务器部署)
- [Kubernetes部署](#kubernetes部署)
- [反向代理配置](#反向代理配置)
- [安全配置](#安全配置)
- [监控和日志](#监控和日志)
- [客户端连接示例](#客户端连接示例)

## 本地开发部署

### 快速启动

```bash
# 克隆项目
git clone https://github.com/chongliujia/mcp-go-template.git
cd mcp-go-template

# 复制配置文件
cp config.example.yaml config.yaml

# 启动开发服务
./scripts/start.sh development
```

### 手动启动

```bash
# 构建项目
go build -o server cmd/server/main.go

# 启动服务
./server -config=config.yaml -log-level=debug
```

### 访问地址

- **WebSocket端点**: `ws://localhost:8030/mcp`
- **健康检查**: `http://localhost:8030/health`
- **根路径**: `http://localhost:8030/`

### 配置文件

```yaml
# config.yaml
server:
  host: "localhost"
  port: 8030
  timeout: 30

logging:
  level: "debug"
  format: "text"

mcp:
  name: "mcp-go-template"
  version: "1.0.0"
  capabilities:
    tools:
      enabled: true
    resources:
      enabled: true
    prompts:
      enabled: true
    logging:
      enabled: true
```

## 使用ngrok快速公开服务

ngrok是一个快速将本地服务暴露到公网的工具，适合开发测试。

### 安装ngrok

```bash
# macOS
brew install ngrok

# 或下载二进制文件
# https://ngrok.com/download
```

### 注册并获取认证令牌

1. 访问 [ngrok.com](https://ngrok.com) 注册账号
2. 获取认证令牌
3. 配置认证令牌：

```bash
ngrok config add-authtoken YOUR_AUTH_TOKEN
```

### 启动服务

```bash
# 终端1：启动MCP服务
./scripts/start.sh development

# 终端2：启动ngrok
ngrok http 8030
```

### 获取公网地址

ngrok启动后会显示类似信息：