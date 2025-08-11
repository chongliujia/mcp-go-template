package examples

import (
	"context"
	"strings"
	"testing"
)

func TestDocumentAnalyzerTool_Definition(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	def := analyzer.Definition()
	
	if def == nil {
		t.Fatal("Definition should not be nil")
	}
	
	if def.Name != "document_analyzer" {
		t.Errorf("Expected name 'document_analyzer', got '%s'", def.Name)
	}
	
	if def.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestDocumentAnalyzerTool_Execute_Text(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	testText := "This is a test document. It contains multiple sentences. This helps test the document analyzer functionality. We can measure word count, sentence count, and other metrics."
	
	params := map[string]interface{}{
		"input_type": "text",
		"content":    testText,
	}
	
	result, err := analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful analysis, got error: %v", result.Content[0].Text)
	}
	
	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items (text and JSON)")
	}
}

func TestDocumentAnalyzerTool_Execute_EmptyText(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"input_type": "text",
		"content":    "",
	}
	
	result, err := analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for empty text content")
	}
}

func TestDocumentAnalyzerTool_Execute_InvalidInputType(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"input_type": "invalid",
		"content":    "test content",
	}
	
	result, err := analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for invalid input type")
	}
}

func TestDocumentAnalyzerTool_Execute_MissingParameters(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	// Missing input_type
	params := map[string]interface{}{
		"content": "test content",
	}
	
	result, err := analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for missing input_type")
	}
	
	// Missing content
	params = map[string]interface{}{
		"input_type": "text",
	}
	
	result, err = analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for missing content")
	}
}

func TestDocumentAnalyzerTool_Execute_WithOptions(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	testText := "This is a test document for analyzing. It contains multiple sentences and should extract keywords."
	
	params := map[string]interface{}{
		"input_type":        "text",
		"content":           testText,
		"analysis_depth":    "comprehensive",
		"extract_keywords":  true,
		"extract_entities":  true,
		"generate_summary":  true,
		"max_keywords":      10,
	}
	
	result, err := analyzer.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful analysis, got error: %v", result.Content[0].Text)
	}
}

func TestDocumentAnalyzerTool_CountWords(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"This is a test sentence.", 5},
		{"  multiple   spaces   between   words  ", 4},
	}
	
	for _, test := range tests {
		result := analyzer.countWords(test.input)
		if result != test.expected {
			t.Errorf("countWords(%q): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestDocumentAnalyzerTool_CountSentences(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected int
	}{
		{"", 1}, // Minimum one sentence
		{"Hello world", 1},
		{"First sentence. Second sentence.", 2},
		{"Question? Answer! Exclamation.", 3},
		{"Multiple punctuation... Works well!", 2},
	}
	
	for _, test := range tests {
		result := analyzer.countSentences(test.input)
		if result != test.expected {
			t.Errorf("countSentences(%q): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestDocumentAnalyzerTool_CountParagraphs(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected int
	}{
		{"", 1}, // Minimum one paragraph
		{"Single paragraph", 1},
		{"First paragraph\n\nSecond paragraph", 2},
		{"One\n\nTwo\n\nThree", 3},
	}
	
	for _, test := range tests {
		result := analyzer.countParagraphs(test.input)
		if result != test.expected {
			t.Errorf("countParagraphs(%q): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestDocumentAnalyzerTool_CalculateReadingTime(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		wordCount int
		contains  string
	}{
		{0, "0 seconds"},
		{50, "seconds"},
		{200, "1.0 minutes"},
		{400, "2.0 minutes"},
		{1000, "5.0 minutes"},
	}
	
	for _, test := range tests {
		result := analyzer.calculateReadingTime(test.wordCount)
		if !strings.Contains(result, test.contains) {
			t.Errorf("calculateReadingTime(%d): expected to contain %q, got %q", test.wordCount, test.contains, result)
		}
	}
}

func TestDocumentAnalyzerTool_DetectLanguage(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"The quick brown fox jumps over the lazy dog", "English"},
		{"This is a test with common English words", "English"},
		{"Random text without common words", "Unknown"},
		{"", "Unknown"},
	}
	
	for _, test := range tests {
		result := analyzer.detectLanguage(test.input)
		if result != test.expected {
			t.Errorf("detectLanguage(%q): expected %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestDocumentAnalyzerTool_TokenizeText(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"Hello world!", []string{"hello", "world"}},
		{"Test, with punctuation.", []string{"test", "with", "punctuation"}},
		{"Numbers123 and symbols@#$", []string{"numbers", "and", "symbols"}},
	}
	
	for _, test := range tests {
		result := analyzer.tokenizeText(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("tokenizeText(%q): expected length %d, got length %d", test.input, len(test.expected), len(result))
			continue
		}
		
		for i, word := range result {
			if word != test.expected[i] {
				t.Errorf("tokenizeText(%q): expected word %d to be %q, got %q", test.input, i, test.expected[i], word)
			}
		}
	}
}

func TestDocumentAnalyzerTool_IsStopWord(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	stopWords := []string{"the", "and", "is", "in", "to", "of", "a"}
	nonStopWords := []string{"document", "analysis", "important", "keyword"}
	
	for _, word := range stopWords {
		if !analyzer.isStopWord(word) {
			t.Errorf("Expected %q to be a stop word", word)
		}
	}
	
	for _, word := range nonStopWords {
		if analyzer.isStopWord(word) {
			t.Errorf("Expected %q to not be a stop word", word)
		}
	}
}

func TestDocumentAnalyzerTool_ExtractKeywords(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	text := "Document analysis is important. Document processing and analysis help understand content. Analysis provides insights."
	keywords := analyzer.extractKeywords(text, 5)
	
	if len(keywords) == 0 {
		t.Error("Expected at least some keywords")
	}
	
	// Check that keywords are sorted by frequency
	for i := 1; i < len(keywords); i++ {
		if keywords[i-1].Frequency < keywords[i].Frequency {
			t.Error("Keywords should be sorted by frequency in descending order")
			break
		}
	}
	
	// Check that no stop words are included
	for _, keyword := range keywords {
		if analyzer.isStopWord(keyword.Word) {
			t.Errorf("Stop word %q should not be included in keywords", keyword.Word)
		}
	}
}

func TestDocumentAnalyzerTool_CalculateLexicalDiversity(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input     string
		minScore  float64
		maxScore  float64
	}{
		{"", 0, 0},
		{"same same same same", 0, 0.5}, // Low diversity
		{"unique different words every time", 0.5, 1.0}, // High diversity
	}
	
	for _, test := range tests {
		score := analyzer.calculateLexicalDiversity(test.input)
		if score < test.minScore || score > test.maxScore {
			t.Errorf("calculateLexicalDiversity(%q): expected score between %f and %f, got %f", test.input, test.minScore, test.maxScore, score)
		}
	}
}

func TestDocumentAnalyzerTool_CalculateBasicComplexity(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input       string
		description string
	}{
		{"Short. Easy. Text.", "Simple text should have low complexity"},
		{"This is a moderately complex sentence with several words and some complexity.", "Medium text should have medium complexity"},
		{"This is an extremely long and complex sentence that contains many words, multiple clauses, extensive descriptions, detailed explanations, and various linguistic elements that significantly increase the overall complexity and reading difficulty of the text.", "Complex text should have high complexity"},
	}
	
	for _, test := range tests {
		score := analyzer.calculateBasicComplexity(test.input)
		if score < 0 || score > 1 {
			t.Errorf("calculateBasicComplexity: %s - score %f should be between 0 and 1", test.description, score)
		}
	}
}

func TestDocumentAnalyzerTool_CalculateSentimentScore(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input       string
		minScore    float64
		maxScore    float64
		description string
	}{
		{"This is great and amazing and wonderful!", 0.1, 1.0, "Positive text"},
		{"This is bad and terrible and awful!", -1.0, -0.1, "Negative text"},
		{"This is neutral text without sentiment words.", -0.1, 0.1, "Neutral text"},
	}
	
	for _, test := range tests {
		score := analyzer.calculateSentimentScore(test.input)
		if score < test.minScore || score > test.maxScore {
			t.Errorf("calculateSentimentScore: %s - expected score between %f and %f, got %f", test.description, test.minScore, test.maxScore, score)
		}
	}
}

func TestDocumentAnalyzerTool_StripHTMLRegex(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello world</p>", "Hello world"},
		{"<div><span>Test</span> content</div>", "Test content"},
		{"<script>alert('test')</script>Hello", "Hello"},
		{"<style>body { color: red; }</style>Content", "Content"},
		{"<!-- comment -->Text", "Text"},
		{"Plain text", "Plain text"},
	}
	
	for _, test := range tests {
		result := strings.TrimSpace(analyzer.stripHTMLRegex(test.input))
		expected := strings.TrimSpace(test.expected)
		if result != expected {
			t.Errorf("stripHTMLRegex(%q): expected %q, got %q", test.input, expected, result)
		}
	}
}

func TestDocumentAnalyzerTool_IsBlockElement(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	blockElements := []string{"div", "p", "h1", "h2", "h3", "ul", "ol", "li", "table", "tr", "td"}
	inlineElements := []string{"span", "a", "em", "strong", "img", "input", "b", "i"}
	
	for _, element := range blockElements {
		if !analyzer.isBlockElement(element) {
			t.Errorf("Expected %q to be a block element", element)
		}
	}
	
	for _, element := range inlineElements {
		if analyzer.isBlockElement(element) {
			t.Errorf("Expected %q to not be a block element", element)
		}
	}
}

func TestDocumentAnalyzerTool_AnalyzeDocumentStructure(t *testing.T) {
	analyzer := NewDocumentAnalyzerTool()
	
	text := "# Header 1\n## Header 2\n- List item\n1. Numbered item\nCheck out https://example.com\n![Image](image.jpg)"
	structure := analyzer.analyzeDocumentStructure(text)
	
	if !structure.HasHeaders {
		t.Error("Expected to detect headers")
	}
	
	if !structure.HasLists {
		t.Error("Expected to detect lists")
	}
	
	if !structure.HasLinks {
		t.Error("Expected to detect links")
	}
	
	if structure.LinkCount == 0 {
		t.Error("Expected link count to be greater than 0")
	}
	
	if structure.ImageCount == 0 {
		t.Error("Expected image count to be greater than 0")
	}
}

// Benchmark tests
func BenchmarkDocumentAnalyzerTool_Execute(b *testing.B) {
	analyzer := NewDocumentAnalyzerTool()
	ctx := context.Background()
	
	testText := strings.Repeat("This is a benchmark test for document analysis. It contains multiple sentences and should provide consistent performance metrics. ", 100)
	
	params := map[string]interface{}{
		"input_type": "text",
		"content":    testText,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkDocumentAnalyzerTool_ExtractKeywords(b *testing.B) {
	analyzer := NewDocumentAnalyzerTool()
	
	testText := strings.Repeat("document analysis keyword extraction performance benchmark test example content processing natural language ", 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.extractKeywords(testText, 20)
	}
}