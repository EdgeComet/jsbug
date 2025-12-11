package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/config"
)

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	logger        *zap.Logger
	httpServer    *http.Server
	startTime     time.Time
	mux           *http.ServeMux
	renderHandler *RenderHandler
	robotsHandler *RobotsHandler
	sseManager    *SSEManager
	sseHandler    *SSEHandler
}

// New creates a new Server instance
func New(cfg *config.Config, logger *zap.Logger) *Server {
	sseManager := NewSSEManager(logger)

	s := &Server{
		config:     cfg,
		logger:     logger,
		startTime:  time.Now(),
		mux:        http.NewServeMux(),
		sseManager: sseManager,
		sseHandler: NewSSEHandler(sseManager, logger),
	}

	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      s.corsMiddleware(s.mux),
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
	}

	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/health", s.healthHandler)
	s.mux.Handle("/api/render/stream", s.sseHandler)

	// Register render handler if set
	if s.renderHandler != nil {
		s.mux.Handle("/api/render", s.renderHandler)
	}
}

// SetRenderHandler sets the render handler for the server
func (s *Server) SetRenderHandler(handler *RenderHandler) {
	s.renderHandler = handler
	// Re-register routes to include the new handler
	s.mux.Handle("/api/render", handler)
}

// SetRobotsHandler sets the robots handler for the server
func (s *Server) SetRobotsHandler(handler *RobotsHandler) {
	s.robotsHandler = handler
	s.mux.Handle("/api/robots", handler)
}

// SSEManager returns the server's SSE manager
func (s *Server) SSEManager() *SSEManager {
	return s.sseManager
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server",
		zap.String("addr", s.httpServer.Addr),
	)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// corsMiddleware handles CORS headers and preflight requests
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the origin is in the allowed list
func (s *Server) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range s.config.Server.CORSOrigins {
		if allowed == "*" {
			return true
		}
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}

	return false
}

// Uptime returns the server uptime in seconds
func (s *Server) Uptime() int64 {
	return int64(time.Since(s.startTime).Seconds())
}
