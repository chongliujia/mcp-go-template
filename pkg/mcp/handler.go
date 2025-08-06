package mcp

import (
	"context"
	"fmt"
)

// Handler defines the interface for MCP request handlers
type Handler interface {
	HandleMessage(ctx context.Context, message *Message) (*Message, error)
	Initialize(params *InitializeParams) (*InitializeResult, error)
	ListTools() ([]*Tool, error)
	CallTool(params *CallToolParams) (*CallToolResult, error)
	ListResources() ([]*Resource, error)
	ReadResource(params *ReadResourceParams) (*ReadResourceResult, error)
	ListPrompts() ([]*Prompt, error)
	GetPrompt(params *GetPromptParams) (*GetPromptResult, error)
}

// BaseHandler provides a base implementation of the Handler interface
type BaseHandler struct {
	serverInfo   ServerInfo
	capabilities ServerCapabilities
	tools        map[string]ToolHandler
	resources    map[string]ResourceHandler
	prompts      map[string]PromptHandler
	initialized  bool
}

// ToolHandler defines the interface for tool implementations
type ToolHandler interface {
	Definition() *Tool
	Execute(ctx context.Context, params map[string]interface{}) (*CallToolResult, error)
}

// ResourceHandler defines the interface for resource implementations
type ResourceHandler interface {
	Definition() *Resource
	Read(ctx context.Context, uri string) (*ReadResourceResult, error)
}

// PromptHandler defines the interface for prompt implementations
type PromptHandler interface {
	Definition() *Prompt
	Generate(ctx context.Context, params map[string]interface{}) (*GetPromptResult, error)
}

// NewBaseHandler creates a new BaseHandler with the given server info and capabilities
func NewBaseHandler(serverInfo ServerInfo, capabilities ServerCapabilities) *BaseHandler {
	return &BaseHandler{
		serverInfo:   serverInfo,
		capabilities: capabilities,
		tools:        make(map[string]ToolHandler),
		resources:    make(map[string]ResourceHandler),
		prompts:      make(map[string]PromptHandler),
		initialized:  false,
	}
}

// RegisterTool registers a tool handler
func (h *BaseHandler) RegisterTool(handler ToolHandler) error {
	tool := handler.Definition()
	if tool == nil {
		return fmt.Errorf("tool definition cannot be nil")
	}
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	h.tools[tool.Name] = handler
	return nil
}

// RegisterResource registers a resource handler
func (h *BaseHandler) RegisterResource(handler ResourceHandler) error {
	resource := handler.Definition()
	if resource == nil {
		return fmt.Errorf("resource definition cannot be nil")
	}
	if resource.URI == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}
	h.resources[resource.URI] = handler
	return nil
}

// RegisterPrompt registers a prompt handler
func (h *BaseHandler) RegisterPrompt(handler PromptHandler) error {
	prompt := handler.Definition()
	if prompt == nil {
		return fmt.Errorf("prompt definition cannot be nil")
	}
	if prompt.Name == "" {
		return fmt.Errorf("prompt name cannot be empty")
	}
	h.prompts[prompt.Name] = handler
	return nil
}

// HandleMessage handles an incoming MCP message
func (h *BaseHandler) HandleMessage(ctx context.Context, message *Message) (*Message, error) {
	if message == nil {
		return NewErrorResponse(nil, InvalidRequest, "message cannot be nil", nil), nil
	}

	if message.IsRequest() {
		return h.handleRequest(ctx, message)
	}

	if message.IsNotification() {
		return h.handleNotification(ctx, message)
	}

	return NewErrorResponse(message.ID, InvalidRequest, "invalid message format", nil), nil
}

// handleRequest handles MCP requests
func (h *BaseHandler) handleRequest(ctx context.Context, message *Message) (*Message, error) {
	switch message.Method {
	case "initialize":
		var params InitializeParams
		if err := message.UnmarshalParams(&params); err != nil {
			return NewErrorResponse(message.ID, InvalidParams, "invalid initialize params", err.Error()), nil
		}
		
		result, err := h.Initialize(&params)
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "initialization failed", err.Error()), nil
		}
		
		return NewSuccessResponse(message.ID, result), nil

	case "tools/list":
		tools, err := h.ListTools()
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "failed to list tools", err.Error()), nil
		}
		
		result := map[string]interface{}{
			"tools": tools,
		}
		return NewSuccessResponse(message.ID, result), nil

	case "tools/call":
		var params CallToolParams
		if err := message.UnmarshalParams(&params); err != nil {
			return NewErrorResponse(message.ID, InvalidParams, "invalid tool call params", err.Error()), nil
		}
		
		result, err := h.CallTool(&params)
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "tool call failed", err.Error()), nil
		}
		
		return NewSuccessResponse(message.ID, result), nil

	case "resources/list":
		resources, err := h.ListResources()
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "failed to list resources", err.Error()), nil
		}
		
		result := map[string]interface{}{
			"resources": resources,
		}
		return NewSuccessResponse(message.ID, result), nil

	case "resources/read":
		var params ReadResourceParams
		if err := message.UnmarshalParams(&params); err != nil {
			return NewErrorResponse(message.ID, InvalidParams, "invalid resource read params", err.Error()), nil
		}
		
		result, err := h.ReadResource(&params)
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "resource read failed", err.Error()), nil
		}
		
		return NewSuccessResponse(message.ID, result), nil

	case "prompts/list":
		prompts, err := h.ListPrompts()
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "failed to list prompts", err.Error()), nil
		}
		
		result := map[string]interface{}{
			"prompts": prompts,
		}
		return NewSuccessResponse(message.ID, result), nil

	case "prompts/get":
		var params GetPromptParams
		if err := message.UnmarshalParams(&params); err != nil {
			return NewErrorResponse(message.ID, InvalidParams, "invalid prompt get params", err.Error()), nil
		}
		
		result, err := h.GetPrompt(&params)
		if err != nil {
			return NewErrorResponse(message.ID, InternalError, "prompt get failed", err.Error()), nil
		}
		
		return NewSuccessResponse(message.ID, result), nil

	default:
		return NewErrorResponse(message.ID, MethodNotFound, fmt.Sprintf("method '%s' not found", message.Method), nil), nil
	}
}

// handleNotification handles MCP notifications
func (h *BaseHandler) handleNotification(ctx context.Context, message *Message) (*Message, error) {
	switch message.Method {
	case "initialized":
		// Client has completed initialization
		h.initialized = true
		return nil, nil
		
	case "notifications/cancelled":
		// Handle request cancellation
		return nil, nil
		
	default:
		// Unknown notification, ignore
		return nil, nil
	}
}

// Initialize handles the initialize request
func (h *BaseHandler) Initialize(params *InitializeParams) (*InitializeResult, error) {
	if params.ProtocolVersion != MCPVersion {
		return nil, fmt.Errorf("unsupported protocol version: %s", params.ProtocolVersion)
	}

	result := &InitializeResult{
		ProtocolVersion: MCPVersion,
		Capabilities:    h.capabilities,
		ServerInfo:      h.serverInfo,
	}

	return result, nil
}

// ListTools returns all registered tools
func (h *BaseHandler) ListTools() ([]*Tool, error) {
	tools := make([]*Tool, 0, len(h.tools))
	for _, handler := range h.tools {
		tools = append(tools, handler.Definition())
	}
	return tools, nil
}

// CallTool executes a tool with the given parameters
func (h *BaseHandler) CallTool(params *CallToolParams) (*CallToolResult, error) {
	if !h.initialized {
		return nil, fmt.Errorf("handler not initialized")
	}

	handler, exists := h.tools[params.Name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", params.Name)
	}

	ctx := context.Background()
	result, err := handler.Execute(ctx, params.Arguments)
	if err != nil {
		return &CallToolResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Tool execution failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return result, nil
}

// ListResources returns all registered resources
func (h *BaseHandler) ListResources() ([]*Resource, error) {
	resources := make([]*Resource, 0, len(h.resources))
	for _, handler := range h.resources {
		resources = append(resources, handler.Definition())
	}
	return resources, nil
}

// ReadResource reads a resource with the given URI
func (h *BaseHandler) ReadResource(params *ReadResourceParams) (*ReadResourceResult, error) {
	if !h.initialized {
		return nil, fmt.Errorf("handler not initialized")
	}

	handler, exists := h.resources[params.URI]
	if !exists {
		return nil, fmt.Errorf("resource '%s' not found", params.URI)
	}

	ctx := context.Background()
	return handler.Read(ctx, params.URI)
}

// ListPrompts returns all registered prompts
func (h *BaseHandler) ListPrompts() ([]*Prompt, error) {
	prompts := make([]*Prompt, 0, len(h.prompts))
	for _, handler := range h.prompts {
		prompts = append(prompts, handler.Definition())
	}
	return prompts, nil
}

// GetPrompt generates a prompt with the given parameters
func (h *BaseHandler) GetPrompt(params *GetPromptParams) (*GetPromptResult, error) {
	if !h.initialized {
		return nil, fmt.Errorf("handler not initialized")
	}

	handler, exists := h.prompts[params.Name]
	if !exists {
		return nil, fmt.Errorf("prompt '%s' not found", params.Name)
	}

	ctx := context.Background()
	return handler.Generate(ctx, params.Arguments)
}

// IsInitialized returns whether the handler has been initialized
func (h *BaseHandler) IsInitialized() bool {
	return h.initialized
}

// GetServerInfo returns the server information
func (h *BaseHandler) GetServerInfo() ServerInfo {
	return h.serverInfo
}

// GetCapabilities returns the server capabilities
func (h *BaseHandler) GetCapabilities() ServerCapabilities {
	return h.capabilities
}