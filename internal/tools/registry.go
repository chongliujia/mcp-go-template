package tools

import (
	"fmt"
	"sync"

	"github.com/chongliujia/mcp-go-template/internal/tools/examples"
	"github.com/chongliujia/mcp-go-template/pkg/mcp"
	"github.com/chongliujia/mcp-go-template/pkg/utils"
)

// Registry manages tool registration and discovery
type Registry struct {
	tools map[string]mcp.ToolHandler
	mutex sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]mcp.ToolHandler),
	}
}

// Register registers a tool handler
func (r *Registry) Register(handler mcp.ToolHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tool := handler.Definition()
	if tool == nil {
		return fmt.Errorf("tool definition cannot be nil")
	}
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool '%s' is already registered", tool.Name)
	}

	r.tools[tool.Name] = handler
	utils.Infof("Registered tool: %s", tool.Name)
	return nil
}

// Unregister removes a tool from the registry
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool '%s' is not registered", name)
	}

	delete(r.tools, name)
	utils.Infof("Unregistered tool: %s", name)
	return nil
}

// Get retrieves a tool handler by name
func (r *Registry) Get(name string) (mcp.ToolHandler, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	handler, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	return handler, nil
}

// List returns all registered tools
func (r *Registry) List() []*mcp.Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]*mcp.Tool, 0, len(r.tools))
	for _, handler := range r.tools {
		tools = append(tools, handler.Definition())
	}

	return tools
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.tools)
}

// RegisterDefaultTools registers all default example tools
func (r *Registry) RegisterDefaultTools() error {
	// Register calculator tool
	if err := r.Register(examples.NewCalculatorTool()); err != nil {
		return fmt.Errorf("failed to register calculator tool: %w", err)
	}

	// Register web search tool
	if err := r.Register(examples.NewWebSearchTool()); err != nil {
		return fmt.Errorf("failed to register web search tool: %w", err)
	}

	// Register document analyzer tool
	if err := r.Register(examples.NewDocumentAnalyzerTool()); err != nil {
		return fmt.Errorf("failed to register document analyzer tool: %w", err)
	}

	// Register knowledge graph tool
	if err := r.Register(examples.NewKnowledgeGraphTool()); err != nil {
		return fmt.Errorf("failed to register knowledge graph tool: %w", err)
	}

	utils.Infof("Successfully registered %d default tools", r.Count())
	return nil
}

// GetToolNames returns a list of all registered tool names
func (r *Registry) GetToolNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}

// HasTool checks if a tool is registered
func (r *Registry) HasTool(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.tools[name]
	return exists
}

// Clear removes all registered tools
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tools = make(map[string]mcp.ToolHandler)
	utils.Info("Cleared all registered tools")
}