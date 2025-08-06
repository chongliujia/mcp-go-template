package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/chongliujia/mcp-go-template/internal/config"
	"github.com/chongliujia/mcp-go-template/pkg/mcp"
	"github.com/chongliujia/mcp-go-template/pkg/utils"
)

// Server represents the MCP server
type Server struct {
	config   *config.Config
	handler  mcp.Handler
	upgrader websocket.Upgrader
	logger   *logrus.Logger
}

// New creates a new MCP server
func New(cfg *config.Config, handler mcp.Handler) *Server {
	return &Server{
		config:  cfg,
		handler: handler,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin in development
				// In production, implement proper origin checking
				return true
			},
		},
		logger: utils.GetLogger(),
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/", s.handleRoot)

	server := &http.Server{
		Addr:         s.config.GetAddress(),
		Handler:      mux,
		ReadTimeout:  time.Duration(s.config.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.Timeout) * time.Second,
	}

	s.logger.WithFields(logrus.Fields{
		"address": s.config.GetAddress(),
		"name":    s.config.MCP.Name,
		"version": s.config.MCP.Version,
	}).Info("Starting MCP server")

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if s.config.Security.EnableTLS {
			errCh <- server.ListenAndServeTLS(s.config.Security.CertFile, s.config.Security.KeyFile)
		} else {
			errCh <- server.ListenAndServe()
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down server...")
		
		// Create a context with timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}

// handleWebSocket handles WebSocket connections for MCP communication
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check allowed IPs if configured
	if len(s.config.Security.AllowedIPs) > 0 {
		clientIP := s.getClientIP(r)
		allowed := false
		for _, ip := range s.config.Security.AllowedIPs {
			if ip == clientIP {
				allowed = true
				break
			}
		}
		if !allowed {
			s.logger.WithField("client_ip", clientIP).Warn("Connection rejected: IP not allowed")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	s.logger.WithField("client", conn.RemoteAddr()).Info("New WebSocket connection")

	// Handle the WebSocket connection
	s.handleConnection(conn)
}

// handleConnection handles a single WebSocket connection
func (s *Server) handleConnection(conn *websocket.Conn) {
	for {
		// Read message
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithError(err).Error("WebSocket read error")
			}
			break
		}

		if messageType != websocket.TextMessage {
			s.logger.Warn("Received non-text message, ignoring")
			continue
		}

		// Parse MCP message
		var message mcp.Message
		if err := json.Unmarshal(data, &message); err != nil {
			s.logger.WithError(err).Error("Failed to parse MCP message")
			
			// Send error response
			errorResponse := mcp.NewErrorResponse(nil, mcp.ParseError, "Invalid JSON", err.Error())
			s.sendMessage(conn, errorResponse)
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"method": message.Method,
			"id":     message.ID,
		}).Debug("Received MCP message")

		// Handle the message
		response, err := s.handler.HandleMessage(context.Background(), &message)
		if err != nil {
			s.logger.WithError(err).Error("Message handling failed")
			
			// Send internal error response
			errorResponse := mcp.NewErrorResponse(message.ID, mcp.InternalError, "Internal server error", err.Error())
			s.sendMessage(conn, errorResponse)
			continue
		}

		// Send response if there is one
		if response != nil {
			if err := s.sendMessage(conn, response); err != nil {
				s.logger.WithError(err).Error("Failed to send response")
				break
			}
		}
	}

	s.logger.WithField("client", conn.RemoteAddr()).Info("WebSocket connection closed")
}

// sendMessage sends a message over the WebSocket connection
func (s *Server) sendMessage(conn *websocket.Conn, message *mcp.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"method": message.Method,
		"id":     message.ID,
		"error":  message.Error != nil,
	}).Debug("Sent MCP message")

	return nil
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"name":    s.config.MCP.Name,
		"version": s.config.MCP.Version,
		"time":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleRoot handles root path requests
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"name":        s.config.MCP.Name,
		"version":     s.config.MCP.Version,
		"description": s.config.MCP.Description,
		"endpoints": map[string]string{
			"websocket": "/mcp",
			"health":    "/health",
		},
		"protocol_version": mcp.MCPVersion,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// getClientIP extracts the client IP from the request
func (s *Server) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Take the first IP if there are multiple
		if idx := len(xForwardedFor); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, c := range xForwardedFor {
					if c == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xForwardedFor[:commaIdx]
				}
			}
		}
		return xForwardedFor
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}