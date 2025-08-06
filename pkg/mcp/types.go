package mcp

import (
	"encoding/json"
	"fmt"
)

// MCP Protocol Version
const MCPVersion = "2024-11-05"

// RequestID represents a unique identifier for MCP requests
type RequestID interface{}

// Message represents the base MCP message structure
type Message struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      RequestID   `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents MCP error information
type ErrorInfo struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes as defined in MCP specification
const (
	// Standard JSON-RPC errors
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// MCP-specific errors
	InvalidMCPVersion = -32000
	UnknownCapability = -32001
	ResourceNotFound  = -32002
	ToolNotFound      = -32003
	PromptNotFound    = -32004
)

// InitializeParams represents the parameters for the initialize request
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// InitializeResult represents the result of the initialize request
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// ClientCapabilities represents what the client can do
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     map[string]interface{} `json:"sampling,omitempty"`
}

// ServerCapabilities represents what the server can do
type ServerCapabilities struct {
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// LoggingCapability represents logging capabilities
type LoggingCapability struct{}

// PromptsCapability represents prompt capabilities
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resource capabilities
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability represents tool capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ClientInfo represents information about the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema ToolSchema  `json:"inputSchema"`
}

// ToolSchema represents the JSON schema for tool input
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// CallToolParams represents parameters for calling a tool
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents different types of content in MCP
type Content struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     string      `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
	Blob     interface{} `json:"blob,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ReadResourceParams represents parameters for reading a resource
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ReadResourceResult represents the result of reading a resource
type ReadResourceResult struct {
	Contents []ResourceContents `json:"contents"`
}

// ResourceContents represents the contents of a resource
type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents an argument for a prompt template
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// GetPromptParams represents parameters for getting a prompt
type GetPromptParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Description string         `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// LoggingLevel represents different logging levels
type LoggingLevel string

const (
	LoggingLevelDebug   LoggingLevel = "debug"
	LoggingLevelInfo    LoggingLevel = "info"
	LoggingLevelNotice  LoggingLevel = "notice"
	LoggingLevelWarning LoggingLevel = "warning"
	LoggingLevelError   LoggingLevel = "error"
	LoggingLevelAlert   LoggingLevel = "alert"
	LoggingLevelCritical LoggingLevel = "critical"
	LoggingLevelEmergency LoggingLevel = "emergency"
)

// LoggingMessage represents a logging message
type LoggingMessage struct {
	Level  LoggingLevel `json:"level"`
	Data   interface{}  `json:"data"`
	Logger string       `json:"logger,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id RequestID, code int, message string, data interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(id RequestID, result interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewRequest creates a new request message
func NewRequest(id RequestID, method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewNotification creates a new notification message
func NewNotification(method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// IsRequest checks if the message is a request
func (m *Message) IsRequest() bool {
	return m.Method != "" && m.ID != nil
}

// IsNotification checks if the message is a notification
func (m *Message) IsNotification() bool {
	return m.Method != "" && m.ID == nil
}

// IsResponse checks if the message is a response
func (m *Message) IsResponse() bool {
	return m.Method == "" && m.ID != nil
}

// HasError checks if the message contains an error
func (m *Message) HasError() bool {
	return m.Error != nil
}

// Error implements the error interface for ErrorInfo
func (e *ErrorInfo) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("MCP Error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("MCP Error %d: %s", e.Code, e.Message)
}

// UnmarshalParams unmarshals the params field into the provided structure
func (m *Message) UnmarshalParams(v interface{}) error {
	if m.Params == nil {
		return fmt.Errorf("no params to unmarshal")
	}
	
	data, err := json.Marshal(m.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	
	return json.Unmarshal(data, v)
}

// UnmarshalResult unmarshals the result field into the provided structure
func (m *Message) UnmarshalResult(v interface{}) error {
	if m.Result == nil {
		return fmt.Errorf("no result to unmarshal")
	}
	
	data, err := json.Marshal(m.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return json.Unmarshal(data, v)
}