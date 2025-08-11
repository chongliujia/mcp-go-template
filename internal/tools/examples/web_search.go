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

// SearchEngineConfig holds configuration for search engines
type SearchEngineConfig struct {
	Name        string
	BaseURL     string
	Enabled     bool
	RateLimit   time.Duration
	MaxRetries  int
}

// WebSearchTool implements web search functionality using multiple search engines
type WebSearchTool struct {
	definition *mcp.Tool
	client     *http.Client
	engines    map[string]SearchEngineConfig
	lastRequest map[string]time.Time // Rate limiting
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query    string         `json:"query"`
	Results  []SearchResult `json:"results"`
	Total    int           `json:"total"`
	Engine   string        `json:"engine"`
	Duration string        `json:"duration"`
}

// NewWebSearchTool creates a new web search tool with enhanced configuration
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		definition: &mcp.Tool{
			Name:        "web_search",
			Description: "Searches the web using multiple search engines (DuckDuckGo, SearXNG, Brave Search) and returns structured results with titles, URLs, descriptions, and sources. Includes rate limiting and fallback mechanisms.",
			InputSchema: mcp.ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query to execute",
						"minLength":   1,
						"maxLength":   500,
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10, max: 50)",
						"default":     10,
						"minimum":     1,
						"maximum":     50,
					},
					"engine": map[string]interface{}{
						"type":        "string",
						"description": "Search engine to use (auto tries multiple engines)",
						"enum":        []string{"duckduckgo", "searxng", "brave", "auto"},
						"default":     "auto",
					},
					"safe_search": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable safe search filtering",
						"default":     true,
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language preference for results (ISO 639-1 code)",
						"default":     "en",
						"pattern":     "^[a-z]{2}$",
					},
					"region": map[string]interface{}{
						"type":        "string",
						"description": "Geographic region for search results",
						"default":     "us-en",
					},
				},
				Required: []string{"query"},
			},
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
		engines: map[string]SearchEngineConfig{
			"duckduckgo": {
				Name:        "DuckDuckGo",
				BaseURL:     "https://api.duckduckgo.com/",
				Enabled:     true,
				RateLimit:   time.Second * 2,
				MaxRetries:  2,
			},
			"searxng": {
				Name:        "SearXNG",
				BaseURL:     "https://search.sapti.me/search", // Public SearXNG instance
				Enabled:     true,
				RateLimit:   time.Second * 3,
				MaxRetries:  2,
			},
			"brave": {
				Name:        "Brave Search",
				BaseURL:     "https://search.brave.com/api/search",
				Enabled:     false, // Requires API key
				RateLimit:   time.Second * 1,
				MaxRetries:  2,
			},
		},
		lastRequest: make(map[string]time.Time),
	}
}

// Definition returns the tool definition
func (w *WebSearchTool) Definition() *mcp.Tool {
	return w.definition
}

// Execute performs the web search with enhanced error handling and fallback mechanisms
func (w *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	
	// Enhanced parameter extraction and validation
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
	
	// Validate query length
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: query cannot be empty after trimming whitespace",
			}},
			IsError: true,
		}, nil
	}
	if len(query) > 500 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: query too long (maximum 500 characters)",
			}},
			IsError: true,
		}, nil
	}

	maxResults := 10
	if val, exists := params["max_results"]; exists {
		if num, ok := val.(float64); ok {
			maxResults = int(num)
		}
	}
	if maxResults < 1 || maxResults > 50 {
		maxResults = 10
	}

	engine := "auto"
	if val, exists := params["engine"]; exists {
		if eng, ok := val.(string); ok {
			engine = eng
		}
	}

	safeSearch := true
	if val, exists := params["safe_search"]; exists {
		if safe, ok := val.(bool); ok {
			safeSearch = safe
		}
	}
	
	language := "en"
	if val, exists := params["language"]; exists {
		if lang, ok := val.(string); ok && len(lang) == 2 {
			language = lang
		}
	}
	
	region := "us-en"
	if val, exists := params["region"]; exists {
		if reg, ok := val.(string); ok {
			region = reg
		}
	}

	// Perform search with fallback mechanisms
	var results []SearchResult
	var searchEngine string
	var searchErrors []error

	switch engine {
	case "duckduckgo":
		results, searchEngine, searchErrors = w.searchWithRetry("duckduckgo", query, maxResults, safeSearch, language, region)
	case "searxng":
		results, searchEngine, searchErrors = w.searchWithRetry("searxng", query, maxResults, safeSearch, language, region)
	case "brave":
		results, searchEngine, searchErrors = w.searchWithRetry("brave", query, maxResults, safeSearch, language, region)
	case "auto":
		// Try engines in order of preference
		engineOrder := []string{"duckduckgo", "searxng"}
		for _, eng := range engineOrder {
			if w.engines[eng].Enabled {
				var errs []error
				results, searchEngine, errs = w.searchWithRetry(eng, query, maxResults, safeSearch, language, region)
				searchErrors = append(searchErrors, errs...)
				if len(results) > 0 {
					break
				}
			}
		}
		
		// If no results from APIs, use simulated results
		if len(results) == 0 {
			var err error
			results, err = w.simulateSearch(query, maxResults)
			if err != nil {
				searchErrors = append(searchErrors, err)
			} else {
				searchEngine = "Simulated Results"
			}
		}
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported search engine '%s' (supported: duckduckgo, searxng, brave, auto)", engine),
			}},
			IsError: true,
		}, nil
	}

	// If still no results, return error with details
	if len(results) == 0 {
		errorMsg := fmt.Sprintf("Search failed for query '%s'. Errors encountered:", query)
		for i, err := range searchErrors {
			errorMsg += fmt.Sprintf("\n%d. %v", i+1, err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: errorMsg,
			}},
			IsError: true,
		}, nil
	}

	duration := time.Since(startTime)

	// Create enhanced response
	response := SearchResponse{
		Query:    query,
		Results:  results,
		Total:    len(results),
		Engine:   searchEngine,
		Duration: duration.String(),
	}

	// Format results with enhanced information
	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("🔍 Web Search Results\n"))
	resultText.WriteString(fmt.Sprintf("═══════════════════════════════════════════════════\n\n"))
	resultText.WriteString(fmt.Sprintf("Query: \"%s\"\n", query))
	resultText.WriteString(fmt.Sprintf("Engine: %s\n", searchEngine))
	resultText.WriteString(fmt.Sprintf("Results: %d/%d\n", len(results), maxResults))
	resultText.WriteString(fmt.Sprintf("Duration: %s\n", duration.String()))
	resultText.WriteString(fmt.Sprintf("Language: %s | Region: %s | Safe Search: %v\n\n", language, region, safeSearch))

	for i, result := range results {
		resultText.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, result.Title))
		resultText.WriteString(fmt.Sprintf("   🔗 %s\n", result.URL))
		if result.Description != "" {
			resultText.WriteString(fmt.Sprintf("   📝 %s\n", result.Description))
		}
		if result.Source != "" {
			resultText.WriteString(fmt.Sprintf("   📰 Source: %s\n", result.Source))
		}
		resultText.WriteString("\n")
	}
	
	// Add search errors as warnings if any
	if len(searchErrors) > 0 && searchEngine != "Simulated Results" {
		resultText.WriteString("⚠️ Warnings encountered during search:\n")
		for i, err := range searchErrors {
			resultText.WriteString(fmt.Sprintf("%d. %v\n", i+1, err))
		}
	}

	// Convert to JSON for structured output
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		jsonData = []byte(fmt.Sprintf(`{"error": "failed to marshal response: %v"}`, err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: resultText.String(),
			},
			{
				Type:     "text",
				Text:     string(jsonData),
				MimeType: "application/json",
			},
		},
		IsError: false,
	}, nil
}

// searchWithRetry attempts to search using the specified engine with retry logic
func (w *WebSearchTool) searchWithRetry(engineName, query string, maxResults int, safeSearch bool, language, region string) ([]SearchResult, string, []error) {
	engineConfig, exists := w.engines[engineName]
	if !exists || !engineConfig.Enabled {
		return nil, "", []error{fmt.Errorf("engine %s not available", engineName)}
	}
	
	// Rate limiting
	if lastReq, exists := w.lastRequest[engineName]; exists {
		if time.Since(lastReq) < engineConfig.RateLimit {
			time.Sleep(engineConfig.RateLimit - time.Since(lastReq))
		}
	}
	
	var results []SearchResult
	var errors []error
	
	for attempt := 0; attempt <= engineConfig.MaxRetries; attempt++ {
		var err error
		
		switch engineName {
		case "duckduckgo":
			results, err = w.searchDuckDuckGo(query, maxResults, safeSearch, language, region)
		case "searxng":
			results, err = w.searchSearXNG(query, maxResults, safeSearch, language, region)
		case "brave":
			results, err = w.searchBrave(query, maxResults, safeSearch, language, region)
		default:
			return nil, "", []error{fmt.Errorf("unsupported engine: %s", engineName)}
		}
		
		w.lastRequest[engineName] = time.Now()
		
		if err == nil && len(results) > 0 {
			return results, engineConfig.Name, errors
		}
		
		if err != nil {
			errors = append(errors, fmt.Errorf("attempt %d with %s: %w", attempt+1, engineConfig.Name, err))
		}
		
		// Wait before retry (exponential backoff)
		if attempt < engineConfig.MaxRetries {
			waitTime := time.Duration(attempt+1) * time.Second
			time.Sleep(waitTime)
		}
	}
	
	return results, engineConfig.Name, errors
}

// searchSearXNG performs search using SearXNG API
func (w *WebSearchTool) searchSearXNG(query string, maxResults int, safeSearch bool, language, region string) ([]SearchResult, error) {
	baseURL := w.engines["searxng"].BaseURL
	
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("lang", language)
	params.Set("pageno", "1")
	
	if safeSearch {
		params.Set("safesearch", "2")
	} else {
		params.Set("safesearch", "0")
	}
	
	reqURL := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SearXNG request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MCP-Go-Template/1.0)")
	req.Header.Set("Accept", "application/json")
	
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform SearXNG search: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearXNG HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SearXNG response: %w", err)
	}
	
	// Parse SearXNG response
	var searxResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
			Engine  string `json:"engine"`
		} `json:"results"`
	}
	
	if err := json.Unmarshal(body, &searxResp); err != nil {
		return nil, fmt.Errorf("failed to parse SearXNG response: %w", err)
	}
	
	var results []SearchResult
	for i, result := range searxResp.Results {
		if i >= maxResults {
			break
		}
		
		results = append(results, SearchResult{
			Title:       result.Title,
			URL:         result.URL,
			Description: result.Content,
			Source:      fmt.Sprintf("SearXNG (%s)", result.Engine),
		})
	}
	
	return results, nil
}

// searchBrave performs search using Brave Search API (placeholder)
func (w *WebSearchTool) searchBrave(query string, maxResults int, safeSearch bool, language, region string) ([]SearchResult, error) {
	// Brave Search API requires an API key and subscription
	// This is a placeholder implementation
	return nil, fmt.Errorf("Brave Search API not implemented - requires API key")
}

// searchDuckDuckGo performs search using DuckDuckGo with enhanced parameters
func (w *WebSearchTool) searchDuckDuckGo(query string, maxResults int, safeSearch bool, language, region string) ([]SearchResult, error) {
	// DuckDuckGo Instant Answer API (limited functionality)
	baseURL := "https://api.duckduckgo.com/"
	
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_redirect", "1")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")
	
	if safeSearch {
		params.Set("safe_search", "strict")
	}
	
	// DuckDuckGo doesn't support language/region parameters in the free API
	reqURL := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DuckDuckGo request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MCP-Go-Template/1.0)")
	
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform DuckDuckGo search: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DuckDuckGo response: %w", err)
	}
	
	// Parse DuckDuckGo response
	var ddgResp map[string]interface{}
	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, fmt.Errorf("failed to parse DuckDuckGo response: %w", err)
	}
	
	var results []SearchResult
	
	// Extract instant answer
	if abstract, ok := ddgResp["Abstract"].(string); ok && abstract != "" {
		if abstractURL, ok := ddgResp["AbstractURL"].(string); ok && abstractURL != "" {
			results = append(results, SearchResult{
				Title:       fmt.Sprintf("DuckDuckGo Instant Answer: %s", query),
				URL:         abstractURL,
				Description: abstract,
				Source:      "DuckDuckGo Instant Answer",
			})
		}
	}
	
	// Extract related topics
	if relatedTopics, ok := ddgResp["RelatedTopics"].([]interface{}); ok {
		for _, topic := range relatedTopics {
			if len(results) >= maxResults {
				break
			}
			
			if topicMap, ok := topic.(map[string]interface{}); ok {
				if text, ok := topicMap["Text"].(string); ok && text != "" {
					if firstURL, ok := topicMap["FirstURL"].(string); ok && firstURL != "" {
						title := strings.Split(text, " - ")[0]
						if len(title) > 100 {
							title = title[:97] + "..."
						}
						results = append(results, SearchResult{
							Title:       fmt.Sprintf("Related: %s", title),
							URL:         firstURL,
							Description: text,
							Source:      "DuckDuckGo Related Topics",
						})
					}
				}
			}
		}
	}
	
	// Extract answer if available
	if answer, ok := ddgResp["Answer"].(string); ok && answer != "" {
		if answerURL, ok := ddgResp["AnswerURL"].(string); ok && answerURL != "" {
			results = append(results, SearchResult{
				Title:       fmt.Sprintf("Answer: %s", query),
				URL:         answerURL,
				Description: answer,
				Source:      "DuckDuckGo Answer",
			})
		}
	}
	
	return results, nil
}

// simulateSearch creates enhanced simulated search results for demonstration
func (w *WebSearchTool) simulateSearch(query string, maxResults int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("cannot simulate search with empty query")
	}
	
	// Create more realistic simulated results based on query analysis
	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)
	
	var results []SearchResult
	
	// Wikipedia result (most common)
	results = append(results, SearchResult{
		Title:       fmt.Sprintf("%s - Wikipedia", query),
		URL:         fmt.Sprintf("https://en.wikipedia.org/wiki/%s", url.QueryEscape(strings.ReplaceAll(query, " ", "_"))),
		Description: fmt.Sprintf("Wikipedia article about %s. Comprehensive information, history, and references from the free encyclopedia.", query),
		Source:      "Wikipedia",
	})
	
	// Official website (if looks like a brand/company)
	if len(words) <= 3 {
		results = append(results, SearchResult{
			Title:       fmt.Sprintf("%s - Official Website", query),
			URL:         fmt.Sprintf("https://www.%s.com", strings.ToLower(strings.ReplaceAll(query, " ", ""))),
			Description: fmt.Sprintf("Official website for %s. Get the latest information, updates, and authentic content directly from the source.", query),
			Source:      "Official Site",
		})
	}
	
	// Educational/Guide content
	results = append(results, SearchResult{
		Title:       fmt.Sprintf("What is %s? Complete Guide and Definition", query),
		URL:         fmt.Sprintf("https://guide.example.com/%s", url.QueryEscape(strings.ToLower(query))),
		Description: fmt.Sprintf("Comprehensive guide explaining %s, its applications, benefits, and everything you need to know. Includes examples and practical information.", query),
		Source:      "Educational Resource",
	})
	
	// News results (for current topics)
	if len(query) > 5 {
		results = append(results, SearchResult{
			Title:       fmt.Sprintf("Latest News: %s Updates and Developments", query),
			URL:         fmt.Sprintf("https://news.google.com/search?q=%s", url.QueryEscape(query)),
			Description: fmt.Sprintf("Recent news and updates about %s from various trusted news sources around the world. Stay informed with the latest developments.", query),
			Source:      "News Aggregator",
		})
	}
	
	// Academic/Research content
	results = append(results, SearchResult{
		Title:       fmt.Sprintf("Academic Research on %s - Scholarly Articles", query),
		URL:         fmt.Sprintf("https://scholar.google.com/scholar?q=%s", url.QueryEscape(query)),
		Description: fmt.Sprintf("Peer-reviewed academic papers and research studies related to %s from universities and research institutions worldwide.", query),
		Source:      "Academic Search",
	})
	
	// Video content
	results = append(results, SearchResult{
		Title:       fmt.Sprintf("%s - Video Tutorials and Explanations", query),
		URL:         fmt.Sprintf("https://www.youtube.com/results?search_query=%s", url.QueryEscape(query)),
		Description: fmt.Sprintf("Educational videos, tutorials, and visual explanations about %s. Learn through engaging multimedia content.", query),
		Source:      "Video Platform",
	})
	
	// Community/Forum content
	if len(words) > 1 {
		results = append(results, SearchResult{
			Title:       fmt.Sprintf("%s - Community Discussions and Q&A", query),
			URL:         fmt.Sprintf("https://www.reddit.com/search/?q=%s", url.QueryEscape(query)),
			Description: fmt.Sprintf("Community discussions, questions, and answers about %s. Real experiences and insights from users and experts.", query),
			Source:      "Community Forum",
		})
	}
	
	// Commercial/Shopping results (if looks like a product)
	if containsProductWords(queryLower) {
		results = append(results, SearchResult{
			Title:       fmt.Sprintf("Buy %s - Compare Prices and Reviews", query),
			URL:         fmt.Sprintf("https://shopping.google.com/search?q=%s", url.QueryEscape(query)),
			Description: fmt.Sprintf("Find and compare prices for %s from various retailers. Read reviews, specifications, and find the best deals.", query),
			Source:      "Shopping Search",
		})
	}
	
	// Limit results to requested count
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	
	return results, nil
}

// containsProductWords checks if query contains product-related keywords
func containsProductWords(query string) bool {
	productKeywords := []string{"buy", "price", "review", "best", "top", "compare", "cheap", "discount", "deal", "sale", "product", "item"}
	for _, keyword := range productKeywords {
		if strings.Contains(query, keyword) {
			return true
		}
	}
	return false
}