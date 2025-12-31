package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Client wraps minio client with production-ready features
type S3Client struct {
	client *minio.Client
	bucket string
	config *Config
	logger *log.Logger
}

// NewS3Client creates and initializes a new S3 client
func NewS3Client(config *Config, logger *log.Logger) (*S3Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create minio client with production-ready settings
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	s3Client := &S3Client{
		client: client,
		bucket: config.Bucket,
		config: config,
		logger: logger,
	}

	// Verify connection
	if err := s3Client.VerifyConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify S3 connection: %w", err)
	}

	return s3Client, nil
}

// VerifyConnection tests the S3 connection
func (s *S3Client) VerifyConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", s.bucket)
	}

	s.logger.Printf("✓ Successfully connected to S3 bucket: %s", s.bucket)
	return nil
}

// CreateBucket creates a new bucket if it doesn't exist
func (s *S3Client) CreateBucket(ctx context.Context, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if exists {
		s.logger.Printf("Bucket %s already exists", bucketName)
		return nil
	}

	err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	s.logger.Printf("✓ Created bucket: %s", bucketName)
	return nil
}

// ListBuckets lists all available buckets
func (s *S3Client) ListBuckets(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var names []string
	for _, bucket := range buckets {
		names = append(names, bucket.Name)
	}

	return names, nil
}
