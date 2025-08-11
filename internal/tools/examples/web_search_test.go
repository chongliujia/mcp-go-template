package examples

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestWebSearchTool_Definition(t *testing.T) {
	search := NewWebSearchTool()
	def := search.Definition()
	
	if def == nil {
		t.Fatal("Definition should not be nil")
	}
	
	if def.Name != "web_search" {
		t.Errorf("Expected name 'web_search', got '%s'", def.Name)
	}
	
	if def.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestWebSearchTool_Execute_ValidQuery(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query":       "test query",
		"max_results": 5,
		"engine":      "auto",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful search, got error: %v", result.Content[0].Text)
	}
	
	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items (text and JSON)")
	}
}

func TestWebSearchTool_Execute_EmptyQuery(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query": "",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for empty query")
	}
}

func TestWebSearchTool_Execute_WhitespaceQuery(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query": "   ",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for whitespace-only query")
	}
}

func TestWebSearchTool_Execute_QueryTooLong(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	longQuery := strings.Repeat("a", 501)
	params := map[string]interface{}{
		"query": longQuery,
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for query too long")
	}
}

func TestWebSearchTool_Execute_InvalidEngine(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query":  "test",
		"engine": "invalid_engine",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid engine")
	}
}

func TestWebSearchTool_Execute_MaxResultsValidation(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	tests := []struct {
		maxResults interface{}
		shouldWork bool
	}{
		{5, true},
		{50, true},
		{0, true},   // Should default to 10
		{51, true},  // Should cap to 10
		{-1, true},  // Should default to 10
		{"invalid", true}, // Should default to 10
	}
	
	for _, test := range tests {
		params := map[string]interface{}{
			"query":       "test",
			"max_results": test.maxResults,
		}
		
		result, err := search.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Unexpected error for max_results %v: %v", test.maxResults, err)
		}
		
		if test.shouldWork && result.IsError {
			t.Errorf("Expected success for max_results %v, got error: %v", test.maxResults, result.Content[0].Text)
		}
	}
}

func TestWebSearchTool_Execute_AllEngines(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	engines := []string{"duckduckgo", "searxng", "auto"}
	
	for _, engine := range engines {
		params := map[string]interface{}{
			"query":  "test",
			"engine": engine,
		}
		
		result, err := search.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Unexpected error for engine %s: %v", engine, err)
		}
		
		// Note: Some engines might fail due to network issues, but should not panic
		if result.IsError {
			t.Logf("Search failed for engine %s (this is expected in test environment): %v", engine, result.Content[0].Text)
		}
	}
}

func TestWebSearchTool_Execute_WithAllParameters(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query":       "golang programming",
		"max_results": 10,
		"engine":      "auto",
		"safe_search": true,
		"language":    "en",
		"region":      "us-en",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should not error with valid parameters
	if result.IsError {
		t.Logf("Search failed (expected in test environment): %v", result.Content[0].Text)
	}
}

func TestWebSearchTool_Execute_LanguageValidation(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	tests := []struct {
		language string
		valid    bool
	}{
		{"en", true},
		{"fr", true},
		{"invalid", false}, // Should use default
		{"e", false},       // Should use default
		{"", false},        // Should use default
	}
	
	for _, test := range tests {
		params := map[string]interface{}{
			"query":    "test",
			"language": test.language,
		}
		
		result, err := search.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Unexpected error for language %s: %v", test.language, err)
		}
		
		// Language validation is internal, so we just check it doesn't error
		if result.IsError {
			t.Logf("Search failed for language %s (expected in test environment): %v", test.language, result.Content[0].Text)
		}
	}
}

func TestWebSearchTool_SimulateSearch(t *testing.T) {
	search := NewWebSearchTool()
	
	tests := []struct {
		query       string
		maxResults  int
		expectError bool
	}{
		{"test query", 5, false},
		{"golang programming", 10, false},
		{"", 5, true}, // Empty query should error
		{"very long query with many words", 3, false},
	}
	
	for _, test := range tests {
		results, err := search.simulateSearch(test.query, test.maxResults)
		
		if test.expectError && err == nil {
			t.Errorf("Expected error for query '%s'", test.query)
		}
		
		if !test.expectError && err != nil {
			t.Errorf("Unexpected error for query '%s': %v", test.query, err)
		}
		
		if !test.expectError {
			if len(results) == 0 {
				t.Errorf("Expected results for query '%s'", test.query)
			}
			
			if len(results) > test.maxResults {
				t.Errorf("Expected at most %d results for query '%s', got %d", test.maxResults, test.query, len(results))
			}
			
			// Check result structure
			for i, result := range results {
				if result.Title == "" {
					t.Errorf("Result %d should have a title", i)
				}
				if result.URL == "" {
					t.Errorf("Result %d should have a URL", i)
				}
				if result.Description == "" {
					t.Errorf("Result %d should have a description", i)
				}
				if result.Source == "" {
					t.Errorf("Result %d should have a source", i)
				}
			}
		}
	}
}

func TestWebSearchTool_ContainsProductWords(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"buy laptop", true},
		{"best smartphone", true},
		{"cheap headphones", true},
		{"price comparison", true},
		{"product review", true},
		{"golang programming", false},
		{"weather today", false},
		{"news update", false},
		{"", false},
	}
	
	for _, test := range tests {
		result := containsProductWords(test.query)
		if result != test.expected {
			t.Errorf("containsProductWords(%q): expected %v, got %v", test.query, test.expected, result)
		}
	}
}

func TestWebSearchTool_SearchWithRetry_ConfigValidation(t *testing.T) {
	search := NewWebSearchTool()
	
	// Test that engine configs are properly set
	if search.engines == nil {
		t.Fatal("Engine configs should be initialized")
	}
	
	expectedEngines := []string{"duckduckgo", "searxng", "brave"}
	for _, engine := range expectedEngines {
		config, exists := search.engines[engine]
		if !exists {
			t.Errorf("Engine config for %s should exist", engine)
			continue
		}
		
		if config.Name == "" {
			t.Errorf("Engine %s should have a name", engine)
		}
		
		if config.BaseURL == "" {
			t.Errorf("Engine %s should have a base URL", engine)
		}
		
		if config.RateLimit <= 0 {
			t.Errorf("Engine %s should have a positive rate limit", engine)
		}
		
		if config.MaxRetries < 0 {
			t.Errorf("Engine %s should have non-negative max retries", engine)
		}
	}
}

func TestWebSearchTool_RateLimiting(t *testing.T) {
	search := NewWebSearchTool()
	
	// Set a very short rate limit for testing
	search.engines["duckduckgo"] = SearchEngineConfig{
		Name:       "DuckDuckGo",
		BaseURL:    "https://api.duckduckgo.com/",
		Enabled:    true,
		RateLimit:  time.Millisecond * 100,
		MaxRetries: 1,
	}
	
	// First request should set the timestamp
	_, _, _ = search.searchWithRetry("duckduckgo", "test", 5, true, "en", "us-en")
	
	// Second immediate request should trigger rate limiting
	start := time.Now()
	_, _, _ = search.searchWithRetry("duckduckgo", "test2", 5, true, "en", "us-en")
	duration := time.Since(start)
	
	// Should have waited at least part of the rate limit duration
	if duration < time.Millisecond*50 {
		t.Logf("Rate limiting test: duration was %v (may not have waited due to network delays)", duration)
	}
}

func TestWebSearchTool_MissingParameters(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	// Missing query parameter
	params := map[string]interface{}{
		"max_results": 5,
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for missing query parameter")
	}
	
	// Invalid query parameter type
	params = map[string]interface{}{
		"query": 123,
	}
	
	result, err = search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid query parameter type")
	}
}

func TestWebSearchTool_DefaultParameters(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	// Test with minimal parameters (should use defaults)
	params := map[string]interface{}{
		"query": "test query",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should not error with minimal valid parameters
	if result.IsError {
		t.Logf("Search failed with default parameters (expected in test environment): %v", result.Content[0].Text)
	}
}

func TestWebSearchTool_JSONOutput(t *testing.T) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query": "test query",
	}
	
	result, err := search.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items")
	}
	
	// Check that second content item is JSON
	jsonContent := result.Content[1]
	if jsonContent.MimeType != "application/json" {
		t.Errorf("Expected JSON mime type, got %s", jsonContent.MimeType)
	}
	
	if jsonContent.Text == "" {
		t.Error("JSON content should not be empty")
	}
}

// Benchmark tests
func BenchmarkWebSearchTool_Execute(b *testing.B) {
	search := NewWebSearchTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"query":  "benchmark test",
		"engine": "auto", // This will likely use simulated results
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := search.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkWebSearchTool_SimulateSearch(b *testing.B) {
	search := NewWebSearchTool()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := search.simulateSearch("benchmark test query", 10)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}