package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// UploadFile uploads a file to S3 with retry logic
func (s *S3Client) UploadFile(ctx context.Context, objectName, filePath string) error {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		info, err := s.uploadFileOnce(ctx, objectName, filePath)

		if err == nil {
			s.logger.Printf("✓ Uploaded %s (size: %d bytes)", objectName, info.Size)
			return nil
		}

		if !isRetryable(err) {
			return fmt.Errorf("upload failed (non-retryable): %w", err)
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			s.logger.Printf("⚠ Retry attempt %d after %v (error: %v)", attempt+1, backoff, err)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("upload failed after %d retries", maxRetries)
}

func (s *S3Client) uploadFileOnce(ctx context.Context, objectName, filePath string) (minio.UploadInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to stat file: %w", err)
	}

	return s.client.PutObject(ctx, s.bucket, objectName, file, fileInfo.Size(),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
}

// DownloadFile downloads a file from S3 with retry logic
func (s *S3Client) DownloadFile(ctx context.Context, objectName, filePath string) error {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := s.downloadFileOnce(ctx, objectName, filePath)

		if err == nil {
			s.logger.Printf("✓ Downloaded %s to %s", objectName, filePath)
			return nil
		}

		if !isRetryable(err) {
			return fmt.Errorf("download failed (non-retryable): %w", err)
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			s.logger.Printf("⚠ Retry attempt %d after %v (error: %v)", attempt+1, backoff, err)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("download failed after %d retries", maxRetries)
}

func (s *S3Client) downloadFileOnce(ctx context.Context, objectName, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.client.FGetObject(ctx, s.bucket, objectName, filePath, minio.GetObjectOptions{})
}

// ListObjects lists all objects in the bucket with optional prefix
func (s *S3Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var objects []string
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	for object := range s.client.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("list failed: %w", object.Err)
		}
		objects = append(objects, object.Key)
	}

	s.logger.Printf("✓ Listed %d objects with prefix '%s'", len(objects), prefix)
	return objects, nil
}

// DeleteObject deletes an object from S3
func (s *S3Client) DeleteObject(ctx context.Context, objectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	s.logger.Printf("✓ Deleted object: %s", objectName)
	return nil
}

// GetObject retrieves an object as a reader
func (s *S3Client) GetObject(ctx context.Context, objectName string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	object, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return object, nil
}

// GeneratePresignedURL generates a presigned URL for downloading
func (s *S3Client) GeneratePresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	s.logger.Printf("✓ Generated presigned URL for %s (expires in %v)", objectName, expiry)
	return url.String(), nil
}

// isRetryable determines if an error is retryable
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"i/o timeout",
		"temporary failure",
		"service unavailable",
		"connection refused",
		"broken pipe",
		"eof",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}
