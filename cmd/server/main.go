package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/chongliujia/mcp-go-template/internal/config"
	"github.com/chongliujia/mcp-go-template/internal/server"
	"github.com/chongliujia/mcp-go-template/internal/tools/examples"
	"github.com/chongliujia/mcp-go-template/pkg/mcp"
	"github.com/chongliujia/mcp-go-template/pkg/utils"
)

const (
	AppName    = "mcp-go-template"
	AppVersion = "1.0.0"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		logLevel   = flag.String("log-level", "", "Log level (debug, info, warn, error)")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Show version if requested
	if *version {
		utils.Infof("%s version %s", AppName, AppVersion)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		utils.Fatalf("Failed to load configuration: %v", err)
	}

	// Override log level if specified
	if *logLevel != "" {
		cfg.Logging.Level = *logLevel
	}

	// Configure logging
	utils.SetLogLevel(utils.LogLevel(cfg.Logging.Level))
	if cfg.Logging.Format == "text" {
		utils.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger := utils.GetLogger()
	logger.WithFields(logrus.Fields{
		"name":    cfg.MCP.Name,
		"version": cfg.MCP.Version,
	}).Info("Starting MCP server")

	// Create server capabilities based on configuration
	capabilities := createServerCapabilities(cfg)

	// Create server info
	serverInfo := mcp.ServerInfo{
		Name:    cfg.MCP.Name,
		Version: cfg.MCP.Version,
	}

	// Create MCP handler
	handler := mcp.NewBaseHandler(serverInfo, capabilities)

	// Register example tools if tools are enabled
	if cfg.IsToolsEnabled() {
		if err := registerTools(handler); err != nil {
			logger.WithError(err).Fatal("Failed to register tools")
		}
	}

	// Create and configure server
	srv := server.New(cfg, handler)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigCh:
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
		
		// Wait for server to shutdown
		if err := <-errCh; err != nil {
			logger.WithError(err).Error("Server shutdown error")
			os.Exit(1)
		}
		
	case err := <-errCh:
		if err != nil {
			logger.WithError(err).Fatal("Server error")
		}
	}

	logger.Info("Server stopped")
}

// createServerCapabilities creates server capabilities based on configuration
func createServerCapabilities(cfg *config.Config) mcp.ServerCapabilities {
	capabilities := mcp.ServerCapabilities{}

	if cfg.IsLoggingEnabled() {
		capabilities.Logging = &mcp.LoggingCapability{}
	}

	if cfg.IsToolsEnabled() {
		capabilities.Tools = &mcp.ToolsCapability{
			ListChanged: cfg.MCP.Capabilities.Tools.ListChanged,
		}
	}

	if cfg.IsResourcesEnabled() {
		capabilities.Resources = &mcp.ResourcesCapability{
			Subscribe:   cfg.MCP.Capabilities.Resources.Subscribe,
			ListChanged: cfg.MCP.Capabilities.Resources.ListChanged,
		}
	}

	if cfg.IsPromptsEnabled() {
		capabilities.Prompts = &mcp.PromptsCapability{
			ListChanged: cfg.MCP.Capabilities.Prompts.ListChanged,
		}
	}

	return capabilities
}

// registerTools registers example tools for deep research
func registerTools(handler *mcp.BaseHandler) error {
	// Register calculator tool
	calculator := examples.NewCalculatorTool()
	if err := handler.RegisterTool(calculator); err != nil {
		return err
	}
	utils.Info("Registered calculator tool")

	// Register web search tool for research
	webSearch := examples.NewWebSearchTool()
	if err := handler.RegisterTool(webSearch); err != nil {
		return err
	}
	utils.Info("Registered web search tool")

	// Register document analyzer for research
	docAnalyzer := examples.NewDocumentAnalyzerTool()
	if err := handler.RegisterTool(docAnalyzer); err != nil {
		return err
	}
	utils.Info("Registered document analyzer tool")

	// Register knowledge graph tool for deep research
	knowledgeGraph := examples.NewKnowledgeGraphTool()
	if err := handler.RegisterTool(knowledgeGraph); err != nil {
		return err
	}
	utils.Info("Registered knowledge graph tool")

	utils.Infof("Successfully registered %d research tools", 4)
	return nil
}