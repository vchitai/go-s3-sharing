package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vchitai/go-s3-sharing/internal/domain"
	"github.com/vchitai/go-s3-sharing/internal/service"
)

// Handler handles HTTP requests for the S3 sharing service
type Handler struct {
	shareService *service.ShareService
	logger       *slog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(shareService *service.ShareService, logger *slog.Logger) *Handler {
	return &Handler{
		shareService: shareService,
		logger:       logger,
	}
}

// HandleImage handles image sharing requests
func (h *Handler) HandleImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Skip API routes and health checks - these should be handled by specific handlers
	if strings.HasPrefix(r.URL.Path, "/api/") ||
		r.URL.Path == "/health" ||
		r.URL.Path == "/ready" {
		http.NotFound(w, r)
		return
	}

	// Parse URL path: /yy/mm/dd/secret/path/to/file.jpg
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 5 {
		http.NotFound(w, r)
		return
	}

	// Extract components
	dateStr := strings.Join(parts[0:3], "-") // e.g. "25-09-13"
	secret := parts[3]
	s3Path := strings.Join(parts[4:], "/")

	// Validate date
	expiresAt, err := h.parseDate(dateStr)
	if err != nil {
		h.writeError(w, "invalid date format", http.StatusBadRequest)
		h.logger.Error("invalid date", "date", dateStr, "error", err)
		return
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		h.writeError(w, "link expired", http.StatusForbidden)
		h.logger.Info("expired link accessed", "date", dateStr, "age", time.Since(expiresAt))
		return
	}

	// Validate share
	err = h.shareService.ValidateShare(ctx, s3Path, secret)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			h.writeError(w, "unauthorized", http.StatusUnauthorized)
		case domain.ErrInvalidPath:
			h.writeError(w, "invalid path", http.StatusBadRequest)
		default:
			h.writeError(w, "internal error", http.StatusInternalServerError)
			h.logger.Error("share validation failed", "error", err)
		}
		return
	}

	// Get object from storage
	reader, err := h.shareService.GetObject(ctx, s3Path)
	if err != nil {
		h.writeError(w, "not found", http.StatusNotFound)
		h.logger.Error("failed to get object", "path", s3Path, "error", err)
		return
	}
	defer reader.Close()

	// Set response headers
	w.Header().Set("Content-Type", reader.ContentType())
	w.Header().Set("Content-Length", strconv.FormatInt(reader.Size(), 10))
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)

	// Stream the object
	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Error("failed to stream object", "path", s3Path, "error", err)
	}
}

// HandleCreateShare handles share creation requests
func (h *Handler) HandleCreateShare(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.S3Path == "" {
		h.writeError(w, "s3_path is required", http.StatusBadRequest)
		return
	}

	if req.Secret == "" {
		h.writeError(w, "secret is required", http.StatusBadRequest)
		return
	}

	// Set default expiration if not provided
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	// Create share
	shareReq := &domain.ShareRequest{
		S3Path:    req.S3Path,
		Secret:    req.Secret,
		ExpiresAt: expiresAt,
	}

	resp, err := h.shareService.CreateShare(ctx, shareReq)
	if err != nil {
		h.writeError(w, "failed to create share", http.StatusInternalServerError)
		h.logger.Error("failed to create share", "error", err)
		return
	}

	// Return response
	response := CreateShareResponse{
		URL:       resp.URL,
		ExpiresAt: resp.ExpiresAt,
		MaxAge:    int(resp.MaxAge.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseDate parses a date string in YY-MM-DD format
func (h *Handler) parseDate(dateStr string) (time.Time, error) {
	return time.Parse("06-01-02", dateStr)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}

// CreateShareRequest represents a request to create a share
type CreateShareRequest struct {
	S3Path    string    `json:"s3_path"`
	Secret    string    `json:"secret"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// CreateShareResponse represents a response after creating a share
type CreateShareResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxAge    int       `json:"max_age_seconds"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
