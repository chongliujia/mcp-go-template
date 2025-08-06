package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// DocumentAnalyzerTool provides document analysis capabilities for research
type DocumentAnalyzerTool struct {
	httpClient *http.Client
}

// AnalysisResult represents the result of document analysis
type AnalysisResult struct {
	WordCount      int                    `json:"word_count"`
	LineCount      int                    `json:"line_count"`
	CharCount      int                    `json:"char_count"`
	TopKeywords    []KeywordFrequency     `json:"top_keywords"`
	SentenceCount  int                    `json:"sentence_count"`
	AvgWordsPerSentence float64           `json:"avg_words_per_sentence"`
	ReadingTime    int                    `json:"reading_time_minutes"`
	Language       string                 `json:"detected_language"`
	Entities       []string               `json:"entities"`
	Summary        string                 `json:"summary"`
}

// KeywordFrequency represents a keyword and its frequency
type KeywordFrequency struct {
	Word      string `json:"word"`
	Frequency int    `json:"frequency"`
}

// NewDocumentAnalyzerTool creates a new document analyzer tool instance
func NewDocumentAnalyzerTool() *DocumentAnalyzerTool {
	return &DocumentAnalyzerTool{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Definition returns the tool definition for MCP
func (d *DocumentAnalyzerTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "document_analyzer",
		Description: "Analyze documents for research purposes - extract keywords, entities, statistics, and generate summaries from text files or URLs.",
		InputSchema: mcp.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "File path or URL to analyze",
				},
				"source_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"file", "url", "text"},
					"description": "Type of source: file path, URL, or direct text input",
					"default":     "file",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Direct text input (only used when source_type is 'text')",
				},
				"analysis_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"basic", "detailed", "keywords_only", "summary_only"},
					"description": "Type of analysis to perform",
					"default":     "basic",
				},
				"max_keywords": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of keywords to extract (default: 20)",
					"default":     20,
					"minimum":     5,
					"maximum":     100,
				},
			},
			Required: []string{"source_type"},
		},
	}
}

// Execute performs document analysis
func (d *DocumentAnalyzerTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	sourceType, ok := params["source_type"].(string)
	if !ok {
		sourceType = "file"
	}

	analysisType := "basic"
	if at, ok := params["analysis_type"].(string); ok {
		analysisType = at
	}

	maxKeywords := 20
	if mk, ok := params["max_keywords"]; ok {
		if mkFloat, ok := mk.(float64); ok {
			maxKeywords = int(mkFloat)
		}
	}

	var content string
	var err error

	// Get content based on source type
	switch sourceType {
	case "file":
		source, ok := params["source"].(string)
		if !ok || source == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: source parameter is required for file analysis",
				}},
				IsError: true,
			}, nil
		}
		content, err = d.readFile(source)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: fmt.Sprintf("Error reading file: %v", err),
				}},
				IsError: true,
			}, nil
		}

	case "url":
		source, ok := params["source"].(string)
		if !ok || source == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: source parameter is required for URL analysis",
				}},
				IsError: true,
			}, nil
		}
		content, err = d.fetchURL(ctx, source)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: fmt.Sprintf("Error fetching URL: %v", err),
				}},
				IsError: true,
			}, nil
		}

	case "text":
		text, ok := params["text"].(string)
		if !ok || text == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: text parameter is required for direct text analysis",
				}},
				IsError: true,
			}, nil
		}
		content = text

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported source type '%s'", sourceType),
			}},
			IsError: true,
		}, nil
	}

	// Perform analysis
	result := d.analyzeText(content, maxKeywords)

	// Format response based on analysis type
	var responseText string
	switch analysisType {
	case "keywords_only":
		responseText = d.formatKeywordsOnly(result)
	case "summary_only":
		responseText = d.formatSummaryOnly(result)
	case "detailed":
		responseText = d.formatDetailedAnalysis(result)
	default: // basic
		responseText = d.formatBasicAnalysis(result)
	}

	// Also provide JSON format
	jsonResult, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: responseText,
			},
			{
				Type:     "text",
				Text:     string(jsonResult),
				MimeType: "application/json",
			},
		},
		IsError: false,
	}, nil
}

// readFile reads content from a file
func (d *DocumentAnalyzerTool) readFile(filePath string) (string, error) {
	// Check if file exists and is readable
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("file not accessible: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file")
	}

	// Check file size (limit to 10MB)
	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("file too large (max 10MB)")
	}

	// Read file based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt", ".md", ".rst", ".log", "":
		return d.readTextFile(filePath)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

// readTextFile reads a plain text file
func (d *DocumentAnalyzerTool) readTextFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// fetchURL fetches content from a URL
func (d *DocumentAnalyzerTool) fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "MCP-Go-Template/1.0 (Document Analyzer)")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit

	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// analyzeText performs comprehensive text analysis
func (d *DocumentAnalyzerTool) analyzeText(text string, maxKeywords int) *AnalysisResult {
	result := &AnalysisResult{}

	// Basic statistics
	result.CharCount = len(text)
	result.LineCount = strings.Count(text, "\n") + 1
	
	// Word count and keyword extraction
	words := d.extractWords(text)
	result.WordCount = len(words)
	result.TopKeywords = d.extractKeywords(words, maxKeywords)

	// Sentence analysis
	sentences := d.extractSentences(text)
	result.SentenceCount = len(sentences)
	if result.SentenceCount > 0 {
		result.AvgWordsPerSentence = float64(result.WordCount) / float64(result.SentenceCount)
	}

	// Reading time (assuming 200 words per minute)
	result.ReadingTime = (result.WordCount + 199) / 200

	// Simple language detection (very basic)
	result.Language = d.detectLanguage(text)

	// Extract entities (simple implementation)
	result.Entities = d.extractEntities(text)

	// Generate summary
	result.Summary = d.generateSummary(sentences, 3)

	return result
}

// extractWords extracts words from text
func (d *DocumentAnalyzerTool) extractWords(text string) []string {
	// Convert to lowercase and remove punctuation
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	cleaned := reg.ReplaceAllString(text, " ")
	cleaned = strings.ToLower(cleaned)

	// Split into words
	words := strings.Fields(cleaned)

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true, "by": true, "is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "have": true, "has": true, "had": true, "do": true, "does": true, "did": true, "will": true, "would": true, "could": true, "should": true, "may": true, "might": true, "must": true, "can": true, "this": true, "that": true, "these": true, "those": true, "i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true, "me": true, "him": true, "her": true, "us": true, "them": true, "my": true, "your": true, "his": true, "its": true, "our": true, "their": true,
	}

	var filtered []string
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// extractKeywords extracts top keywords by frequency
func (d *DocumentAnalyzerTool) extractKeywords(words []string, maxKeywords int) []KeywordFrequency {
	frequency := make(map[string]int)
	for _, word := range words {
		frequency[word]++
	}

	var keywords []KeywordFrequency
	for word, freq := range frequency {
		keywords = append(keywords, KeywordFrequency{
			Word:      word,
			Frequency: freq,
		})
	}

	// Sort by frequency (descending)
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Frequency > keywords[j].Frequency
	})

	if len(keywords) > maxKeywords {
		keywords = keywords[:maxKeywords]
	}

	return keywords
}

// extractSentences extracts sentences from text
func (d *DocumentAnalyzerTool) extractSentences(text string) []string {
	// Simple sentence splitting on periods, exclamation marks, and question marks
	reg := regexp.MustCompile(`[.!?]+`)
	sentences := reg.Split(text, -1)

	var filtered []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 { // Filter out very short sentences
			filtered = append(filtered, sentence)
		}
	}

	return filtered
}

// detectLanguage performs simple language detection
func (d *DocumentAnalyzerTool) detectLanguage(text string) string {
	// Very simple language detection based on common words
	text = strings.ToLower(text)
	
	englishWords := []string{"the", "and", "that", "have", "for", "not", "with", "you", "this", "but"}
	chineseChars := 0
	englishMatches := 0

	for _, word := range englishWords {
		if strings.Contains(text, " "+word+" ") {
			englishMatches++
		}
	}

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			chineseChars++
		}
	}

	if chineseChars > len(text)/20 {
		return "Chinese"
	} else if englishMatches > 3 {
		return "English"
	}

	return "Unknown"
}

// extractEntities extracts named entities (simple implementation)
func (d *DocumentAnalyzerTool) extractEntities(text string) []string {
	// Simple entity extraction using capitalization patterns
	reg := regexp.MustCompile(`\b[A-Z][a-zA-Z]{2,}\b`)
	matches := reg.FindAllString(text, -1)

	// Deduplicate and filter
	entityMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 3 && !isCommonWord(match) {
			entityMap[match] = true
		}
	}

	var entities []string
	for entity := range entityMap {
		entities = append(entities, entity)
	}

	// Limit to top 20 entities
	if len(entities) > 20 {
		entities = entities[:20]
	}

	return entities
}

// isCommonWord checks if a capitalized word is a common word
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"The": true, "This": true, "That": true, "These": true, "Those": true,
		"And": true, "But": true, "For": true, "Not": true, "With": true,
		"You": true, "Your": true, "They": true, "Their": true, "There": true,
		"When": true, "Where": true, "What": true, "Why": true, "How": true,
		"First": true, "Last": true, "Next": true, "Previous": true, "After": true,
		"Before": true, "During": true, "Since": true, "Until": true, "While": true,
	}
	return commonWords[word]
}

// generateSummary generates a simple extractive summary
func (d *DocumentAnalyzerTool) generateSummary(sentences []string, maxSentences int) string {
	if len(sentences) == 0 {
		return "No content to summarize."
	}

	if len(sentences) <= maxSentences {
		return strings.Join(sentences, ". ") + "."
	}

	// Simple approach: take first, middle, and last sentences
	var selected []string
	selected = append(selected, sentences[0])
	
	if maxSentences > 2 && len(sentences) > 2 {
		middle := len(sentences) / 2
		selected = append(selected, sentences[middle])
	}
	
	if maxSentences > 1 && len(sentences) > 1 {
		selected = append(selected, sentences[len(sentences)-1])
	}

	return strings.Join(selected, ". ") + "."
}

// formatBasicAnalysis formats basic analysis results
func (d *DocumentAnalyzerTool) formatBasicAnalysis(result *AnalysisResult) string {
	var builder strings.Builder
	builder.WriteString("📊 **Document Analysis Results**\n\n")
	builder.WriteString(fmt.Sprintf("**Statistics:**\n"))
	builder.WriteString(fmt.Sprintf("- Words: %d\n", result.WordCount))
	builder.WriteString(fmt.Sprintf("- Lines: %d\n", result.LineCount))
	builder.WriteString(fmt.Sprintf("- Characters: %d\n", result.CharCount))
	builder.WriteString(fmt.Sprintf("- Sentences: %d\n", result.SentenceCount))
	builder.WriteString(fmt.Sprintf("- Average words per sentence: %.1f\n", result.AvgWordsPerSentence))
	builder.WriteString(fmt.Sprintf("- Estimated reading time: %d minutes\n", result.ReadingTime))
	builder.WriteString(fmt.Sprintf("- Detected language: %s\n\n", result.Language))

	builder.WriteString("**Top Keywords:**\n")
	for i, kw := range result.TopKeywords {
		if i >= 10 { // Limit to top 10 for basic analysis
			break
		}
		builder.WriteString(fmt.Sprintf("- %s (%d)\n", kw.Word, kw.Frequency))
	}

	builder.WriteString(fmt.Sprintf("\n**Summary:**\n%s\n", result.Summary))

	return builder.String()
}

// formatDetailedAnalysis formats detailed analysis results
func (d *DocumentAnalyzerTool) formatDetailedAnalysis(result *AnalysisResult) string {
	basic := d.formatBasicAnalysis(result)
	
	var builder strings.Builder
	builder.WriteString(basic)
	
	if len(result.Entities) > 0 {
		builder.WriteString("\n**Extracted Entities:**\n")
		for _, entity := range result.Entities {
			builder.WriteString(fmt.Sprintf("- %s\n", entity))
		}
	}

	if len(result.TopKeywords) > 10 {
		builder.WriteString("\n**Additional Keywords:**\n")
		for i := 10; i < len(result.TopKeywords) && i < 20; i++ {
			kw := result.TopKeywords[i]
			builder.WriteString(fmt.Sprintf("- %s (%d)\n", kw.Word, kw.Frequency))
		}
	}

	return builder.String()
}

// formatKeywordsOnly formats only keywords
func (d *DocumentAnalyzerTool) formatKeywordsOnly(result *AnalysisResult) string {
	var builder strings.Builder
	builder.WriteString("🔍 **Extracted Keywords**\n\n")
	
	for _, kw := range result.TopKeywords {
		builder.WriteString(fmt.Sprintf("- %s (%d occurrences)\n", kw.Word, kw.Frequency))
	}

	return builder.String()
}

// formatSummaryOnly formats only the summary
func (d *DocumentAnalyzerTool) formatSummaryOnly(result *AnalysisResult) string {
	return fmt.Sprintf("📝 **Document Summary**\n\n%s", result.Summary)
}