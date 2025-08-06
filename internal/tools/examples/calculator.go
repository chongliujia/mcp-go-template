package examples

import (
	"context"
	"fmt"

	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// CalculatorTool provides basic mathematical operations
type CalculatorTool struct{}

// NewCalculatorTool creates a new calculator tool instance
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

// Definition returns the tool definition for MCP
func (c *CalculatorTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "calculator",
		Description: "Perform basic mathematical operations (add, subtract, multiply, divide)",
		InputSchema: mcp.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
					"description": "The mathematical operation to perform",
				},
				"a": map[string]interface{}{
					"type":        "number",
					"description": "The first number",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "The second number",
				},
			},
			Required: []string{"operation", "a", "b"},
		},
	}
}

// Execute performs the mathematical operation
func (c *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract parameters
	operation, ok := params["operation"].(string)
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: operation parameter must be a string",
			}},
			IsError: true,
		}, nil
	}

	aValue, ok := params["a"]
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: parameter 'a' is required",
			}},
			IsError: true,
		}, nil
	}

	bValue, ok := params["b"]
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: parameter 'b' is required",
			}},
			IsError: true,
		}, nil
	}

	// Convert to float64
	a, err := convertToFloat64(aValue)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: parameter 'a' must be a number: %v", err),
			}},
			IsError: true,
		}, nil
	}

	b, err := convertToFloat64(bValue)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: parameter 'b' must be a number: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Perform the operation
	var result float64
	var operationSymbol string

	switch operation {
	case "add":
		result = a + b
		operationSymbol = "+"
	case "subtract":
		result = a - b
		operationSymbol = "-"
	case "multiply":
		result = a * b
		operationSymbol = "*"
	case "divide":
		if b == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: division by zero is not allowed",
				}},
				IsError: true,
			}, nil
		}
		result = a / b
		operationSymbol = "/"
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported operation '%s'. Supported operations: add, subtract, multiply, divide", operation),
			}},
			IsError: true,
		}, nil
	}

	// Format the result
	resultText := fmt.Sprintf("%.2f %s %.2f = %.6f", a, operationSymbol, b, result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: resultText,
		}},
		IsError: false,
	}, nil
}

// convertToFloat64 converts various numeric types to float64
func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}