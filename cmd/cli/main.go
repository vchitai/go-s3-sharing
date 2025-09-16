package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"
	"crypto/tls"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/vchitai/go-s3-sharing/internal/config"
	"github.com/vchitai/go-s3-sharing/internal/domain"
	"github.com/vchitai/go-s3-sharing/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-s3-sharing-cli <s3-path> [expiration-hours]")
		fmt.Println("Example: go-s3-sharing-cli images/photo.jpg 24")
		os.Exit(1)
	}

	s3Path := os.Args[1]
	expirationHours := 24

	if len(os.Args) > 2 {
		if _, err := fmt.Sscanf(os.Args[2], "%d", &expirationHours); err != nil {
			log.Fatalf("invalid expiration hours: %v", err)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize AWS S3 client
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize Redis client
	redisOptions := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	if cfg.Redis.TLSEnabled {
		redisOptions.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // #nosec G402
		}
	}
	redisClient := redis.NewClient(redisOptions)

	// Initialize services
	storageService := service.NewS3Service(s3Client, cfg.AWS.Bucket)
	cacheService := service.NewRedisService(redisClient)

	shareConfig := &service.ShareConfig{
		MaxAgeDays: cfg.Security.MaxAgeDays,
		BaseURL:    cfg.BaseURL, // This should come from config
	}

	shareService := service.NewShareService(storageService, cacheService, shareConfig)

	// Generate a secure secret
	secret, err := generateSecret()
	if err != nil {
		log.Fatalf("failed to generate secret: %v", err)
	}

	// Create share request
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	req := &domain.ShareRequest{
		S3Path:    s3Path,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	// Create share
	resp, err := shareService.CreateShare(ctx, req)
	if err != nil {
		log.Fatalf("failed to create share: %v", err)
	}

	// Output results
	fmt.Printf("Shareable URL: %s\n", resp.URL)
	fmt.Printf("Expires at: %s\n", resp.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("Max age: %s\n", resp.MaxAge)
}

// generateSecret generates a cryptographically secure random secret
func generateSecret() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
