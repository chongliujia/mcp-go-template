package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	MCP      MCPConfig      `mapstructure:"mcp"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Timeout int    `mapstructure:"timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MCPConfig represents MCP-specific configuration
type MCPConfig struct {
	Name         string            `mapstructure:"name"`
	Version      string            `mapstructure:"version"`
	Description  string            `mapstructure:"description"`
	Instructions string            `mapstructure:"instructions"`
	Capabilities CapabilityConfig  `mapstructure:"capabilities"`
	Metadata     map[string]string `mapstructure:"metadata"`
}

// CapabilityConfig represents MCP capability configuration
type CapabilityConfig struct {
	Tools     ToolsConfig     `mapstructure:"tools"`
	Resources ResourcesConfig `mapstructure:"resources"`
	Prompts   PromptsConfig   `mapstructure:"prompts"`
	Logging   bool            `mapstructure:"logging"`
}

// ToolsConfig represents tools capability configuration
type ToolsConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	ListChanged bool `mapstructure:"list_changed"`
}

// ResourcesConfig represents resources capability configuration
type ResourcesConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	Subscribe   bool `mapstructure:"subscribe"`
	ListChanged bool `mapstructure:"list_changed"`
}

// PromptsConfig represents prompts capability configuration
type PromptsConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	ListChanged bool `mapstructure:"list_changed"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableTLS  bool     `mapstructure:"enable_tls"`
	CertFile   string   `mapstructure:"cert_file"`
	KeyFile    string   `mapstructure:"key_file"`
	AllowedIPs []string `mapstructure:"allowed_ips"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "localhost",
			Port:    8030,
			Timeout: 30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		MCP: MCPConfig{
			Name:        "mcp-go-template",
			Version:     "1.0.0",
			Description: "A Go-based MCP server template",
			Capabilities: CapabilityConfig{
				Tools: ToolsConfig{
					Enabled:     true,
					ListChanged: false,
				},
				Resources: ResourcesConfig{
					Enabled:     true,
					Subscribe:   false,
					ListChanged: false,
				},
				Prompts: PromptsConfig{
					Enabled:     true,
					ListChanged: false,
				},
				Logging: true,
			},
			Metadata: make(map[string]string),
		},
		Security: SecurityConfig{
			EnableTLS:  false,
			AllowedIPs: []string{},
		},
	}
}

// Load loads configuration from various sources
func Load(configPath string) (*Config, error) {
	// Set default values
	config := DefaultConfig()
	
	// Configure viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/mcp-go-template")
		viper.AddConfigPath("$HOME/.mcp-go-template")
	}

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MCP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values in viper
	setDefaults(config)

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and environment variables
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// setDefaults sets default values in viper
func setDefaults(config *Config) {
	viper.SetDefault("server.host", config.Server.Host)
	viper.SetDefault("server.port", config.Server.Port)
	viper.SetDefault("server.timeout", config.Server.Timeout)
	
	viper.SetDefault("logging.level", config.Logging.Level)
	viper.SetDefault("logging.format", config.Logging.Format)
	
	viper.SetDefault("mcp.name", config.MCP.Name)
	viper.SetDefault("mcp.version", config.MCP.Version)
	viper.SetDefault("mcp.description", config.MCP.Description)
	viper.SetDefault("mcp.instructions", config.MCP.Instructions)
	
	viper.SetDefault("mcp.capabilities.tools.enabled", config.MCP.Capabilities.Tools.Enabled)
	viper.SetDefault("mcp.capabilities.tools.list_changed", config.MCP.Capabilities.Tools.ListChanged)
	viper.SetDefault("mcp.capabilities.resources.enabled", config.MCP.Capabilities.Resources.Enabled)
	viper.SetDefault("mcp.capabilities.resources.subscribe", config.MCP.Capabilities.Resources.Subscribe)
	viper.SetDefault("mcp.capabilities.resources.list_changed", config.MCP.Capabilities.Resources.ListChanged)
	viper.SetDefault("mcp.capabilities.prompts.enabled", config.MCP.Capabilities.Prompts.Enabled)
	viper.SetDefault("mcp.capabilities.prompts.list_changed", config.MCP.Capabilities.Prompts.ListChanged)
	viper.SetDefault("mcp.capabilities.logging", config.MCP.Capabilities.Logging)
	
	viper.SetDefault("security.enable_tls", config.Security.EnableTLS)
	viper.SetDefault("security.cert_file", config.Security.CertFile)
	viper.SetDefault("security.key_file", config.Security.KeyFile)
	viper.SetDefault("security.allowed_ips", config.Security.AllowedIPs)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Server.Timeout <= 0 {
		return fmt.Errorf("server timeout must be positive: %d", config.Server.Timeout)
	}

	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[config.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", config.Logging.Format)
	}

	if config.MCP.Name == "" {
		return fmt.Errorf("MCP name cannot be empty")
	}

	if config.MCP.Version == "" {
		return fmt.Errorf("MCP version cannot be empty")
	}

	if config.Security.EnableTLS {
		if config.Security.CertFile == "" {
			return fmt.Errorf("cert file is required when TLS is enabled")
		}
		if config.Security.KeyFile == "" {
			return fmt.Errorf("key file is required when TLS is enabled")
		}
	}

	return nil
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsToolsEnabled returns whether tools capability is enabled
func (c *Config) IsToolsEnabled() bool {
	return c.MCP.Capabilities.Tools.Enabled
}

// IsResourcesEnabled returns whether resources capability is enabled
func (c *Config) IsResourcesEnabled() bool {
	return c.MCP.Capabilities.Resources.Enabled
}

// IsPromptsEnabled returns whether prompts capability is enabled
func (c *Config) IsPromptsEnabled() bool {
	return c.MCP.Capabilities.Prompts.Enabled
}

// IsLoggingEnabled returns whether logging capability is enabled
func (c *Config) IsLoggingEnabled() bool {
	return c.MCP.Capabilities.Logging
}