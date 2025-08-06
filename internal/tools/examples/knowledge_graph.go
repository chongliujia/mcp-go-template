package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/chongliujia/mcp-go-template/pkg/mcp"
)

// KnowledgeGraphTool builds and analyzes knowledge graphs from text
type KnowledgeGraphTool struct{}

// Entity represents a knowledge graph entity
type Entity struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Attributes  map[string]string `json:"attributes"`
	Mentions    int               `json:"mentions"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Weight int    `json:"weight"`
}

// KnowledgeGraph represents the complete knowledge graph
type KnowledgeGraph struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Statistics    GraphStats     `json:"statistics"`
}

// GraphStats represents statistics about the knowledge graph
type GraphStats struct {
	EntityCount       int                    `json:"entity_count"`
	RelationshipCount int                    `json:"relationship_count"`
	EntityTypes       map[string]int         `json:"entity_types"`
	RelationshipTypes map[string]int         `json:"relationship_types"`
	TopEntities       []EntityFrequency      `json:"top_entities"`
}

// EntityFrequency represents an entity with its frequency
type EntityFrequency struct {
	Entity string `json:"entity"`
	Count  int    `json:"count"`
}

// NewKnowledgeGraphTool creates a new knowledge graph tool instance
func NewKnowledgeGraphTool() *KnowledgeGraphTool {
	return &KnowledgeGraphTool{}
}

// Definition returns the tool definition for MCP
func (k *KnowledgeGraphTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "knowledge_graph",
		Description: "Build and analyze knowledge graphs from text - extract entities, relationships, and semantic connections for deep research analysis.",
		InputSchema: mcp.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "The text to analyze and build knowledge graph from",
				},
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"build", "analyze", "visualize", "query"},
					"description": "Operation to perform: build graph, analyze existing, visualize, or query",
					"default":     "build",
				},
				"entity_types": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Types of entities to extract (person, organization, location, concept, etc.)",
					"default":     []string{"person", "organization", "location", "concept"},
				},
				"max_entities": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of entities to extract (default: 50)",
					"default":     50,
					"minimum":     10,
					"maximum":     200,
				},
				"relationship_threshold": map[string]interface{}{
					"type":        "number",
					"description": "Minimum weight threshold for relationships (default: 1.0)",
					"default":     1.0,
					"minimum":     0.1,
					"maximum":     10.0,
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Query string for graph querying (only used with query operation)",
				},
			},
			Required: []string{"text"},
		},
	}
}

// Execute performs knowledge graph operations
func (k *KnowledgeGraphTool) Execute(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	text, ok := params["text"].(string)
	if !ok || text == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: text parameter is required and must be non-empty",
			}},
			IsError: true,
		}, nil
	}

	operation := "build"
	if op, ok := params["operation"].(string); ok {
		operation = op
	}

	entityTypes := []string{"person", "organization", "location", "concept"}
	if et, ok := params["entity_types"].([]interface{}); ok {
		entityTypes = make([]string, 0, len(et))
		for _, t := range et {
			if ts, ok := t.(string); ok {
				entityTypes = append(entityTypes, ts)
			}
		}
	}

	maxEntities := 50
	if me, ok := params["max_entities"]; ok {
		if meFloat, ok := me.(float64); ok {
			maxEntities = int(meFloat)
		}
	}

	relationshipThreshold := 1.0
	if rt, ok := params["relationship_threshold"]; ok {
		if rtFloat, ok := rt.(float64); ok {
			relationshipThreshold = rtFloat
		}
	}

	switch operation {
	case "build":
		return k.buildKnowledgeGraph(text, entityTypes, maxEntities, relationshipThreshold)
	case "analyze":
		return k.analyzeText(text, entityTypes)
	case "visualize":
		return k.visualizeGraph(text, entityTypes, maxEntities)
	case "query":
		query, ok := params["query"].(string)
		if !ok || query == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "Error: query parameter is required for query operation",
				}},
				IsError: true,
			}, nil
		}
		return k.queryGraph(text, query, entityTypes)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: unsupported operation '%s'", operation),
			}},
			IsError: true,
		}, nil
	}
}

// buildKnowledgeGraph builds a complete knowledge graph from text
func (k *KnowledgeGraphTool) buildKnowledgeGraph(text string, entityTypes []string, maxEntities int, threshold float64) (*mcp.CallToolResult, error) {
	// Extract entities
	entities := k.extractEntities(text, entityTypes, maxEntities)
	
	// Build relationships
	relationships := k.extractRelationships(text, entities, threshold)
	
	// Calculate statistics
	stats := k.calculateStatistics(entities, relationships)
	
	// Create knowledge graph
	graph := KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
		Statistics:    stats,
	}

	// Format response
	responseText := k.formatKnowledgeGraph(graph)
	
	// JSON format
	jsonGraph, _ := json.MarshalIndent(graph, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: responseText,
			},
			{
				Type:     "text",
				Text:     string(jsonGraph),
				MimeType: "application/json",
			},
		},
		IsError: false,
	}, nil
}

// extractEntities extracts entities from text based on specified types
func (k *KnowledgeGraphTool) extractEntities(text string, entityTypes []string, maxEntities int) []Entity {
	entities := make(map[string]*Entity)
	
	// Extract different types of entities
	for _, entityType := range entityTypes {
		switch entityType {
		case "person":
			k.extractPersons(text, entities)
		case "organization":
			k.extractOrganizations(text, entities)
		case "location":
			k.extractLocations(text, entities)
		case "concept":
			k.extractConcepts(text, entities)
		case "date":
			k.extractDates(text, entities)
		case "number":
			k.extractNumbers(text, entities)
		}
	}

	// Convert to slice and sort by mentions
	var entityList []Entity
	for _, entity := range entities {
		entityList = append(entityList, *entity)
	}

	sort.Slice(entityList, func(i, j int) bool {
		return entityList[i].Mentions > entityList[j].Mentions
	})

	// Limit results
	if len(entityList) > maxEntities {
		entityList = entityList[:maxEntities]
	}

	return entityList
}

// extractPersons extracts person entities
func (k *KnowledgeGraphTool) extractPersons(text string, entities map[string]*Entity) {
	// Pattern for names (simplified)
	pattern := regexp.MustCompile(`\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`)
	matches := pattern.FindAllString(text, -1)

	for _, match := range matches {
		if k.isLikelyPersonName(match) {
			id := strings.ToLower(strings.ReplaceAll(match, " ", "_"))
			if entity, exists := entities[id]; exists {
				entity.Mentions++
			} else {
				entities[id] = &Entity{
					ID:          id,
					Name:        match,
					Type:        "person",
					Attributes:  make(map[string]string),
					Mentions:    1,
				}
			}
		}
	}
}

// extractOrganizations extracts organization entities
func (k *KnowledgeGraphTool) extractOrganizations(text string, entities map[string]*Entity) {
	// Common organization suffixes and names
	patterns := []string{
		`\b[A-Z][a-zA-Z\s&]+(?:Inc|Corp|LLC|Ltd|Company|Corporation|Organization|Institute|University|College|School)\b`,
		`\b(?:Apple|Google|Microsoft|Amazon|Meta|Tesla|Netflix|IBM|Oracle|Adobe|Intel|AMD|Nvidia|OpenAI|Anthropic)\b`,
		`\b[A-Z][A-Z]+\b`, // Acronyms
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindAllString(text, -1)

		for _, match := range matches {
			if len(match) > 2 && !k.isCommonWord(match) {
				id := strings.ToLower(strings.ReplaceAll(match, " ", "_"))
				if entity, exists := entities[id]; exists {
					entity.Mentions++
				} else {
					entities[id] = &Entity{
						ID:          id,
						Name:        match,
						Type:        "organization",
						Attributes:  make(map[string]string),
						Mentions:    1,
					}
				}
			}
		}
	}
}

// extractLocations extracts location entities
func (k *KnowledgeGraphTool) extractLocations(text string, entities map[string]*Entity) {
	// Common location patterns
	locations := []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose",
		"London", "Paris", "Tokyo", "Beijing", "Shanghai", "Mumbai", "Delhi", "Bangkok", "Jakarta", "Manila",
		"California", "Texas", "Florida", "New York", "Pennsylvania", "Illinois", "Ohio", "Georgia", "North Carolina", "Michigan",
		"United States", "China", "India", "Japan", "Germany", "United Kingdom", "France", "Brazil", "Italy", "Canada",
	}

	lowerText := strings.ToLower(text)
	for _, location := range locations {
		count := strings.Count(lowerText, strings.ToLower(location))
		if count > 0 {
			id := strings.ToLower(strings.ReplaceAll(location, " ", "_"))
			entities[id] = &Entity{
				ID:          id,
				Name:        location,
				Type:        "location",
				Attributes:  make(map[string]string),
				Mentions:    count,
			}
		}
	}
}

// extractConcepts extracts conceptual entities
func (k *KnowledgeGraphTool) extractConcepts(text string, entities map[string]*Entity) {
	// Common technical and conceptual terms
	concepts := []string{
		"artificial intelligence", "machine learning", "deep learning", "neural network", "algorithm",
		"blockchain", "cryptocurrency", "bitcoin", "ethereum", "database", "software", "hardware",
		"cloud computing", "cybersecurity", "data science", "big data", "analytics", "automation",
		"innovation", "technology", "research", "development", "strategy", "management", "leadership",
		"sustainability", "climate change", "renewable energy", "environment", "economics", "finance",
	}

	lowerText := strings.ToLower(text)
	for _, concept := range concepts {
		count := strings.Count(lowerText, concept)
		if count > 0 {
			id := strings.ToLower(strings.ReplaceAll(concept, " ", "_"))
			entities[id] = &Entity{
				ID:          id,
				Name:        concept,
				Type:        "concept",
				Attributes:  make(map[string]string),
				Mentions:    count,
			}
		}
	}
}

// extractDates extracts date entities
func (k *KnowledgeGraphTool) extractDates(text string, entities map[string]*Entity) {
	patterns := []string{
		`\b\d{4}-\d{2}-\d{2}\b`,                    // YYYY-MM-DD
		`\b\d{1,2}/\d{1,2}/\d{4}\b`,               // MM/DD/YYYY
		`\b\d{1,2}/\d{1,2}/\d{2}\b`,               // MM/DD/YY
		`\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2},?\s+\d{4}\b`, // Month DD, YYYY
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindAllString(text, -1)

		for _, match := range matches {
			id := strings.ToLower(strings.ReplaceAll(match, " ", "_"))
			if entity, exists := entities[id]; exists {
				entity.Mentions++
			} else {
				entities[id] = &Entity{
					ID:          id,
					Name:        match,
					Type:        "date",
					Attributes:  make(map[string]string),
					Mentions:    1,
				}
			}
		}
	}
}

// extractNumbers extracts numeric entities
func (k *KnowledgeGraphTool) extractNumbers(text string, entities map[string]*Entity) {
	pattern := regexp.MustCompile(`\b\d+(?:\.\d+)?(?:[KMB]|million|billion|thousand)?\b`)
	matches := pattern.FindAllString(text, -1)

	for _, match := range matches {
		if len(match) > 2 { // Skip small numbers
			id := strings.ToLower(match)
			if entity, exists := entities[id]; exists {
				entity.Mentions++
			} else {
				entities[id] = &Entity{
					ID:          id,
					Name:        match,
					Type:        "number",
					Attributes:  make(map[string]string),
					Mentions:    1,
				}
			}
		}
	}
}

// extractRelationships extracts relationships between entities
func (k *KnowledgeGraphTool) extractRelationships(text string, entities []Entity, threshold float64) []Relationship {
	var relationships []Relationship
	relationshipMap := make(map[string]*Relationship)

	sentences := k.splitIntoSentences(text)

	// Look for co-occurrences in sentences
	for _, sentence := range sentences {
		lowerSentence := strings.ToLower(sentence)
		var foundEntities []Entity

		// Find entities in this sentence
		for _, entity := range entities {
			if strings.Contains(lowerSentence, strings.ToLower(entity.Name)) {
				foundEntities = append(foundEntities, entity)
			}
		}

		// Create relationships between co-occurring entities
		for i, entity1 := range foundEntities {
			for j, entity2 := range foundEntities {
				if i != j {
					relType := k.inferRelationshipType(sentence, entity1, entity2)
					relID := fmt.Sprintf("%s_%s_%s", entity1.ID, relType, entity2.ID)

					if rel, exists := relationshipMap[relID]; exists {
						rel.Weight++
					} else {
						relationshipMap[relID] = &Relationship{
							ID:     relID,
							Source: entity1.ID,
							Target: entity2.ID,
							Type:   relType,
							Weight: 1,
						}
					}
				}
			}
		}
	}

	// Filter by threshold and convert to slice
	for _, rel := range relationshipMap {
		if float64(rel.Weight) >= threshold {
			relationships = append(relationships, *rel)
		}
	}

	// Sort by weight
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].Weight > relationships[j].Weight
	})

	return relationships
}

// inferRelationshipType infers the type of relationship between entities
func (k *KnowledgeGraphTool) inferRelationshipType(sentence string, entity1, entity2 Entity) string {
	lowerSentence := strings.ToLower(sentence)

	// Different relationship patterns
	if strings.Contains(lowerSentence, "work") || strings.Contains(lowerSentence, "employ") {
		return "works_at"
	}
	if strings.Contains(lowerSentence, "founded") || strings.Contains(lowerSentence, "created") {
		return "founded"
	}
	if strings.Contains(lowerSentence, "located") || strings.Contains(lowerSentence, "based") {
		return "located_in"
	}
	if strings.Contains(lowerSentence, "partner") || strings.Contains(lowerSentence, "collaborate") {
		return "partners_with"
	}
	if strings.Contains(lowerSentence, "acquire") || strings.Contains(lowerSentence, "bought") {
		return "acquired"
	}
	if strings.Contains(lowerSentence, "compete") || strings.Contains(lowerSentence, "rival") {
		return "competes_with"
	}

	// Default relationship based on entity types
	if entity1.Type == "person" && entity2.Type == "organization" {
		return "associated_with"
	}
	if entity1.Type == "organization" && entity2.Type == "location" {
		return "based_in"
	}

	return "related_to"
}

// calculateStatistics calculates graph statistics
func (k *KnowledgeGraphTool) calculateStatistics(entities []Entity, relationships []Relationship) GraphStats {
	stats := GraphStats{
		EntityCount:       len(entities),
		RelationshipCount: len(relationships),
		EntityTypes:       make(map[string]int),
		RelationshipTypes: make(map[string]int),
	}

	// Count entity types
	for _, entity := range entities {
		stats.EntityTypes[entity.Type]++
	}

	// Count relationship types
	for _, rel := range relationships {
		stats.RelationshipTypes[rel.Type]++
	}

	// Top entities by mentions
	var entityFreqs []EntityFrequency
	for _, entity := range entities {
		entityFreqs = append(entityFreqs, EntityFrequency{
			Entity: entity.Name,
			Count:  entity.Mentions,
		})
	}

	sort.Slice(entityFreqs, func(i, j int) bool {
		return entityFreqs[i].Count > entityFreqs[j].Count
	})

	if len(entityFreqs) > 10 {
		entityFreqs = entityFreqs[:10]
	}
	stats.TopEntities = entityFreqs

	return stats
}

// Helper functions
func (k *KnowledgeGraphTool) isLikelyPersonName(name string) bool {
	// Simple heuristics for person names
	parts := strings.Fields(name)
	if len(parts) != 2 {
		return false
	}
	
	// Check if both parts are capitalized and reasonable length
	for _, part := range parts {
		if len(part) < 2 || len(part) > 15 {
			return false
		}
	}
	
	// Exclude common non-person combinations
	excludePatterns := []string{
		"New York", "Los Angeles", "San Francisco", "Las Vegas", "Real Estate",
		"High School", "Middle School", "Public Health", "Social Media",
		"Machine Learning", "Artificial Intelligence", "Big Data",
	}
	
	for _, pattern := range excludePatterns {
		if name == pattern {
			return false
		}
	}
	
	return true
}

func (k *KnowledgeGraphTool) isCommonWord(word string) bool {
	common := map[string]bool{
		"THE": true, "AND": true, "FOR": true, "ARE": true, "BUT": true, "NOT": true,
		"YOU": true, "ALL": true, "CAN": true, "HAD": true, "HER": true, "WAS": true,
		"ONE": true, "OUR": true, "OUT": true, "DAY": true, "GET": true, "HAS": true,
		"HIM": true, "HIS": true, "HOW": true, "ITS": true, "NEW": true, "NOW": true,
		"OLD": true, "SEE": true, "TWO": true, "WHO": true, "BOY": true, "DID": true,
		"TOO": true, "USE": true, "WAY": true, "SHE": true, "MAY": true, "SAY": true,
	}
	return common[strings.ToUpper(word)]
}

func (k *KnowledgeGraphTool) splitIntoSentences(text string) []string {
	pattern := regexp.MustCompile(`[.!?]+\s+`)
	sentences := pattern.Split(text, -1)
	
	var result []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 { // Filter very short sentences
			result = append(result, sentence)
		}
	}
	
	return result
}

// Response formatting functions
func (k *KnowledgeGraphTool) formatKnowledgeGraph(graph KnowledgeGraph) string {
	var builder strings.Builder
	
	builder.WriteString("🕸️ **Knowledge Graph Analysis**\n\n")
	
	// Statistics
	builder.WriteString("**Graph Statistics:**\n")
	builder.WriteString(fmt.Sprintf("- Entities: %d\n", graph.Statistics.EntityCount))
	builder.WriteString(fmt.Sprintf("- Relationships: %d\n", graph.Statistics.RelationshipCount))
	
	builder.WriteString("\n**Entity Types:**\n")
	for entityType, count := range graph.Statistics.EntityTypes {
		builder.WriteString(fmt.Sprintf("- %s: %d\n", entityType, count))
	}
	
	builder.WriteString("\n**Top Entities:**\n")
	for _, entity := range graph.Statistics.TopEntities {
		builder.WriteString(fmt.Sprintf("- %s (%d mentions)\n", entity.Entity, entity.Count))
	}
	
	builder.WriteString("\n**Relationship Types:**\n")
	for relType, count := range graph.Statistics.RelationshipTypes {
		builder.WriteString(fmt.Sprintf("- %s: %d\n", relType, count))
	}
	
	builder.WriteString("\n**Key Relationships:**\n")
	for i, rel := range graph.Relationships {
		if i >= 10 { // Show top 10 relationships
			break
		}
		
		// Find entity names
		var sourceName, targetName string
		for _, entity := range graph.Entities {
			if entity.ID == rel.Source {
				sourceName = entity.Name
			}
			if entity.ID == rel.Target {
				targetName = entity.Name
			}
		}
		
		builder.WriteString(fmt.Sprintf("- %s --[%s]--> %s (weight: %d)\n", 
			sourceName, rel.Type, targetName, rel.Weight))
	}
	
	return builder.String()
}

// Additional operations
func (k *KnowledgeGraphTool) analyzeText(text string, entityTypes []string) (*mcp.CallToolResult, error) {
	entities := k.extractEntities(text, entityTypes, 30)
	
	var builder strings.Builder
	builder.WriteString("🔍 **Entity Analysis**\n\n")
	
	typeGroups := make(map[string][]Entity)
	for _, entity := range entities {
		typeGroups[entity.Type] = append(typeGroups[entity.Type], entity)
	}
	
	for entityType, group := range typeGroups {
		builder.WriteString(fmt.Sprintf("**%s (%d):**\n", strings.Title(entityType), len(group)))
		for _, entity := range group {
			builder.WriteString(fmt.Sprintf("- %s (%d mentions)\n", entity.Name, entity.Mentions))
		}
		builder.WriteString("\n")
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: builder.String(),
		}},
		IsError: false,
	}, nil
}

func (k *KnowledgeGraphTool) visualizeGraph(text string, entityTypes []string, maxEntities int) (*mcp.CallToolResult, error) {
	entities := k.extractEntities(text, entityTypes, maxEntities)
	relationships := k.extractRelationships(text, entities, 1.0)
	
	// Create a simple text-based visualization
	var builder strings.Builder
	builder.WriteString("📊 **Knowledge Graph Visualization**\n\n")
	
	// Group entities by type for better visualization
	typeGroups := make(map[string][]Entity)
	for _, entity := range entities {
		typeGroups[entity.Type] = append(typeGroups[entity.Type], entity)
	}
	
	// Show entity clusters
	for entityType, group := range typeGroups {
		builder.WriteString(fmt.Sprintf("**%s Cluster:**\n", strings.Title(entityType)))
		for _, entity := range group {
			builder.WriteString(fmt.Sprintf("  ● %s\n", entity.Name))
		}
		builder.WriteString("\n")
	}
	
	// Show key connections
	builder.WriteString("**Key Connections:**\n")
	for i, rel := range relationships {
		if i >= 15 { // Limit connections shown
			break
		}
		
		var sourceName, targetName string
		for _, entity := range entities {
			if entity.ID == rel.Source {
				sourceName = entity.Name
			}
			if entity.ID == rel.Target {
				targetName = entity.Name
			}
		}
		
		arrow := "→"
		if rel.Weight > 2 {
			arrow = "⇒"
		}
		
		builder.WriteString(fmt.Sprintf("  %s %s %s [%s]\n", 
			sourceName, arrow, targetName, rel.Type))
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: builder.String(),
		}},
		IsError: false,
	}, nil
}

func (k *KnowledgeGraphTool) queryGraph(text string, query string, entityTypes []string) (*mcp.CallToolResult, error) {
	entities := k.extractEntities(text, entityTypes, 100)
	
	queryLower := strings.ToLower(query)
	var matchedEntities []Entity
	
	// Simple query matching
	for _, entity := range entities {
		if strings.Contains(strings.ToLower(entity.Name), queryLower) ||
		   strings.Contains(strings.ToLower(entity.Type), queryLower) {
			matchedEntities = append(matchedEntities, entity)
		}
	}
	
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("🔍 **Query Results for '%s'**\n\n", query))
	
	if len(matchedEntities) == 0 {
		builder.WriteString("No matching entities found.\n")
	} else {
		builder.WriteString(fmt.Sprintf("Found %d matching entities:\n\n", len(matchedEntities)))
		
		for _, entity := range matchedEntities {
			builder.WriteString(fmt.Sprintf("- **%s** (%s) - %d mentions\n", 
				entity.Name, entity.Type, entity.Mentions))
		}
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: builder.String(),
		}},
		IsError: false,
	}, nil
}