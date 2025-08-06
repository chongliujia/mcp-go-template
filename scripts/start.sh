#!/bin/bash

# MCP Go Template Startup Script
# Usage: ./scripts/start.sh [development|production]

set -e

MODE=${1:-development}
CONFIG_FILE=""
LOG_LEVEL=""

echo "🚀 Starting MCP Go Template Server in $MODE mode..."

case $MODE in
    development)
        CONFIG_FILE="config.example.yaml"
        LOG_LEVEL="debug"
        echo "📝 Using development configuration"
        ;;
    production)
        CONFIG_FILE="config.yaml"
        LOG_LEVEL="info"
        echo "🔧 Using production configuration"
        if [[ ! -f "$CONFIG_FILE" ]]; then
            echo "❌ Production config file not found: $CONFIG_FILE"
            echo "💡 Copy config.example.yaml to config.yaml and customize it"
            exit 1
        fi
        ;;
    *)
        echo "❌ Invalid mode: $MODE"
        echo "Usage: $0 [development|production]"
        exit 1
        ;;
esac

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Build the application
echo "🔨 Building application..."
go build -o bin/mcp-server ./cmd/server

# Create necessary directories
mkdir -p logs data

# Start the server
echo "🌟 Starting server..."
echo "📊 Server will be available at http://localhost:8030"
echo "🔗 WebSocket endpoint: ws://localhost:8030/mcp"
echo "❤️  Health check: http://localhost:8030/health"
echo ""

if [[ -f "$CONFIG_FILE" ]]; then
    ./bin/mcp-server -config="$CONFIG_FILE" -log-level="$LOG_LEVEL"
else
    ./bin/mcp-server -log-level="$LOG_LEVEL"
fi