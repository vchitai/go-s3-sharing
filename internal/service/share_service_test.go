package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vchitai/go-s3-sharing/internal/domain"
)

// mockStorageService is a mock implementation of StorageService
type mockStorageService struct {
	objects map[string]*domain.ObjectMetadata
}

func (m *mockStorageService) GetObject(ctx context.Context, key string) (domain.ObjectReader, error) {
	if _, exists := m.objects[key]; !exists {
		return nil, domain.ErrNotFound
	}
	return &mockObjectReader{}, nil
}

func (m *mockStorageService) HeadObject(ctx context.Context, key string) (*domain.ObjectMetadata, error) {
	if metadata, exists := m.objects[key]; exists {
		return metadata, nil
	}
	return nil, domain.ErrNotFound
}

// mockCacheService is a mock implementation of CacheService
type mockCacheService struct {
	store map[string]string
}

func (m *mockCacheService) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockCacheService) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.store[key]; exists {
		return value, nil
	}
	return "", domain.ErrNotFound
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	delete(m.store, key)
	return nil
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

func TestShareService_CreateShare(t *testing.T) {
	tests := []struct {
		name        string
		req         *domain.ShareRequest
		setupMocks  func(*mockStorageService, *mockCacheService)
		expectError bool
		errorType   error
	}{
		{
			name: "successful share creation",
			req: &domain.ShareRequest{
				S3Path:    "images/photo.jpg",
				Secret:    "test-secret",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(storage *mockStorageService, cache *mockCacheService) {
				storage.objects["images/photo.jpg"] = &domain.ObjectMetadata{
					ContentType: "image/jpeg",
					Size:        1024,
				}
			},
			expectError: false,
		},
		{
			name: "object not found",
			req: &domain.ShareRequest{
				S3Path:    "images/nonexistent.jpg",
				Secret:    "test-secret",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(storage *mockStorageService, cache *mockCacheService) {
				// No objects in storage
			},
			expectError: true,
			errorType:   domain.ErrNotFound,
		},
		{
			name: "invalid S3 path",
			req: &domain.ShareRequest{
				S3Path:    "../etc/passwd",
				Secret:    "test-secret",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(storage *mockStorageService, cache *mockCacheService) {
				// No setup needed
			},
			expectError: true,
			errorType:   domain.ErrInvalidPath,
		},
		{
			name: "expired link",
			req: &domain.ShareRequest{
				S3Path:    "images/photo.jpg",
				Secret:    "test-secret",
				ExpiresAt: time.Now().Add(-24 * time.Hour), // Past time
			},
			setupMocks: func(storage *mockStorageService, cache *mockCacheService) {
				storage.objects["images/photo.jpg"] = &domain.ObjectMetadata{
					ContentType: "image/jpeg",
					Size:        1024,
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &mockStorageService{objects: make(map[string]*domain.ObjectMetadata)}
			cache := &mockCacheService{store: make(map[string]string)}

			tt.setupMocks(storage, cache)

			service := NewShareService(storage, cache, &ShareConfig{
				MaxAgeDays: 90,
				BaseURL:    "https://example.com",
			})

			resp, err := service.CreateShare(context.Background(), tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("expected error type %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Errorf("expected response but got nil")
				return
			}

			if resp.URL == "" {
				t.Errorf("expected URL but got empty string")
			}

			if resp.ExpiresAt != tt.req.ExpiresAt {
				t.Errorf("expected expires at %v, got %v", tt.req.ExpiresAt, resp.ExpiresAt)
			}
		})
	}
}

func TestShareService_ValidateShare(t *testing.T) {
	tests := []struct {
		name        string
		s3Path      string
		secret      string
		setupMocks  func(*mockCacheService)
		expectError bool
		errorType   error
	}{
		{
			name:   "valid share",
			s3Path: "images/photo.jpg",
			secret: "test-secret",
			setupMocks: func(cache *mockCacheService) {
				cache.store["image-auth:images/photo.jpg"] = "test-secret"
			},
			expectError: false,
		},
		{
			name:   "invalid secret",
			s3Path: "images/photo.jpg",
			secret: "wrong-secret",
			setupMocks: func(cache *mockCacheService) {
				cache.store["image-auth:images/photo.jpg"] = "test-secret"
			},
			expectError: true,
			errorType:   domain.ErrUnauthorized,
		},
		{
			name:   "share not found",
			s3Path: "images/photo.jpg",
			secret: "test-secret",
			setupMocks: func(cache *mockCacheService) {
				// No shares in cache
			},
			expectError: true,
			errorType:   domain.ErrUnauthorized,
		},
		{
			name:   "invalid path",
			s3Path: "../etc/passwd",
			secret: "test-secret",
			setupMocks: func(cache *mockCacheService) {
				// No setup needed
			},
			expectError: true,
			errorType:   domain.ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockCacheService{store: make(map[string]string)}
			tt.setupMocks(cache)

			service := NewShareService(nil, cache, &ShareConfig{
				MaxAgeDays: 90,
				BaseURL:    "https://example.com",
			})

			err := service.ValidateShare(context.Background(), tt.s3Path, tt.secret)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("expected error type %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
