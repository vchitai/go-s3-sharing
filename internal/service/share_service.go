package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/vchitai/go-s3-sharing/internal/domain"
)

// ShareService implements the domain ShareService interface
type ShareService struct {
	storage domain.StorageService
	cache   domain.CacheService
	config  *ShareConfig
}

// ShareConfig holds configuration for the share service
type ShareConfig struct {
	MaxAgeDays int
	BaseURL    string
}

// NewShareService creates a new share service
func NewShareService(storage domain.StorageService, cache domain.CacheService, config *ShareConfig) *ShareService {
	return &ShareService{
		storage: storage,
		cache:   cache,
		config:  config,
	}
}

// CreateShare creates a new shareable link
func (s *ShareService) CreateShare(ctx context.Context, req *domain.ShareRequest) (*domain.ShareResponse, error) {
	// Validate S3 path
	if !s.isValidS3Path(req.S3Path) {
		return nil, domain.ErrInvalidPath
	}

	// Check if object exists
	_, err := s.storage.HeadObject(ctx, req.S3Path)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}

	// Generate cache key
	cacheKey := s.generateCacheKey(req.S3Path)

	// Store in cache
	expiration := time.Until(req.ExpiresAt)
	if expiration <= 0 {
		return nil, fmt.Errorf("expiration time must be in the future")
	}

	err = s.cache.Set(ctx, cacheKey, req.Secret, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store share in cache: %w", err)
	}

	// Generate shareable URL
	url := s.generateShareURL(req.S3Path, req.ExpiresAt)

	return &domain.ShareResponse{
		URL:       url,
		ExpiresAt: req.ExpiresAt,
		MaxAge:    expiration,
	}, nil
}

// ValidateShare validates a share request
func (s *ShareService) ValidateShare(ctx context.Context, s3Path, secret string) error {
	// Validate S3 path
	if !s.isValidS3Path(s3Path) {
		return domain.ErrInvalidPath
	}

	// Check cache
	cacheKey := s.generateCacheKey(s3Path)
	cachedSecret, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.ErrUnauthorized
		}
		return fmt.Errorf("failed to validate share: %w", err)
	}

	// Validate secret
	if cachedSecret != secret {
		return domain.ErrUnauthorized
	}

	return nil
}

// GetObject retrieves an object for sharing
func (s *ShareService) GetObject(ctx context.Context, s3Path string) (domain.ObjectReader, error) {
	// Validate S3 path
	if !s.isValidS3Path(s3Path) {
		return nil, domain.ErrInvalidPath
	}

	// Get object from storage
	reader, err := s.storage.GetObject(ctx, s3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return reader, nil
}

// isValidS3Path validates that the S3 path is safe
func (s *ShareService) isValidS3Path(s3Path string) bool {
	// Clean the path to prevent directory traversal
	cleanPath := path.Clean(s3Path)

	// Ensure it doesn't start with "/" or contain ".."
	if strings.HasPrefix(cleanPath, "/") || strings.Contains(cleanPath, "..") {
		return false
	}

	// Ensure it's not empty
	return cleanPath != "" && cleanPath != "."
}

// generateCacheKey creates a cache key for the S3 path
func (s *ShareService) generateCacheKey(s3Path string) string {
	return fmt.Sprintf("image-auth:%s", s3Path)
}

// generateShareURL creates a shareable URL
func (s *ShareService) generateShareURL(s3Path string, expiresAt time.Time) string {
	// Format date as YY/MM/DD
	dateStr := expiresAt.Format("06/01/02")

	// Generate a secret (in real implementation, this should be cryptographically secure)
	secret := s.generateSecret(s3Path, expiresAt)

	// Construct URL
	return fmt.Sprintf("%s/%s/%s/%s", s.config.BaseURL, dateStr, secret, s3Path)
}

// generateSecret generates a secret for the share (simplified for demo)
func (s *ShareService) generateSecret(s3Path string, expiresAt time.Time) string {
	// In a real implementation, use a proper secret generation method
	// This is just for demonstration
	return fmt.Sprintf("secret_%x", len(s3Path)+int(expiresAt.Unix()))
}
