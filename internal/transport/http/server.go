package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vchitai/go-s3-sharing/internal/config"
	"github.com/vchitai/go-s3-sharing/internal/service"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	logger *slog.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, shareService *service.ShareService, logger *slog.Logger) *Server {
	handler := NewHandler(shareService, logger)

	mux := http.NewServeMux()
	// Register specific routes first (most specific to least specific)
	mux.HandleFunc("/api/shares", handler.HandleCreateShare)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/ready", handler.HandleReady)
	// Register the catch-all image handler last
	mux.HandleFunc("/", handler.HandleImage)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		server: server,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping server")
	return s.server.Shutdown(ctx)
}

// HandleHealth handles health check requests
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}

// HandleReady handles readiness check requests
func (h *Handler) HandleReady(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, check dependencies (Redis, S3)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ready"}`)
}
