package examples

import (
	"context"
	"testing"
)

func TestCalculatorTool_Definition(t *testing.T) {
	calc := NewCalculatorTool()
	def := calc.Definition()
	
	if def == nil {
		t.Fatal("Definition should not be nil")
	}
	
	if def.Name != "calculator" {
		t.Errorf("Expected name 'calculator', got '%s'", def.Name)
	}
	
	if def.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestCalculatorTool_Execute_Addition(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "add",
		"a":         5.0,
		"b":         3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful calculation, got error: %v", result.Content[0].Text)
	}
	
	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items")
	}
}

func TestCalculatorTool_Execute_Subtraction(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "subtract",
		"a":         10.0,
		"b":         4.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful calculation, got error: %v", result.Content[0].Text)
	}
}

func TestCalculatorTool_Execute_Multiplication(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "multiply",
		"a":         6.0,
		"b":         7.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful calculation, got error: %v", result.Content[0].Text)
	}
}

func TestCalculatorTool_Execute_Division(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "divide",
		"a":         15.0,
		"b":         3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful calculation, got error: %v", result.Content[0].Text)
	}
}

func TestCalculatorTool_Execute_DivisionByZero(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "divide",
		"a":         10.0,
		"b":         0.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for division by zero")
	}
}

func TestCalculatorTool_Execute_Power(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "power",
		"a":         2.0,
		"b":         3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful calculation, got error: %v", result.Content[0].Text)
	}
}

func TestCalculatorTool_Execute_PowerNonInteger(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "power",
		"a":         2.0,
		"b":         3.5,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for non-integer exponent")
	}
}

func TestCalculatorTool_Execute_PowerTooLarge(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "power",
		"a":         2.0,
		"b":         1001.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for exponent too large")
	}
}

func TestCalculatorTool_Execute_InvalidOperation(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"operation": "invalid",
		"a":         5.0,
		"b":         3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid operation")
	}
}

func TestCalculatorTool_Execute_MissingParameters(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	// Missing operation
	params := map[string]interface{}{
		"a": 5.0,
		"b": 3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for missing operation")
	}
	
	// Missing parameter a
	params = map[string]interface{}{
		"operation": "add",
		"b":         3.0,
	}
	
	result, err = calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for missing parameter a")
	}
}

func TestCalculatorTool_Execute_InvalidParameterTypes(t *testing.T) {
	calc := NewCalculatorTool()
	ctx := context.Background()
	
	// Invalid operation type
	params := map[string]interface{}{
		"operation": 123,
		"a":         5.0,
		"b":         3.0,
	}
	
	result, err := calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid operation type")
	}
	
	// Invalid parameter type
	params = map[string]interface{}{
		"operation": "add",
		"a":         "not a number",
		"b":         3.0,
	}
	
	result, err = calc.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid parameter type")
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		hasError bool
	}{
		{float64(5.5), 5.5, false},
		{float32(5.5), 5.5, false},
		{int(5), 5.0, false},
		{int64(5), 5.0, false},
		{int32(5), 5.0, false},
		{"5.5", 5.5, false},
		{"invalid", 0, true},
		{true, 0, true},
	}
	
	for _, test := range tests {
		result, err := parseNumber(test.input)
		if test.hasError && err == nil {
			t.Errorf("Expected error for input %v", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input %v: %v", test.input, err)
		}
		if !test.hasError && result != test.expected {
			t.Errorf("Expected %f for input %v, got %f", test.expected, test.input, result)
		}
	}
}

func TestPower(t *testing.T) {
	tests := []struct {
		base     float64
		exp      int
		expected float64
	}{
		{2, 0, 1},
		{2, 3, 8},
		{5, 2, 25},
		{2, -3, 0.125},
		{10, 1, 10},
		{0, 5, 0},
	}
	
	for _, test := range tests {
		result := power(test.base, test.exp)
		if result != test.expected {
			t.Errorf("power(%f, %d): expected %f, got %f", test.base, test.exp, test.expected, result)
		}
	}
}

func TestGetNumberType(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{5.0, "integer"},
		{5.5, "decimal"},
		{0.0, "integer"},
		{-5.0, "integer"},
		{-5.5, "decimal"},
	}
	
	for _, test := range tests {
		result := getNumberType(test.input)
		if result != test.expected {
			t.Errorf("getNumberType(%f): expected %s, got %s", test.input, test.expected, result)
		}
	}
}