package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// WebSearchTool provides web search capabilities for research
type WebSearchTool struct {
	httpClient *http.Client
}

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

// NewWebSearchTool creates a new web search tool instance
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Definition returns the tool definition for MCP
func (w *WebSearchTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web for information using various search engines. Essential for deep research tasks.",
		InputSchema: mcp.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query to execute",
				},
				"engine": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"duckduckgo", "bing", "google"},
					"description": "The search engine to use (default: duckduckgo)",
					"default":     "duckduckgo",
				},
				"max_results": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return (default: 10, max: 50)",
					"default":     10,
					"minimum":     1,
					"maximum":     50,
				},
				"safe_search": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable safe search filtering (default: true)",
					"default":     true,
				},
			},
			Required: []string{"query"},
		},
	}
}

// Execute performs the web search
func (w *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract parameters
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: query parameter is required and must be a non-empty string",
			}},
			IsError: true,
		}, nil
	}

	engine := "duckduckgo"
	if e, ok := params["engine"].(string); ok {
		engine = e
	}

	maxResults := 10
	if mr, ok := params["max_results"]; ok {
		if mrFloat, ok := mr.(float64); ok {
			maxResults = int(mrFloat)
		}
	}
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 50 {
		maxResults = 50
	}

	safeSearch := true
	if ss, ok := params["safe_search"].(bool); ok {
		safeSearch = ss
	}

	// Perform search based on engine
	var results []SearchResult
	var err error

	switch engine {
	case "duckduckgo":
		results, err = w.searchDuckDuckGo(ctx, query, maxResults, safeSearch)
	case "bing":
		results, err = w.searchBing(ctx, query, maxResults, safeSearch)
	case "google":
		results, err = w.searchGoogle(ctx, query, maxResults, safeSearch)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported search engine '%s'", engine),
			}},
			IsError: true,
		}, nil
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Search failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Format results
	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("No search results found for query: '%s'", query),
			}},
			IsError: false,
		}, nil
	}

	// Create formatted response
	var responseBuilder strings.Builder
	responseBuilder.WriteString(fmt.Sprintf("Search Results for '%s' (using %s):\n\n", query, engine))

	for i, result := range results {
		responseBuilder.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, result.Title))
		responseBuilder.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Summary != "" {
			responseBuilder.WriteString(fmt.Sprintf("   Summary: %s\n", result.Summary))
		}
		responseBuilder.WriteString("\n")
	}

	// Also provide JSON format for programmatic access
	jsonResults, _ := json.MarshalIndent(results, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: responseBuilder.String(),
			},
			{
				Type:     "text",
				Text:     string(jsonResults),
				MimeType: "application/json",
			},
		},
		IsError: false,
	}, nil
}

// searchDuckDuckGo performs search using DuckDuckGo Instant Answer API
func (w *WebSearchTool) searchDuckDuckGo(ctx context.Context, query string, maxResults int, safeSearch bool) ([]SearchResult, error) {
	// DuckDuckGo Instant Answer API
	baseURL := "https://api.duckduckgo.com/"
	
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("no_html", "1")
	params.Add("skip_disambig", "1")
	if safeSearch {
		params.Add("safe_search", "strict")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "MCP-Go-Template/1.0 (Research Tool)")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ddgResponse map[string]interface{}
	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []SearchResult

	// Extract abstract if available
	if abstract, ok := ddgResponse["Abstract"].(string); ok && abstract != "" {
		if abstractURL, ok := ddgResponse["AbstractURL"].(string); ok {
			results = append(results, SearchResult{
				Title:   "Abstract",
				URL:     abstractURL,
				Summary: abstract,
			})
		}
	}

	// Extract related topics
	if relatedTopics, ok := ddgResponse["RelatedTopics"].([]interface{}); ok {
		for i, topic := range relatedTopics {
			if i >= maxResults {
				break
			}
			if topicMap, ok := topic.(map[string]interface{}); ok {
				title := ""
				url := ""
				summary := ""

				if t, ok := topicMap["Text"].(string); ok {
					summary = t
					// Extract title from text (first part before " - ")
					if parts := strings.Split(t, " - "); len(parts) > 0 {
						title = parts[0]
					}
				}
				if u, ok := topicMap["FirstURL"].(string); ok {
					url = u
				}

				if title != "" && url != "" {
					results = append(results, SearchResult{
						Title:   title,
						URL:     url,
						Summary: summary,
					})
				}
			}
		}
	}

	return results, nil
}

// searchBing performs search using Bing (placeholder implementation)
func (w *WebSearchTool) searchBing(ctx context.Context, query string, maxResults int, safeSearch bool) ([]SearchResult, error) {
	// Note: This would require Bing Search API key
	// For now, return a placeholder response
	return []SearchResult{
		{
			Title:   "Bing Search Not Implemented",
			URL:     "https://www.bing.com/search?q=" + url.QueryEscape(query),
			Summary: "Bing search requires API key configuration. Please use DuckDuckGo for now.",
		},
	}, nil
}

// searchGoogle performs search using Google (placeholder implementation)
func (w *WebSearchTool) searchGoogle(ctx context.Context, query string, maxResults int, safeSearch bool) ([]SearchResult, error) {
	// Note: This would require Google Custom Search API key
	// For now, return a placeholder response
	return []SearchResult{
		{
			Title:   "Google Search Not Implemented",
			URL:     "https://www.google.com/search?q=" + url.QueryEscape(query),
			Summary: "Google search requires API key configuration. Please use DuckDuckGo for now.",
		},
	}, nil
}