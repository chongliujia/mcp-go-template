package prompts

import (
	"fmt"
	"sync"

	"github.com/chongliujia/mcp-go-template/internal/prompts/examples"
	"github.com/chongliujia/mcp-go-template/pkg/mcp"
	"github.com/chongliujia/mcp-go-template/pkg/utils"
)

// Registry manages prompt registration and discovery
type Registry struct {
	prompts map[string]mcp.PromptHandler
	mutex   sync.RWMutex
}

// NewRegistry creates a new prompt registry
func NewRegistry() *Registry {
	return &Registry{
		prompts: make(map[string]mcp.PromptHandler),
	}
}

// Register registers a prompt handler
func (r *Registry) Register(handler mcp.PromptHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	prompt := handler.Definition()
	if prompt == nil {
		return fmt.Errorf("prompt definition cannot be nil")
	}
	if prompt.Name == "" {
		return fmt.Errorf("prompt name cannot be empty")
	}

	if _, exists := r.prompts[prompt.Name]; exists {
		return fmt.Errorf("prompt '%s' is already registered", prompt.Name)
	}

	r.prompts[prompt.Name] = handler
	utils.Infof("Registered prompt: %s", prompt.Name)
	return nil
}

// Unregister removes a prompt from the registry
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.prompts[name]; !exists {
		return fmt.Errorf("prompt '%s' is not registered", name)
	}

	delete(r.prompts, name)
	utils.Infof("Unregistered prompt: %s", name)
	return nil
}

// Get retrieves a prompt handler by name
func (r *Registry) Get(name string) (mcp.PromptHandler, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	handler, exists := r.prompts[name]
	if !exists {
		return nil, fmt.Errorf("prompt '%s' not found", name)
	}

	return handler, nil
}

// List returns all registered prompts
func (r *Registry) List() []*mcp.Prompt {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	prompts := make([]*mcp.Prompt, 0, len(r.prompts))
	for _, handler := range r.prompts {
		prompts = append(prompts, handler.Definition())
	}

	return prompts
}

// Count returns the number of registered prompts
func (r *Registry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.prompts)
}

// RegisterDefaultPrompts registers all default example prompts
func (r *Registry) RegisterDefaultPrompts() error {
	// Register code analysis prompt
	if err := r.Register(examples.NewCodeAnalysisPrompt()); err != nil {
		return fmt.Errorf("failed to register code analysis prompt: %w", err)
	}

	// Register research prompt
	if err := r.Register(examples.NewResearchPrompt()); err != nil {
		return fmt.Errorf("failed to register research prompt: %w", err)
	}

	// Register summarization prompt
	if err := r.Register(examples.NewSummarizationPrompt()); err != nil {
		return fmt.Errorf("failed to register summarization prompt: %w", err)
	}

	utils.Infof("Successfully registered %d default prompts", r.Count())
	return nil
}

// GetPromptNames returns a list of all registered prompt names
func (r *Registry) GetPromptNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.prompts))
	for name := range r.prompts {
		names = append(names, name)
	}

	return names
}

// HasPrompt checks if a prompt is registered
func (r *Registry) HasPrompt(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.prompts[name]
	return exists
}

// Clear removes all registered prompts
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.prompts = make(map[string]mcp.PromptHandler)
	utils.Info("Cleared all registered prompts")
}