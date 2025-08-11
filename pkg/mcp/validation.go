package mcp

import (
	"fmt"
	"strings"
)

// ValidateToolSchema validates a tool schema for completeness and correctness
func ValidateToolSchema(schema ToolSchema) error {
	if schema.Type == "" {
		return fmt.Errorf("schema type is required")
	}
	
	if schema.Type != "object" {
		return fmt.Errorf("schema type must be 'object', got '%s'", schema.Type)
	}
	
	if schema.Properties == nil {
		return fmt.Errorf("schema properties cannot be nil")
	}
	
	if len(schema.Properties) == 0 {
		return fmt.Errorf("schema must have at least one property")
	}
	
	// Validate that all required properties exist in properties
	for _, required := range schema.Required {
		if _, exists := schema.Properties[required]; !exists {
			return fmt.Errorf("required property '%s' not found in schema properties", required)
		}
	}
	
	// Validate individual properties
	for propName, propDef := range schema.Properties {
		if err := validateProperty(propName, propDef); err != nil {
			return fmt.Errorf("invalid property '%s': %w", propName, err)
		}
	}
	
	return nil
}

// validateProperty validates an individual property definition
func validateProperty(name string, property interface{}) error {
	if property == nil {
		return fmt.Errorf("property definition cannot be nil")
	}
	
	propMap, ok := property.(map[string]interface{})
	if !ok {
		return fmt.Errorf("property must be a map[string]interface{}")
	}
	
	// Type is required
	propType, exists := propMap["type"]
	if !exists {
		return fmt.Errorf("property type is required")
	}
	
	typeStr, ok := propType.(string)
	if !ok {
		return fmt.Errorf("property type must be a string")
	}
	
	// Validate type
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"integer": true,
		"boolean": true,
		"array":   true,
		"object":  true,
	}
	
	if !validTypes[typeStr] {
		return fmt.Errorf("invalid property type '%s'", typeStr)
	}
	
	// Validate type-specific constraints
	switch typeStr {
	case "string":
		if err := validateStringProperty(propMap); err != nil {
			return err
		}
	case "number", "integer":
		if err := validateNumericProperty(propMap); err != nil {
			return err
		}
	case "array":
		if err := validateArrayProperty(propMap); err != nil {
			return err
		}
	case "object":
		if err := validateObjectProperty(propMap); err != nil {
			return err
		}
	}
	
	return nil
}

// validateStringProperty validates string-specific constraints
func validateStringProperty(prop map[string]interface{}) error {
	// Validate enum if present
	if enum, exists := prop["enum"]; exists {
		enumSlice, ok := enum.([]interface{})
		if !ok {
			return fmt.Errorf("enum must be an array")
		}
		if len(enumSlice) == 0 {
			return fmt.Errorf("enum cannot be empty")
		}
		
		// All enum values must be strings
		for i, val := range enumSlice {
			if _, ok := val.(string); !ok {
				return fmt.Errorf("enum value at index %d must be a string", i)
			}
		}
	}
	
	// Validate minLength/maxLength
	if minLen, exists := prop["minLength"]; exists {
		if minLenNum, ok := minLen.(float64); ok {
			if minLenNum < 0 {
				return fmt.Errorf("minLength cannot be negative")
			}
		} else {
			return fmt.Errorf("minLength must be a number")
		}
	}
	
	if maxLen, exists := prop["maxLength"]; exists {
		if maxLenNum, ok := maxLen.(float64); ok {
			if maxLenNum < 0 {
				return fmt.Errorf("maxLength cannot be negative")
			}
		} else {
			return fmt.Errorf("maxLength must be a number")
		}
	}
	
	// Validate pattern if present
	if pattern, exists := prop["pattern"]; exists {
		if _, ok := pattern.(string); !ok {
			return fmt.Errorf("pattern must be a string")
		}
	}
	
	return nil
}

// validateNumericProperty validates number/integer-specific constraints
func validateNumericProperty(prop map[string]interface{}) error {
	// Validate minimum/maximum
	var minimum, maximum *float64
	
	if min, exists := prop["minimum"]; exists {
		if minNum, ok := min.(float64); ok {
			minimum = &minNum
		} else {
			return fmt.Errorf("minimum must be a number")
		}
	}
	
	if max, exists := prop["maximum"]; exists {
		if maxNum, ok := max.(float64); ok {
			maximum = &maxNum
		} else {
			return fmt.Errorf("maximum must be a number")
		}
	}
	
	if minimum != nil && maximum != nil && *minimum > *maximum {
		return fmt.Errorf("minimum (%f) cannot be greater than maximum (%f)", *minimum, *maximum)
	}
	
	// Validate exclusiveMinimum/exclusiveMaximum
	if exclMin, exists := prop["exclusiveMinimum"]; exists {
		if _, ok := exclMin.(bool); !ok {
			if _, ok := exclMin.(float64); !ok {
				return fmt.Errorf("exclusiveMinimum must be a boolean or number")
			}
		}
	}
	
	if exclMax, exists := prop["exclusiveMaximum"]; exists {
		if _, ok := exclMax.(bool); !ok {
			if _, ok := exclMax.(float64); !ok {
				return fmt.Errorf("exclusiveMaximum must be a boolean or number")
			}
		}
	}
	
	return nil
}

// validateArrayProperty validates array-specific constraints
func validateArrayProperty(prop map[string]interface{}) error {
	// Validate minItems/maxItems
	if minItems, exists := prop["minItems"]; exists {
		if minItemsNum, ok := minItems.(float64); ok {
			if minItemsNum < 0 {
				return fmt.Errorf("minItems cannot be negative")
			}
		} else {
			return fmt.Errorf("minItems must be a number")
		}
	}
	
	if maxItems, exists := prop["maxItems"]; exists {
		if maxItemsNum, ok := maxItems.(float64); ok {
			if maxItemsNum < 0 {
				return fmt.Errorf("maxItems cannot be negative")
			}
		} else {
			return fmt.Errorf("maxItems must be a number")
		}
	}
	
	// Validate uniqueItems
	if uniqueItems, exists := prop["uniqueItems"]; exists {
		if _, ok := uniqueItems.(bool); !ok {
			return fmt.Errorf("uniqueItems must be a boolean")
		}
	}
	
	return nil
}

// validateObjectProperty validates object-specific constraints
func validateObjectProperty(prop map[string]interface{}) error {
	// Validate minProperties/maxProperties
	if minProps, exists := prop["minProperties"]; exists {
		if minPropsNum, ok := minProps.(float64); ok {
			if minPropsNum < 0 {
				return fmt.Errorf("minProperties cannot be negative")
			}
		} else {
			return fmt.Errorf("minProperties must be a number")
		}
	}
	
	if maxProps, exists := prop["maxProperties"]; exists {
		if maxPropsNum, ok := maxProps.(float64); ok {
			if maxPropsNum < 0 {
				return fmt.Errorf("maxProperties cannot be negative")
			}
		} else {
			return fmt.Errorf("maxProperties must be a number")
		}
	}
	
	return nil
}

// ValidateToolParameters validates parameters against a tool schema
func ValidateToolParameters(params map[string]interface{}, schema ToolSchema) error {
	if err := ValidateToolSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}
	
	if params == nil {
		params = make(map[string]interface{})
	}
	
	// Check required parameters
	for _, required := range schema.Required {
		if _, exists := params[required]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}
	
	// Validate each provided parameter
	for paramName, paramValue := range params {
		propDef, exists := schema.Properties[paramName]
		if !exists {
			return fmt.Errorf("unknown parameter '%s'", paramName)
		}
		
		if err := validateParameterValue(paramName, paramValue, propDef); err != nil {
			return err
		}
	}
	
	return nil
}

// validateParameterValue validates a parameter value against its property definition
func validateParameterValue(name string, value interface{}, property interface{}) error {
	propMap, ok := property.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid property definition for '%s'", name)
	}
	
	expectedType, exists := propMap["type"]
	if !exists {
		return fmt.Errorf("property type not defined for '%s'", name)
	}
	
	typeStr, ok := expectedType.(string)
	if !ok {
		return fmt.Errorf("invalid type definition for '%s'", name)
	}
	
	// Type validation
	switch typeStr {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string, got %T", name, value)
		}
		return validateStringValue(name, value.(string), propMap)
		
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("parameter '%s' must be a number, got %T", name, value)
		}
		return validateNumberValue(name, value.(float64), propMap)
		
	case "integer":
		// JSON unmarshaling might give us float64 for integers
		if floatVal, ok := value.(float64); ok {
			if floatVal != float64(int64(floatVal)) {
				return fmt.Errorf("parameter '%s' must be an integer, got %f", name, floatVal)
			}
		} else if _, ok := value.(int); !ok {
			return fmt.Errorf("parameter '%s' must be an integer, got %T", name, value)
		}
		
		var numVal float64
		if floatVal, ok := value.(float64); ok {
			numVal = floatVal
		} else if intVal, ok := value.(int); ok {
			numVal = float64(intVal)
		}
		return validateNumberValue(name, numVal, propMap)
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter '%s' must be a boolean, got %T", name, value)
		}
		
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("parameter '%s' must be an array, got %T", name, value)
		}
		return validateArrayValue(name, value.([]interface{}), propMap)
		
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("parameter '%s' must be an object, got %T", name, value)
		}
		
	default:
		return fmt.Errorf("unsupported type '%s' for parameter '%s'", typeStr, name)
	}
	
	return nil
}

// validateStringValue validates a string value against its constraints
func validateStringValue(name, value string, prop map[string]interface{}) error {
	// Enum validation
	if enum, exists := prop["enum"]; exists {
		enumSlice := enum.([]interface{})
		found := false
		for _, enumVal := range enumSlice {
			if enumStr, ok := enumVal.(string); ok && enumStr == value {
				found = true
				break
			}
		}
		if !found {
			enumStrs := make([]string, len(enumSlice))
			for i, enumVal := range enumSlice {
				enumStrs[i] = enumVal.(string)
			}
			return fmt.Errorf("parameter '%s' must be one of [%s], got '%s'", name, strings.Join(enumStrs, ", "), value)
		}
	}
	
	// Length validation
	if minLen, exists := prop["minLength"]; exists {
		if minLenNum := minLen.(float64); float64(len(value)) < minLenNum {
			return fmt.Errorf("parameter '%s' must be at least %g characters long, got %d", name, minLenNum, len(value))
		}
	}
	
	if maxLen, exists := prop["maxLength"]; exists {
		if maxLenNum := maxLen.(float64); float64(len(value)) > maxLenNum {
			return fmt.Errorf("parameter '%s' must be at most %g characters long, got %d", name, maxLenNum, len(value))
		}
	}
	
	return nil
}

// validateNumberValue validates a number value against its constraints
func validateNumberValue(name string, value float64, prop map[string]interface{}) error {
	// Minimum validation
	if min, exists := prop["minimum"]; exists {
		if minNum := min.(float64); value < minNum {
			return fmt.Errorf("parameter '%s' must be >= %g, got %g", name, minNum, value)
		}
	}
	
	// Maximum validation
	if max, exists := prop["maximum"]; exists {
		if maxNum := max.(float64); value > maxNum {
			return fmt.Errorf("parameter '%s' must be <= %g, got %g", name, maxNum, value)
		}
	}
	
	return nil
}

// validateArrayValue validates an array value against its constraints
func validateArrayValue(name string, value []interface{}, prop map[string]interface{}) error {
	// Length validation
	if minItems, exists := prop["minItems"]; exists {
		if minItemsNum := minItems.(float64); float64(len(value)) < minItemsNum {
			return fmt.Errorf("parameter '%s' must have at least %g items, got %d", name, minItemsNum, len(value))
		}
	}
	
	if maxItems, exists := prop["maxItems"]; exists {
		if maxItemsNum := maxItems.(float64); float64(len(value)) > maxItemsNum {
			return fmt.Errorf("parameter '%s' must have at most %g items, got %d", name, maxItemsNum, len(value))
		}
	}
	
	// Unique items validation
	if uniqueItems, exists := prop["uniqueItems"]; exists {
		if shouldBeUnique := uniqueItems.(bool); shouldBeUnique {
			seen := make(map[interface{}]bool)
			for i, item := range value {
				if seen[item] {
					return fmt.Errorf("parameter '%s' must have unique items, duplicate found at index %d", name, i)
				}
				seen[item] = true
			}
		}
	}
	
	return nil
}

// SanitizeParameters cleans and normalizes parameter values
func SanitizeParameters(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return make(map[string]interface{})
	}
	
	sanitized := make(map[string]interface{})
	
	for key, value := range params {
		sanitized[key] = sanitizeValue(value)
	}
	
	return sanitized
}

// sanitizeValue sanitizes individual parameter values
func sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Trim whitespace and limit length for safety
		trimmed := strings.TrimSpace(v)
		if len(trimmed) > 10000 { // Reasonable limit
			trimmed = trimmed[:10000]
		}
		return trimmed
		
	case map[string]interface{}:
		// Recursively sanitize nested objects
		sanitized := make(map[string]interface{})
		for key, val := range v {
			sanitized[key] = sanitizeValue(val)
		}
		return sanitized
		
	case []interface{}:
		// Recursively sanitize arrays
		sanitized := make([]interface{}, len(v))
		for i, val := range v {
			sanitized[i] = sanitizeValue(val)
		}
		return sanitized
		
	default:
		// Return other types as-is
		return value
	}
}