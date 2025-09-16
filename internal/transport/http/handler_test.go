package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vchitai/go-s3-sharing/internal/domain"
)

// mockShareService is a mock implementation of ShareService
type mockShareService struct{}

func (m *mockShareService) CreateShare(ctx context.Context, req *domain.ShareRequest) (*domain.ShareResponse, error) {
	return &domain.ShareResponse{
		URL:       "https://example.com/24/12/31/secret/images/photo.jpg",
		ExpiresAt: req.ExpiresAt,
		MaxAge:    time.Hour,
	}, nil
}

func (m *mockShareService) ValidateShare(ctx context.Context, s3Path, secret string) error {
	return nil
}

func (m *mockShareService) GetObject(ctx context.Context, s3Path string) (domain.ObjectReader, error) {
	return &mockObjectReader{}, nil
}

// mockObjectReader is a mock implementation of ObjectReader
type mockObjectReader struct{}

func (m *mockObjectReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *mockObjectReader) Close() error {
	return nil
}

func (m *mockObjectReader) ContentType() string {
	return "image/jpeg"
}

func (m *mockObjectReader) Size() int64 {
	return 1024
}

func TestHandler_Routing(t *testing.T) {
	// Create a simple test that focuses on routing logic
	// We'll test the routing by checking if the right handlers are called
	// For now, let's test the health endpoints which don't need the share service

	// Test health endpoints directly
	t.Run("Health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler := &Handler{}
		handler.HandleHealth(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"healthy"`)) {
			t.Errorf("expected body to contain status healthy, got %s", w.Body.String())
		}
	})

	t.Run("Ready endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		handler := &Handler{}
		handler.HandleReady(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"ready"`)) {
			t.Errorf("expected body to contain status ready, got %s", w.Body.String())
		}
	})
}

// Test routing logic by checking path validation
func TestHandler_PathValidation(t *testing.T) {
	handler := &Handler{}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "API path should return 404 from image handler",
			path:           "/api/shares",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Health path should return 404 from image handler",
			path:           "/health",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Ready path should return 404 from image handler",
			path:           "/ready",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid image path should return 404",
			path:           "/invalid/path",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Root path should return 404",
			path:           "/",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.HandleImage(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
