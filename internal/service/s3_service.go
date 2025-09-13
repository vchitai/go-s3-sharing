package service

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/vchitai/go-s3-sharing/internal/domain"
)

// s3ObjectReader wraps S3 GetObjectOutput to implement ObjectReader
type s3ObjectReader struct {
	body        io.ReadCloser
	contentType string
	size        int64
}

func (r *s3ObjectReader) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *s3ObjectReader) Close() error {
	return r.body.Close()
}

func (r *s3ObjectReader) ContentType() string {
	return r.contentType
}

func (r *s3ObjectReader) Size() int64 {
	return r.size
}

// S3Service implements StorageService for AWS S3
type S3Service struct {
	client *s3.Client
	bucket string
}

// NewS3Service creates a new S3 service
func NewS3Service(client *s3.Client, bucket string) *S3Service {
	return &S3Service{
		client: client,
		bucket: bucket,
	}
}

// GetObject retrieves an object from S3
func (s *S3Service) GetObject(ctx context.Context, key string) (domain.ObjectReader, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &s3ObjectReader{
		body:        result.Body,
		contentType: contentType,
		size:        size,
	}, nil
}

// HeadObject retrieves object metadata from S3
func (s *S3Service) HeadObject(ctx context.Context, key string) (*domain.ObjectMetadata, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to head object from S3: %w", err)
	}

	metadata := &domain.ObjectMetadata{
		Size: 0,
	}

	if result.ContentType != nil {
		metadata.ContentType = *result.ContentType
	}

	if result.ContentLength != nil {
		metadata.Size = *result.ContentLength
	}

	if result.LastModified != nil {
		metadata.LastModified = *result.LastModified
	}

	return metadata, nil
}
