package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Setup logger
	logger := log.New(os.Stdout, "[S3] ", log.LstdFlags)

	// Parse command line flags
	command := flag.String("cmd", "help", "Command: upload, download, list, delete, presigned, help")
	endpoint := flag.String("endpoint", "", "S3 endpoint (overrides S3_ENDPOINT)")
	accessKey := flag.String("key", "", "Access key (overrides S3_ACCESS_KEY)")
	secretKey := flag.String("secret", "", "Secret key (overrides S3_SECRET_KEY)")
	bucket := flag.String("bucket", "", "Bucket name (overrides S3_BUCKET)")
	object := flag.String("object", "", "Object name")
	file := flag.String("file", "", "File path")
	prefix := flag.String("prefix", "", "Prefix for list operation")
	useExternal := flag.Bool("external", false, "Use external endpoint instead of internal")

	flag.Parse()

	// Handle test command early (doesn't need client)
	if *command == "test" {
		TestConnection()
		return
	}

	// Handle diagnose command early
	if *command == "diagnose" {
		DiagnoseCredentials()
		return
	}

	// Load configuration from .env and CLI overrides
	config := LoadConfigWithOverrides(*endpoint, *accessKey, *secretKey, *bucket)

	// Handle external endpoint flag
	if *useExternal && config.Endpoint == "" {
		config.Endpoint = "objectstorageapi.hzh.sealos.run"
	}

	// Create S3 client
	client, err := NewS3Client(config, logger)
	if err != nil {
		logger.Fatalf("Failed to create S3 client: %v", err)
	}

	ctx := context.Background()

	// Execute command
	switch *command {

	case "upload":
		if *object == "" || *file == "" {
			logger.Fatal("upload requires -object and -file flags")
		}
		if err := client.UploadFile(ctx, *object, *file); err != nil {
			logger.Fatalf("Upload failed: %v", err)
		}

	case "download":
		if *object == "" || *file == "" {
			logger.Fatal("download requires -object and -file flags")
		}
		if err := client.DownloadFile(ctx, *object, *file); err != nil {
			logger.Fatalf("Download failed: %v", err)
		}

	case "list":
		objects, err := client.ListObjects(ctx, *prefix)
		if err != nil {
			logger.Fatalf("List failed: %v", err)
		}
		fmt.Println("\n=== Objects ===")
		for _, obj := range objects {
			fmt.Printf("  %s\n", obj)
		}

	case "delete":
		if *object == "" {
			logger.Fatal("delete requires -object flag")
		}
		if err := client.DeleteObject(ctx, *object); err != nil {
			logger.Fatalf("Delete failed: %v", err)
		}

	case "presigned":
		if *object == "" {
			logger.Fatal("presigned requires -object flag")
		}
		url, err := client.GeneratePresignedURL(ctx, *object, 24*time.Hour)
		if err != nil {
			logger.Fatalf("Failed to generate presigned URL: %v", err)
		}
		fmt.Printf("\nPresigned URL:\n%s\n", url)

	case "demo":
		runDemo(client, logger)

	case "help":
		printHelp()

	default:
		logger.Fatalf("Unknown command: %s", *command)
	}
}

func runDemo(client *S3Client, logger *log.Logger) {
	ctx := context.Background()

	fmt.Println("\n=== S3 Integration Demo ===")

	// 1. List buckets
	fmt.Println("1. Listing buckets...")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		logger.Printf("Failed to list buckets: %v", err)
	} else {
		for _, b := range buckets {
			fmt.Printf("  - %s\n", b)
		}
	}

	// 2. Create test file
	fmt.Println("\n2. Creating test file...")
	testFile := "test-file.txt"
	testContent := "Hello, Sealos S3! This is a test file.\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		logger.Fatalf("Failed to create test file: %v", err)
	}
	fmt.Printf("  Created: %s\n", testFile)

	// 3. Upload file
	fmt.Println("\n3. Uploading file...")
	if err := client.UploadFile(ctx, "demo/test-file.txt", testFile); err != nil {
		logger.Fatalf("Upload failed: %v", err)
	}

	// 4. List objects
	fmt.Println("\n4. Listing objects...")
	objects, err := client.ListObjects(ctx, "demo/")
	if err != nil {
		logger.Printf("Failed to list objects: %v", err)
	} else {
		for _, obj := range objects {
			fmt.Printf("  - %s\n", obj)
		}
	}

	// 5. Generate presigned URL
	fmt.Println("\n5. Generating presigned URL...")
	url, err := client.GeneratePresignedURL(ctx, "demo/test-file.txt", 1*time.Hour)
	if err != nil {
		logger.Printf("Failed to generate presigned URL: %v", err)
	} else {
		fmt.Printf("  URL: %s\n", url)
	}

	// 6. Download file
	fmt.Println("\n6. Downloading file...")
	downloadFile := "downloaded-test-file.txt"
	if err := client.DownloadFile(ctx, "demo/test-file.txt", downloadFile); err != nil {
		logger.Fatalf("Download failed: %v", err)
	}

	// 7. Verify downloaded content
	fmt.Println("\n7. Verifying downloaded content...")
	content, err := os.ReadFile(downloadFile)
	if err != nil {
		logger.Fatalf("Failed to read downloaded file: %v", err)
	}
	fmt.Printf("  Content: %s", string(content))

	// 8. Delete objects
	fmt.Println("\n8. Cleaning up...")
	if err := client.DeleteObject(ctx, "demo/test-file.txt"); err != nil {
		logger.Printf("Failed to delete object: %v", err)
	}

	// 9. Cleanup local files
	_ = os.Remove(testFile)
	_ = os.Remove(downloadFile)

	fmt.Println()
	fmt.Println("✓ Demo completed successfully!")
}

func printHelp() {
	fmt.Print(`S3 Integration Client - Usage Guide

Commands:
  diagnose    Show credential diagnostics
  test        Test S3 connection
  upload      Upload a file to S3
  download    Download a file from S3
  list        List objects in bucket
  delete      Delete an object from S3
  presigned   Generate a presigned URL
  demo        Run a complete demo
  help        Show this help message

Flags:
  -cmd string
        Command to execute (default: "help")
  -endpoint string
        S3 endpoint (overrides S3_ENDPOINT env)
  -key string
        Access key (overrides S3_ACCESS_KEY env)
  -secret string
        Secret key (overrides S3_SECRET_KEY env)
  -bucket string
        Bucket name (overrides S3_BUCKET env)
  -object string
        Object name (required for upload/download/delete/presigned)
  -file string
        File path (required for upload/download)
  -prefix string
        Prefix for list operation

Configuration:
  Create a .env file with your credentials (see .env.example)

Examples:
  # Upload a file
  go run . -cmd upload -object myfile.txt -file ./myfile.txt

  # Download a file
  go run . -cmd download -object myfile.txt -file ./downloaded.txt

  # List objects
  go run . -cmd list -prefix "documents/"

  # Delete an object
  go run . -cmd delete -object myfile.txt

  # Generate presigned URL
  go run . -cmd presigned -object myfile.txt

  # Run demo
  go run . -cmd demo
`)
}
