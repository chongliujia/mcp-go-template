package mcp

import (
	"encoding/json"
	"testing"
)

func TestMessage_IsRequest(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name: "valid request",
			message: Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
			},
			expected: true,
		},
		{
			name: "notification (no ID)",
			message: Message{
				JSONRPC: "2.0",
				Method:  "initialized",
			},
			expected: false,
		},
		{
			name: "response (no method)",
			message: Message{
				JSONRPC: "2.0",
				ID:      "1",
				Result:  "success",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.message.IsRequest(); got != tt.expected {
				t.Errorf("IsRequest() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMessage_IsNotification(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name: "valid notification",
			message: Message{
				JSONRPC: "2.0",
				Method:  "initialized",
			},
			expected: true,
		},
		{
			name: "request (has ID)",
			message: Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.message.IsNotification(); got != tt.expected {
				t.Errorf("IsNotification() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewSuccessResponse(t *testing.T) {
	id := "test-id"
	result := map[string]string{"status": "ok"}

	msg := NewSuccessResponse(id, result)

	if msg.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", msg.JSONRPC)
	}
	if msg.ID != id {
		t.Errorf("Expected ID to be %v, got %v", id, msg.ID)
	}
	if msg.Result == nil {
		t.Error("Expected Result to be set")
	}
	if msg.Error != nil {
		t.Error("Expected Error to be nil")
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := "test-id"
	code := InvalidParams
	message := "Invalid parameters"
	data := "additional error data"

	msg := NewErrorResponse(id, code, message, data)

	if msg.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", msg.JSONRPC)
	}
	if msg.ID != id {
		t.Errorf("Expected ID to be %v, got %v", id, msg.ID)
	}
	if msg.Error == nil {
		t.Fatal("Expected Error to be set")
	}
	if msg.Error.Code != code {
		t.Errorf("Expected error code to be %d, got %d", code, msg.Error.Code)
	}
	if msg.Error.Message != message {
		t.Errorf("Expected error message to be %s, got %s", message, msg.Error.Message)
	}
	if msg.Error.Data != data {
		t.Errorf("Expected error data to be %v, got %v", data, msg.Error.Data)
	}
}

func TestMessage_UnmarshalParams(t *testing.T) {
	// Create a message with params
	params := InitializeParams{
		ProtocolVersion: MCPVersion,
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	msg := &Message{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "initialize",
		Params:  params,
	}

	var result InitializeParams
	err := msg.UnmarshalParams(&result)
	if err != nil {
		t.Fatalf("UnmarshalParams failed: %v", err)
	}

	if result.ProtocolVersion != params.ProtocolVersion {
		t.Errorf("Expected protocol version %s, got %s", params.ProtocolVersion, result.ProtocolVersion)
	}
	if result.ClientInfo.Name != params.ClientInfo.Name {
		t.Errorf("Expected client name %s, got %s", params.ClientInfo.Name, result.ClientInfo.Name)
	}
}

func TestErrorInfo_Error(t *testing.T) {
	err := &ErrorInfo{
		Code:    InvalidParams,
		Message: "Test error",
		Data:    "additional data",
	}

	expected := "MCP Error -32602: Test error (data: additional data)"
	if err.Error() != expected {
		t.Errorf("Expected error string %s, got %s", expected, err.Error())
	}

	// Test without data
	errNoData := &ErrorInfo{
		Code:    InvalidParams,
		Message: "Test error",
	}

	expectedNoData := "MCP Error -32602: Test error"
	if errNoData.Error() != expectedNoData {
		t.Errorf("Expected error string %s, got %s", expectedNoData, errNoData.Error())
	}
}

func TestJSONSerialization(t *testing.T) {
	// Test serialization of a complete message
	msg := NewSuccessResponse("test-123", map[string]interface{}{
		"protocolVersion": MCPVersion,
		"capabilities": ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		"serverInfo": ServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
	})

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if unmarshaled.JSONRPC != msg.JSONRPC {
		t.Errorf("JSONRPC mismatch after serialization")
	}
	if unmarshaled.ID != msg.ID {
		t.Errorf("ID mismatch after serialization")
	}
}