package examples

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// CalculatorTool implements a basic mathematical calculator
type CalculatorTool struct {
	definition *mcp.Tool
}

// NewCalculatorTool creates a new calculator tool
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		definition: &mcp.Tool{
			Name:        "calculator",
			Description: "Performs basic mathematical operations including addition, subtraction, multiplication, division, and power calculations",
			InputSchema: mcp.ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The mathematical operation to perform",
						"enum":        []string{"add", "subtract", "multiply", "divide", "power"},
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
		},
	}
}

// Definition returns the tool definition
func (c *CalculatorTool) Definition() *mcp.Tool {
	return c.definition
}

// Execute performs the mathematical calculation
func (c *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract parameters
	operation, ok := params["operation"].(string)
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: operation must be a string",
			}},
			IsError: true,
		}, nil
	}

	// Convert numbers from interface{}
	aVal, err := parseNumber(params["a"])
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: invalid first number: %v", err),
			}},
			IsError: true,
		}, nil
	}

	bVal, err := parseNumber(params["b"])
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: invalid second number: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Perform calculation with enhanced error checking
	var result float64
	var resultText string

	switch operation {
	case "add":
		result = aVal + bVal
		// Check for overflow
		if math.IsInf(result, 0) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: addition resulted in overflow",
				}},
				IsError: true,
			}, nil
		}
		resultText = fmt.Sprintf("%.6g + %.6g = %.6g", aVal, bVal, result)
	case "subtract":
		result = aVal - bVal
		// Check for overflow
		if math.IsInf(result, 0) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: subtraction resulted in overflow",
				}},
				IsError: true,
			}, nil
		}
		resultText = fmt.Sprintf("%.6g - %.6g = %.6g", aVal, bVal, result)
	case "multiply":
		result = aVal * bVal
		// Check for overflow
		if math.IsInf(result, 0) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: multiplication resulted in overflow",
				}},
				IsError: true,
			}, nil
		}
		resultText = fmt.Sprintf("%.6g × %.6g = %.6g", aVal, bVal, result)
	case "divide":
		if bVal == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: division by zero is not allowed",
				}},
				IsError: true,
			}, nil
		}
		result = aVal / bVal
		// Check for result validity
		if math.IsNaN(result) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: division resulted in invalid number (NaN)",
				}},
				IsError: true,
			}, nil
		}
		resultText = fmt.Sprintf("%.6g ÷ %.6g = %.6g", aVal, bVal, result)
	case "power":
		// Enhanced power implementation with better validation
		if bVal != float64(int(bVal)) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: power operation only supports integer exponents",
				}},
				IsError: true,
			}, nil
		}
		
		exp := int(bVal)
		// Check for extremely large exponents
		if math.Abs(float64(exp)) > 1000 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: exponent too large (maximum ±1000 supported)",
				}},
				IsError: true,
			}, nil
		}
		
		result = power(aVal, exp)
		// Check for overflow/underflow
		if math.IsInf(result, 0) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: fmt.Sprintf("Error: %.6g ^ %d resulted in overflow", aVal, exp),
				}},
				IsError: true,
			}, nil
		}
		if math.IsNaN(result) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: fmt.Sprintf("Error: %.6g ^ %d resulted in invalid number", aVal, exp),
				}},
				IsError: true,
			}, nil
		}
		resultText = fmt.Sprintf("%.6g ^ %d = %.6g", aVal, exp, result)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported operation '%s'", operation),
			}},
			IsError: true,
		}, nil
	}

	// Final validation of result
	if math.IsNaN(result) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: calculation resulted in invalid number (NaN)",
			}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Calculator Result:\n%s", resultText),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Numeric result: %.10g", result),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Result type: %s", getNumberType(result)),
			},
		},
		IsError: false,
	}, nil
}

// getNumberType returns a description of the number type
func getNumberType(val float64) string {
	if math.IsInf(val, 1) {
		return "positive infinity"
	}
	if math.IsInf(val, -1) {
		return "negative infinity"
	}
	if math.IsNaN(val) {
		return "not a number (NaN)"
	}
	if val == float64(int64(val)) {
		return "integer"
	}
	return "decimal"
}

// parseNumber converts interface{} to float64
func parseNumber(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", val)
	}
}

// power calculates a^b for integer exponents with overflow protection
func power(base float64, exp int) float64 {
	if exp == 0 {
		return 1
	}
	if exp < 0 {
		return 1 / power(base, -exp)
	}
	
	// Check for potential overflow
	if math.Abs(base) > 1 && exp > 100 {
		if base > 0 {
			return math.Inf(1)
		} else {
			if exp%2 == 0 {
				return math.Inf(1)
			} else {
				return math.Inf(-1)
			}
		}
	}
	
	// Use math.Pow for better precision with large exponents
	if exp > 50 {
		return math.Pow(base, float64(exp))
	}
	
	result := 1.0
	for i := 0; i < exp; i++ {
		result *= base
		// Check for infinity during calculation
		if math.IsInf(result, 0) {
			return result
		}
	}
	return result
}