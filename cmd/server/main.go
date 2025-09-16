package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/vchitai/go-s3-sharing/internal/config"
	"github.com/vchitai/go-s3-sharing/internal/service"
	"github.com/vchitai/go-s3-sharing/internal/transport/http"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	// Initialize AWS S3 client
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load AWS config", "error", err)
		os.Exit(1)
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

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Initialize services
	storageService := service.NewS3Service(s3Client, cfg.AWS.Bucket)
	cacheService := service.NewRedisService(redisClient)

	shareConfig := &service.ShareConfig{
		MaxAgeDays: cfg.Security.MaxAgeDays,
		BaseURL:    cfg.BaseURL, // This should come from config
	}

	shareService := service.NewShareService(storageService, cacheService, shareConfig)

	// Initialize HTTP server
	server := http.NewServer(cfg, shareService, logger)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
