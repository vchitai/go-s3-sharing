package domain

import (
	"context"
	"time"
)

// ShareRequest represents a request to create a shareable link
type ShareRequest struct {
	S3Path    string
	Secret    string
	ExpiresAt time.Time
}

// ShareResponse represents the response after creating a shareable link
type ShareResponse struct {
	URL       string
	ExpiresAt time.Time
	MaxAge    time.Duration
}

// ShareService defines the interface for sharing operations
type ShareService interface {
	CreateShare(ctx context.Context, req *ShareRequest) (*ShareResponse, error)
	ValidateShare(ctx context.Context, s3Path, secret string) error
	GetObject(ctx context.Context, s3Path string) (ObjectReader, error)
}

// ObjectReader represents a readable object from storage
type ObjectReader interface {
	Read(p []byte) (n int, err error)
	Close() error
	ContentType() string
	Size() int64
}

// CacheService defines the interface for cache operations
type CacheService interface {
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// StorageService defines the interface for object storage operations
type StorageService interface {
	GetObject(ctx context.Context, key string) (ObjectReader, error)
	HeadObject(ctx context.Context, key string) (*ObjectMetadata, error)
}

// ObjectMetadata contains metadata about a stored object
type ObjectMetadata struct {
	ContentType  string
	Size         int64
	LastModified time.Time
}
