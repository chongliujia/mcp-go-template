package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// DocumentAnalyzerTool implements document analysis functionality
type DocumentAnalyzerTool struct {
	definition *mcp.Tool
	client     *http.Client
}

// DocumentAnalysis represents the analysis result of a document
type DocumentAnalysis struct {
	Source         string                 `json:"source"`
	Type           string                 `json:"type"`
	WordCount      int                    `json:"word_count"`
	CharCount      int                    `json:"char_count"`
	SentenceCount  int                    `json:"sentence_count"`
	ParagraphCount int                    `json:"paragraph_count"`
	ReadingTime    string                 `json:"reading_time"`
	Language       string                 `json:"language"`
	Keywords       []KeywordInfo          `json:"keywords"`
	Summary        string                 `json:"summary"`
	Entities       []EntityInfo           `json:"entities"`
	Statistics     DocumentStatistics     `json:"statistics"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// KeywordInfo represents keyword frequency information
type KeywordInfo struct {
	Word      string  `json:"word"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
}

// EntityInfo represents detected entities
type EntityInfo struct {
	Text     string `json:"text"`
	Type     string `json:"type"`
	Count    int    `json:"count"`
	Category string `json:"category"`
}

// DocumentStatistics contains statistical information about the document
type DocumentStatistics struct {
	AvgWordsPerSentence   float64            `json:"avg_words_per_sentence"`
	AvgCharsPerWord       float64            `json:"avg_chars_per_word"`
	LexicalDiversity      float64            `json:"lexical_diversity"`
	ComplexityScore       float64            `json:"complexity_score"`
	SentimentScore        float64            `json:"sentiment_score"`
	TopicDistribution     map[string]float64 `json:"topic_distribution"`
	DocumentStructure     DocumentStructure  `json:"document_structure"`
}

// DocumentStructure represents the structure of the document
type DocumentStructure struct {
	HasHeaders    bool     `json:"has_headers"`
	HasLists      bool     `json:"has_lists"`
	HasLinks      bool     `json:"has_links"`
	HeaderLevels  []string `json:"header_levels"`
	ListTypes     []string `json:"list_types"`
	LinkCount     int      `json:"link_count"`
	ImageCount    int      `json:"image_count"`
}

// NewDocumentAnalyzerTool creates a new document analyzer tool
func NewDocumentAnalyzerTool() *DocumentAnalyzerTool {
	return &DocumentAnalyzerTool{
		definition: &mcp.Tool{
			Name:        "document_analyzer",
			Description: "Analyzes documents from files, URLs, or direct text input. Provides comprehensive analysis including keyword extraction, entity recognition, readability metrics, sentiment analysis, and document structure analysis",
			InputSchema: mcp.ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"input_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of input to analyze",
						"enum":        []string{"text", "file", "url"},
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to analyze (text content, file path, or URL)",
					},
					"analysis_depth": map[string]interface{}{
						"type":        "string",
						"description": "Depth of analysis to perform",
						"enum":        []string{"basic", "standard", "comprehensive"},
						"default":     "standard",
					},
					"extract_keywords": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to extract keywords and their frequencies",
						"default":     true,
					},
					"extract_entities": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to extract named entities",
						"default":     true,
					},
					"generate_summary": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to generate a document summary",
						"default":     true,
					},
					"max_keywords": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of keywords to extract",
						"default":     20,
						"minimum":     5,
						"maximum":     100,
					},
				},
				Required: []string{"input_type", "content"},
			},
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Definition returns the tool definition
func (d *DocumentAnalyzerTool) Definition() *mcp.Tool {
	return d.definition
}

// Execute performs document analysis
func (d *DocumentAnalyzerTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract parameters
	inputType, ok := params["input_type"].(string)
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: input_type is required and must be a string",
			}},
			IsError: true,
		}, nil
	}

	content, ok := params["content"].(string)
	if !ok || content == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: content is required and must be a non-empty string",
			}},
			IsError: true,
		}, nil
	}

	analysisDepth := "standard"
	if val, exists := params["analysis_depth"]; exists {
		if depth, ok := val.(string); ok {
			analysisDepth = depth
		}
	}

	extractKeywords := true
	if val, exists := params["extract_keywords"]; exists {
		if extract, ok := val.(bool); ok {
			extractKeywords = extract
		}
	}

	extractEntities := true
	if val, exists := params["extract_entities"]; exists {
		if extract, ok := val.(bool); ok {
			extractEntities = extract
		}
	}

	generateSummary := true
	if val, exists := params["generate_summary"]; exists {
		if generate, ok := val.(bool); ok {
			generateSummary = generate
		}
	}

	maxKeywords := 20
	if val, exists := params["max_keywords"]; exists {
		if max, ok := val.(float64); ok {
			maxKeywords = int(max)
		}
	}

	// Get document text
	text, source, err := d.getDocumentText(inputType, content)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving document: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Perform analysis
	analysis := d.analyzeDocument(text, source, inputType, analysisDepth, extractKeywords, extractEntities, generateSummary, maxKeywords)
	
	duration := time.Since(startTime)
	analysis.Metadata["analysis_duration"] = duration.String()
	analysis.Metadata["analysis_time"] = time.Now().Format(time.RFC3339)

	// Format results
	resultText := d.formatAnalysisResults(analysis)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		jsonData = []byte(fmt.Sprintf(`{"error": "failed to marshal analysis: %v"}`, err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: resultText,
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

// getDocumentText retrieves text content based on input type with improved error handling
func (d *DocumentAnalyzerTool) getDocumentText(inputType, content string) (string, string, error) {
	switch inputType {
	case "text":
		// Validate text content
		if len(strings.TrimSpace(content)) == 0 {
			return "", "", fmt.Errorf("text content cannot be empty")
		}
		return content, "Direct Text Input", nil
		
	case "file":
		// Validate file path
		if content == "" {
			return "", "", fmt.Errorf("file path cannot be empty")
		}
		
		// Check if file exists
		if _, err := os.Stat(content); os.IsNotExist(err) {
			return "", "", fmt.Errorf("file does not exist: %s", content)
		}
		
		// Check file size (limit to 10MB)
		fileInfo, err := os.Stat(content)
		if err != nil {
			return "", "", fmt.Errorf("failed to get file info for %s: %w", content, err)
		}
		if fileInfo.Size() > 10*1024*1024 { // 10MB limit
			return "", "", fmt.Errorf("file too large (max 10MB): %s is %.1f MB", content, float64(fileInfo.Size())/(1024*1024))
		}
		
		data, err := os.ReadFile(content)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file %s: %w", content, err)
		}
		return string(data), content, nil
		
	case "url":
		// Validate URL
		if content == "" {
			return "", "", fmt.Errorf("URL cannot be empty")
		}
		
		parsedURL, err := url.Parse(content)
		if err != nil {
			return "", "", fmt.Errorf("invalid URL format: %w", err)
		}
		
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return "", "", fmt.Errorf("unsupported URL scheme: %s (only http/https supported)", parsedURL.Scheme)
		}
		
		req, err := http.NewRequest("GET", content, nil)
		if err != nil {
			return "", "", fmt.Errorf("failed to create request for URL %s: %w", content, err)
		}
		
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MCP-Document-Analyzer/1.0)")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "identity") // Disable compression for simplicity
		
		resp, err := d.client.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("failed to fetch URL %s: %w", content, err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", "", fmt.Errorf("HTTP error %d %s for URL %s", resp.StatusCode, resp.Status, content)
		}
		
		// Check content type
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/") && !strings.Contains(contentType, "application/") {
			return "", "", fmt.Errorf("unsupported content type: %s", contentType)
		}
		
		// Limit response size (10MB)
		limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
		body, err := io.ReadAll(limitedReader)
		if err != nil {
			return "", "", fmt.Errorf("failed to read response body for URL %s: %w", content, err)
		}
		
		// Enhanced HTML stripping with content type detection
		var text string
		if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml") {
			text = d.stripHTML(string(body))
		} else {
			// For plain text or other formats, use as-is
			text = string(body)
		}
		
		// Validate that we got some meaningful content
		if len(strings.TrimSpace(text)) == 0 {
			return "", "", fmt.Errorf("no text content extracted from URL %s", content)
		}
		
		return text, content, nil
		
	default:
		return "", "", fmt.Errorf("unsupported input type: %s (supported: text, file, url)", inputType)
	}
}

// analyzeDocument performs comprehensive document analysis
func (d *DocumentAnalyzerTool) analyzeDocument(text, source, inputType, analysisDepth string, extractKeywords, extractEntities, generateSummary bool, maxKeywords int) *DocumentAnalysis {
	analysis := &DocumentAnalysis{
		Source:   source,
		Type:     inputType,
		Metadata: make(map[string]interface{}),
	}

	// Basic statistics
	analysis.CharCount = len(text)
	analysis.WordCount = d.countWords(text)
	analysis.SentenceCount = d.countSentences(text)
	analysis.ParagraphCount = d.countParagraphs(text)
	analysis.ReadingTime = d.calculateReadingTime(analysis.WordCount)
	analysis.Language = d.detectLanguage(text)

	// Document structure analysis
	analysis.Statistics.DocumentStructure = d.analyzeDocumentStructure(text)

	// Basic statistics calculations
	if analysis.WordCount > 0 {
		analysis.Statistics.AvgWordsPerSentence = float64(analysis.WordCount) / float64(analysis.SentenceCount)
		analysis.Statistics.AvgCharsPerWord = float64(analysis.CharCount) / float64(analysis.WordCount)
	}

	// Keyword extraction
	if extractKeywords {
		analysis.Keywords = d.extractKeywords(text, maxKeywords)
		analysis.Statistics.LexicalDiversity = d.calculateLexicalDiversity(text)
	}

	// Entity extraction
	if extractEntities {
		analysis.Entities = d.extractEntities(text)
	}

	// Summary generation
	if generateSummary {
		analysis.Summary = d.generateSummary(text)
	}

	// Advanced analysis based on depth
	switch analysisDepth {
	case "comprehensive":
		analysis.Statistics.ComplexityScore = d.calculateComplexityScore(text)
		analysis.Statistics.SentimentScore = d.calculateSentimentScore(text)
		analysis.Statistics.TopicDistribution = d.analyzeTopicDistribution(text)
		fallthrough
	case "standard":
		// Standard analysis includes all basic metrics
		if analysis.Statistics.ComplexityScore == 0 {
			analysis.Statistics.ComplexityScore = d.calculateBasicComplexity(text)
		}
	case "basic":
		// Basic analysis only includes fundamental metrics (already calculated above)
	}

	return analysis
}

// countWords counts words in the text
func (d *DocumentAnalyzerTool) countWords(text string) int {
	words := strings.Fields(strings.ToLower(text))
	return len(words)
}

// countSentences counts sentences in the text
func (d *DocumentAnalyzerTool) countSentences(text string) int {
	// Simple sentence detection based on punctuation
	re := regexp.MustCompile(`[.!?]+`)
	sentences := re.Split(text, -1)
	count := 0
	for _, sentence := range sentences {
		if strings.TrimSpace(sentence) != "" {
			count++
		}
	}
	if count == 0 {
		count = 1 // Minimum one sentence
	}
	return count
}

// countParagraphs counts paragraphs in the text
func (d *DocumentAnalyzerTool) countParagraphs(text string) int {
	paragraphs := strings.Split(text, "\n\n")
	count := 0
	for _, para := range paragraphs {
		if strings.TrimSpace(para) != "" {
			count++
		}
	}
	if count == 0 {
		count = 1 // Minimum one paragraph
	}
	return count
}

// calculateReadingTime estimates reading time based on average reading speed
func (d *DocumentAnalyzerTool) calculateReadingTime(wordCount int) string {
	wordsPerMinute := 200.0 // Average reading speed
	minutes := float64(wordCount) / wordsPerMinute
	
	if minutes < 1 {
		seconds := minutes * 60
		return fmt.Sprintf("%.0f seconds", seconds)
	}
	
	return fmt.Sprintf("%.1f minutes", minutes)
}

// detectLanguage performs simple language detection
func (d *DocumentAnalyzerTool) detectLanguage(text string) string {
	// Simple heuristic language detection
	text = strings.ToLower(text)
	
	// English indicators
	englishWords := []string{"the", "and", "is", "in", "to", "of", "a", "that", "it", "with", "for", "as", "was", "on", "are", "you"}
	englishCount := 0
	
	for _, word := range englishWords {
		if strings.Contains(text, " "+word+" ") || strings.HasPrefix(text, word+" ") || strings.HasSuffix(text, " "+word) {
			englishCount++
		}
	}
	
	if englishCount >= 3 {
		return "English"
	}
	
	return "Unknown"
}

// extractKeywords extracts keywords and their frequencies
func (d *DocumentAnalyzerTool) extractKeywords(text string, maxKeywords int) []KeywordInfo {
	// Clean and tokenize text
	words := d.tokenizeText(text)
	
	// Count word frequencies
	wordFreq := make(map[string]int)
	for _, word := range words {
		if len(word) >= 3 && !d.isStopWord(word) { // Filter short words and stop words
			wordFreq[word]++
		}
	}
	
	// Convert to KeywordInfo and sort by frequency
	var keywords []KeywordInfo
	totalWords := len(words)
	
	for word, freq := range wordFreq {
		score := float64(freq) / float64(totalWords) // Simple TF score
		keywords = append(keywords, KeywordInfo{
			Word:      word,
			Frequency: freq,
			Score:     score,
		})
	}
	
	// Sort by frequency (descending)
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Frequency > keywords[j].Frequency
	})
	
	// Limit to maxKeywords
	if len(keywords) > maxKeywords {
		keywords = keywords[:maxKeywords]
	}
	
	return keywords
}

// tokenizeText tokenizes text into words
func (d *DocumentAnalyzerTool) tokenizeText(text string) []string {
	// Convert to lowercase and remove punctuation
	re := regexp.MustCompile(`[^\p{L}\s]+`)
	cleaned := re.ReplaceAllString(strings.ToLower(text), " ")
	return strings.Fields(cleaned)
}

// isStopWord checks if a word is a stop word
func (d *DocumentAnalyzerTool) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
		"for": true, "from": true, "has": true, "he": true, "in": true, "is": true, "it": true, "its": true,
		"of": true, "on": true, "that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
		"this": true, "they": true, "have": true, "had": true, "what": true, "said": true, "each": true,
		"which": true, "do": true, "how": true, "their": true, "if": true, "up": true, "out": true, "many": true,
		"then": true, "them": true, "these": true, "so": true, "some": true, "her": true, "would": true,
		"make": true, "like": true, "into": true, "him": true, "time": true, "two": true, "more": true,
		"go": true, "no": true, "way": true, "could": true, "my": true, "than": true, "first": true, "been": true,
		"call": true, "who": true, "oil": true, "sit": true, "now": true, "find": true, "down": true, "day": true,
		"did": true, "get": true, "come": true, "made": true, "may": true, "part": true,
	}
	return stopWords[word]
}

// extractEntities performs simple named entity recognition
func (d *DocumentAnalyzerTool) extractEntities(text string) []EntityInfo {
	var entities []EntityInfo
	entityCounts := make(map[string]map[string]int)
	
	// Initialize entity count maps
	entityCounts["PERSON"] = make(map[string]int)
	entityCounts["LOCATION"] = make(map[string]int)
	entityCounts["ORGANIZATION"] = make(map[string]int)
	entityCounts["DATE"] = make(map[string]int)
	entityCounts["MONEY"] = make(map[string]int)
	
	// Simple regex patterns for entity recognition
	patterns := map[string]*regexp.Regexp{
		"PERSON":       regexp.MustCompile(`\b[A-Z][a-z]+ [A-Z][a-z]+\b`), // Simple name pattern
		"LOCATION":     regexp.MustCompile(`\b(?:New York|London|Paris|Tokyo|Beijing|Los Angeles|Chicago|San Francisco)\b`),
		"ORGANIZATION": regexp.MustCompile(`\b(?:Google|Microsoft|Apple|Amazon|Facebook|IBM|Oracle|Intel)\b`),
		"DATE":         regexp.MustCompile(`\b\d{1,2}/\d{1,2}/\d{4}\b|\b\d{4}-\d{2}-\d{2}\b`),
		"MONEY":        regexp.MustCompile(`\$\d+(?:,\d{3})*(?:\.\d{2})?\b`),
	}
	
	for entityType, pattern := range patterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			entityCounts[entityType][match]++
		}
	}
	
	// Convert to EntityInfo
	for entityType, counts := range entityCounts {
		for text, count := range counts {
			entities = append(entities, EntityInfo{
				Text:     text,
				Type:     entityType,
				Count:    count,
				Category: d.getCategoryForEntityType(entityType),
			})
		}
	}
	
	// Sort by count (descending)
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Count > entities[j].Count
	})
	
	return entities
}

// getCategoryForEntityType returns the category for an entity type
func (d *DocumentAnalyzerTool) getCategoryForEntityType(entityType string) string {
	categories := map[string]string{
		"PERSON":       "People",
		"LOCATION":     "Places",
		"ORGANIZATION": "Organizations",
		"DATE":         "Temporal",
		"MONEY":        "Financial",
	}
	if category, exists := categories[entityType]; exists {
		return category
	}
	return "Other"
}

// generateSummary generates a simple extractive summary
func (d *DocumentAnalyzerTool) generateSummary(text string) string {
	sentences := d.splitIntoSentences(text)
	if len(sentences) <= 2 {
		return strings.Join(sentences, " ")
	}
	
	// Simple extractive summarization: take first and most keyword-rich sentences
	keywords := d.extractKeywords(text, 10)
	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		keywordSet[kw.Word] = true
	}
	
	type sentenceScore struct {
		sentence string
		score    int
		position int
	}
	
	var scores []sentenceScore
	for i, sentence := range sentences {
		score := 0
		words := d.tokenizeText(sentence)
		for _, word := range words {
			if keywordSet[word] {
				score++
			}
		}
		// Bonus for position (earlier sentences get higher scores)
		if i < len(sentences)/3 {
			score += 2
		}
		scores = append(scores, sentenceScore{sentence, score, i})
	}
	
	// Sort by score (descending)
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].position < scores[j].position // Prefer earlier sentences for ties
		}
		return scores[i].score > scores[j].score
	})
	
	// Take top 2-3 sentences for summary
	summaryCount := int(math.Min(3, float64(len(sentences)/3)))
	if summaryCount < 1 {
		summaryCount = 1
	}
	
	var summarySentences []string
	for i := 0; i < summaryCount && i < len(scores); i++ {
		summarySentences = append(summarySentences, strings.TrimSpace(scores[i].sentence))
	}
	
	return strings.Join(summarySentences, " ")
}

// splitIntoSentences splits text into sentences
func (d *DocumentAnalyzerTool) splitIntoSentences(text string) []string {
	re := regexp.MustCompile(`[.!?]+\s+`)
	sentences := re.Split(text, -1)
	var result []string
	for _, sentence := range sentences {
		if trimmed := strings.TrimSpace(sentence); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// calculateLexicalDiversity calculates the lexical diversity of the text
func (d *DocumentAnalyzerTool) calculateLexicalDiversity(text string) float64 {
	words := d.tokenizeText(text)
	if len(words) == 0 {
		return 0
	}
	
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		if len(word) >= 3 && !d.isStopWord(word) {
			uniqueWords[word] = true
		}
	}
	
	return float64(len(uniqueWords)) / float64(len(words))
}

// calculateBasicComplexity calculates a basic complexity score
func (d *DocumentAnalyzerTool) calculateBasicComplexity(text string) float64 {
	words := d.tokenizeText(text)
	sentences := d.splitIntoSentences(text)
	
	if len(sentences) == 0 {
		return 0
	}
	
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	
	// Simple complexity based on sentence length
	if avgWordsPerSentence <= 10 {
		return 0.2 // Very easy
	} else if avgWordsPerSentence <= 15 {
		return 0.4 // Easy
	} else if avgWordsPerSentence <= 20 {
		return 0.6 // Medium
	} else if avgWordsPerSentence <= 25 {
		return 0.8 // Hard
	}
	return 1.0 // Very hard
}

// calculateComplexityScore calculates a comprehensive complexity score
func (d *DocumentAnalyzerTool) calculateComplexityScore(text string) float64 {
	// This is a simplified complexity calculation
	// In a real implementation, you might use more sophisticated metrics like Flesch-Kincaid
	return d.calculateBasicComplexity(text)
}

// calculateSentimentScore performs basic sentiment analysis
func (d *DocumentAnalyzerTool) calculateSentimentScore(text string) float64 {
	// Simple sentiment analysis based on positive/negative word counts
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "positive", "happy", "love", "best"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "negative", "sad", "hate", "worst", "difficult", "problem"}
	
	text = strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}
	
	total := positiveCount + negativeCount
	if total == 0 {
		return 0.0 // Neutral
	}
	
	// Return score between -1 (very negative) and 1 (very positive)
	return (float64(positiveCount) - float64(negativeCount)) / float64(total)
}

// analyzeTopicDistribution performs simple topic analysis
func (d *DocumentAnalyzerTool) analyzeTopicDistribution(text string) map[string]float64 {
	topics := map[string][]string{
		"Technology": {"computer", "software", "technology", "digital", "internet", "data", "system", "application"},
		"Business":   {"business", "company", "market", "financial", "economy", "revenue", "profit", "customer"},
		"Science":    {"research", "study", "analysis", "experiment", "scientific", "method", "theory", "hypothesis"},
		"Health":     {"health", "medical", "treatment", "patient", "disease", "medicine", "hospital", "doctor"},
		"Education":  {"education", "learning", "student", "teacher", "school", "university", "knowledge", "study"},
	}
	
	text = strings.ToLower(text)
	topicScores := make(map[string]float64)
	
	totalWords := len(d.tokenizeText(text))
	if totalWords == 0 {
		return topicScores
	}
	
	for topic, keywords := range topics {
		count := 0
		for _, keyword := range keywords {
			count += strings.Count(text, keyword)
		}
		topicScores[topic] = float64(count) / float64(totalWords)
	}
	
	return topicScores
}

// analyzeDocumentStructure analyzes the structure of the document
func (d *DocumentAnalyzerTool) analyzeDocumentStructure(text string) DocumentStructure {
	structure := DocumentStructure{
		HeaderLevels: []string{},
		ListTypes:    []string{},
	}
	
	// Check for headers (simple markdown-style detection)
	headerRegex := regexp.MustCompile(`(?m)^#+\s+`)
	if headerRegex.MatchString(text) {
		structure.HasHeaders = true
		headers := headerRegex.FindAllString(text, -1)
		for _, header := range headers {
			level := fmt.Sprintf("H%d", strings.Count(header, "#"))
			structure.HeaderLevels = append(structure.HeaderLevels, level)
		}
	}
	
	// Check for lists
	listRegex := regexp.MustCompile(`(?m)^[\s]*[*\-+]\s+|^[\s]*\d+\.\s+`)
	if listRegex.MatchString(text) {
		structure.HasLists = true
		structure.ListTypes = append(structure.ListTypes, "unordered", "ordered")
	}
	
	// Check for links
	linkRegex := regexp.MustCompile(`https?://[^\s]+|\[([^\]]+)\]\([^)]+\)`)
	links := linkRegex.FindAllString(text, -1)
	structure.LinkCount = len(links)
	structure.HasLinks = structure.LinkCount > 0
	
	// Check for images (simple detection)
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)|<img[^>]+>`)
	images := imageRegex.FindAllString(text, -1)
	structure.ImageCount = len(images)
	
	return structure
}

// stripHTML removes HTML tags from text using proper HTML parsing
func (d *DocumentAnalyzerTool) stripHTML(htmlContent string) string {
	// Try proper HTML parsing first
	if cleanText := d.parseHTMLWithParser(htmlContent); cleanText != "" {
		return cleanText
	}
	
	// Fallback to regex-based stripping
	return d.stripHTMLRegex(htmlContent)
}

// parseHTMLWithParser uses golang.org/x/net/html to properly parse HTML
func (d *DocumentAnalyzerTool) parseHTMLWithParser(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// If parsing fails, return empty string to trigger fallback
		return ""
	}
	
	var textBuilder strings.Builder
	d.extractTextFromNode(doc, &textBuilder)
	
	// Clean up whitespace
	text := textBuilder.String()
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// extractTextFromNode recursively extracts text from HTML nodes
func (d *DocumentAnalyzerTool) extractTextFromNode(n *html.Node, textBuilder *strings.Builder) {
	if n.Type == html.TextNode {
		// Skip script and style content
		if n.Parent != nil {
			tagName := strings.ToLower(n.Parent.Data)
			if tagName == "script" || tagName == "style" || tagName == "noscript" {
				return
			}
		}
		textBuilder.WriteString(n.Data)
		textBuilder.WriteString(" ")
	} else if n.Type == html.ElementNode {
		// Add line breaks for block elements
		tagName := strings.ToLower(n.Data)
		if d.isBlockElement(tagName) {
			textBuilder.WriteString("\n")
		}
	}
	
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		d.extractTextFromNode(c, textBuilder)
	}
	
	// Add line breaks after block elements
	if n.Type == html.ElementNode {
		tagName := strings.ToLower(n.Data)
		if d.isBlockElement(tagName) {
			textBuilder.WriteString("\n")
		}
	}
}

// isBlockElement checks if an HTML element is a block-level element
func (d *DocumentAnalyzerTool) isBlockElement(tagName string) bool {
	blockElements := map[string]bool{
		"div": true, "p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"article": true, "section": true, "nav": true, "aside": true, "header": true, "footer": true,
		"main": true, "ul": true, "ol": true, "li": true, "blockquote": true, "pre": true,
		"table": true, "tr": true, "td": true, "th": true, "form": true, "fieldset": true,
		"address": true, "figure": true, "figcaption": true, "hr": true, "br": true,
	}
	return blockElements[tagName]
}

// stripHTMLRegex removes HTML tags using regex (fallback method)
func (d *DocumentAnalyzerTool) stripHTMLRegex(html string) string {
	// Remove script and style tags with their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = scriptRegex.ReplaceAllString(html, "")
	html = styleRegex.ReplaceAllString(html, "")
	
	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	html = commentRegex.ReplaceAllString(html, "")
	
	// Remove all HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	html = tagRegex.ReplaceAllString(html, " ")
	
	// Clean up whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	html = whitespaceRegex.ReplaceAllString(html, " ")
	
	return strings.TrimSpace(html)
}

// formatAnalysisResults formats the analysis results for display
func (d *DocumentAnalyzerTool) formatAnalysisResults(analysis *DocumentAnalysis) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("📄 Document Analysis Results\n"))
	result.WriteString(fmt.Sprintf("═══════════════════════════════════════════════════\n\n"))
	
	result.WriteString(fmt.Sprintf("📊 Basic Statistics:\n"))
	result.WriteString(fmt.Sprintf("  Source: %s\n", analysis.Source))
	result.WriteString(fmt.Sprintf("  Type: %s\n", analysis.Type))
	result.WriteString(fmt.Sprintf("  Language: %s\n", analysis.Language))
	result.WriteString(fmt.Sprintf("  Word Count: %d\n", analysis.WordCount))
	result.WriteString(fmt.Sprintf("  Character Count: %d\n", analysis.CharCount))
	result.WriteString(fmt.Sprintf("  Sentences: %d\n", analysis.SentenceCount))
	result.WriteString(fmt.Sprintf("  Paragraphs: %d\n", analysis.ParagraphCount))
	result.WriteString(fmt.Sprintf("  Reading Time: %s\n\n", analysis.ReadingTime))
	
	result.WriteString(fmt.Sprintf("📈 Advanced Metrics:\n"))
	result.WriteString(fmt.Sprintf("  Avg Words/Sentence: %.1f\n", analysis.Statistics.AvgWordsPerSentence))
	result.WriteString(fmt.Sprintf("  Avg Chars/Word: %.1f\n", analysis.Statistics.AvgCharsPerWord))
	result.WriteString(fmt.Sprintf("  Lexical Diversity: %.3f\n", analysis.Statistics.LexicalDiversity))
	result.WriteString(fmt.Sprintf("  Complexity Score: %.2f\n", analysis.Statistics.ComplexityScore))
	result.WriteString(fmt.Sprintf("  Sentiment Score: %.2f\n\n", analysis.Statistics.SentimentScore))
	
	if len(analysis.Keywords) > 0 {
		result.WriteString(fmt.Sprintf("🔑 Top Keywords:\n"))
		for i, kw := range analysis.Keywords {
			if i >= 10 { // Limit display to top 10
				break
			}
			result.WriteString(fmt.Sprintf("  %d. %s (freq: %d, score: %.4f)\n", i+1, kw.Word, kw.Frequency, kw.Score))
		}
		result.WriteString("\n")
	}
	
	if len(analysis.Entities) > 0 {
		result.WriteString(fmt.Sprintf("🏷️  Named Entities:\n"))
		for _, entity := range analysis.Entities {
			result.WriteString(fmt.Sprintf("  %s (%s): %d occurrences\n", entity.Text, entity.Type, entity.Count))
		}
		result.WriteString("\n")
	}
	
	if analysis.Summary != "" {
		result.WriteString(fmt.Sprintf("📝 Summary:\n"))
		result.WriteString(fmt.Sprintf("  %s\n\n", analysis.Summary))
	}
	
	// Document structure
	structure := analysis.Statistics.DocumentStructure
	result.WriteString(fmt.Sprintf("🏗️  Document Structure:\n"))
	result.WriteString(fmt.Sprintf("  Headers: %t", structure.HasHeaders))
	if structure.HasHeaders {
		result.WriteString(fmt.Sprintf(" (levels: %s)", strings.Join(structure.HeaderLevels, ", ")))
	}
	result.WriteString("\n")
	result.WriteString(fmt.Sprintf("  Lists: %t\n", structure.HasLists))
	result.WriteString(fmt.Sprintf("  Links: %d\n", structure.LinkCount))
	result.WriteString(fmt.Sprintf("  Images: %d\n\n", structure.ImageCount))
	
	if len(analysis.Statistics.TopicDistribution) > 0 {
		result.WriteString(fmt.Sprintf("📊 Topic Distribution:\n"))
		for topic, score := range analysis.Statistics.TopicDistribution {
			if score > 0 {
				result.WriteString(fmt.Sprintf("  %s: %.3f\n", topic, score))
			}
		}
		result.WriteString("\n")
	}
	
	return result.String()
}