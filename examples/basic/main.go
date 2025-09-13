package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/vchitai/go-s3-sharing/internal/config"
	"github.com/vchitai/go-s3-sharing/internal/domain"
	"github.com/vchitai/go-s3-sharing/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize AWS S3 client
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize services
	storageService := service.NewS3Service(s3Client, cfg.AWS.Bucket)
	cacheService := service.NewRedisService(redisClient)

	shareConfig := &service.ShareConfig{
		MaxAgeDays: cfg.Security.MaxAgeDays,
		BaseURL:    "https://your-domain.com",
	}

	shareService := service.NewShareService(storageService, cacheService, shareConfig)

	// Example: Create a shareable link
	s3Path := "images/example.jpg"
	secret, err := generateSecret()
	if err != nil {
		log.Fatalf("Failed to generate secret: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	req := &domain.ShareRequest{
		S3Path:    s3Path,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	fmt.Printf("Creating share for S3 path: %s\n", s3Path)
	fmt.Printf("Secret: %s\n", secret)
	fmt.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))

	resp, err := shareService.CreateShare(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create share: %v", err)
	}

	fmt.Printf("\nShareable URL: %s\n", resp.URL)
	fmt.Printf("Max age: %s\n", resp.MaxAge)

	// Example: Validate a share
	fmt.Printf("\nValidating share...\n")
	err = shareService.ValidateShare(ctx, s3Path, secret)
	if err != nil {
		log.Fatalf("Failed to validate share: %v", err)
	}

	fmt.Println("Share validation successful!")

	// Example: Get object (if it exists)
	fmt.Printf("\nRetrieving object...\n")
	reader, err := shareService.GetObject(ctx, s3Path)
	if err != nil {
		log.Printf("Failed to get object (this is expected if the object doesn't exist): %v", err)
	} else {
		defer reader.Close()
		fmt.Printf("Object retrieved successfully!\n")
		fmt.Printf("Content type: %s\n", reader.ContentType())
		fmt.Printf("Size: %d bytes\n", reader.Size())
	}
}

// generateSecret generates a cryptographically secure random secret
func generateSecret() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
