# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy config files
COPY config.example.yaml ./config.yaml

# Create directories for data
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8030

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8030/health || exit 1

# Set default environment variables
ENV MCP_SERVER_HOST=0.0.0.0
ENV MCP_SERVER_PORT=8030
ENV MCP_LOGGING_LEVEL=info

# Run the application
CMD ["./main"]