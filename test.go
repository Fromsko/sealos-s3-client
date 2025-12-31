package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TestConnection tests the S3 connection with detailed diagnostics
func TestConnection() {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Load config from environment
	config := LoadConfig()

	endpoints := []struct {
		name string
		url  string
	}{
		{"Internal", "object-storage.objectstorage-system.svc.cluster.local"},
		{"External (API)", "objectstorageapi.hzh.sealos.run"},
	}

	accessKey := config.AccessKey
	secretKey := config.SecretKey
	bucket := config.Bucket

	for _, ep := range endpoints {
		fmt.Printf("\n=== Testing %s Endpoint ===\n", ep.name)
		fmt.Printf("Endpoint: %s\n", ep.url)

		client, err := minio.New(ep.url, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: true,
		})

		if err != nil {
			logger.Printf("✗ Failed to create client: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			logger.Printf("✗ Failed to check bucket: %v", err)
			continue
		}

		if exists {
			logger.Printf("✓ Bucket '%s' exists", bucket)

			// List objects
			opts := minio.ListObjectsOptions{MaxKeys: 5}
			count := 0
			for object := range client.ListObjects(ctx, bucket, opts) {
				if object.Err != nil {
					logger.Printf("✗ Error listing objects: %v", object.Err)
					break
				}
				fmt.Printf("  - %s (%d bytes)\n", object.Key, object.Size)
				count++
			}
			logger.Printf("✓ Listed %d objects", count)
		} else {
			logger.Printf("✗ Bucket '%s' does not exist", bucket)
		}
	}
}
