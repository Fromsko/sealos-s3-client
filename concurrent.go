package main

import (
	"context"
	"fmt"
	"sync"
)

// UploadResult holds the result of an upload operation
type UploadResult struct {
	ObjectName string
	FilePath   string
	Error      error
}

// UploadFilesParallel uploads multiple files concurrently with rate limiting
func (s *S3Client) UploadFilesParallel(ctx context.Context, files map[string]string, maxConcurrency int) []UploadResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	sem := make(chan struct{}, maxConcurrency)
	results := make([]UploadResult, 0, len(files))
	resultChan := make(chan UploadResult, len(files))
	var wg sync.WaitGroup

	for objectName, filePath := range files {
		wg.Add(1)
		go func(name, path string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			err := s.UploadFile(ctx, name, path)
			resultChan <- UploadResult{
				ObjectName: name,
				FilePath:   path,
				Error:      err,
			}
		}(objectName, filePath)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// DownloadResult holds the result of a download operation
type DownloadResult struct {
	ObjectName string
	FilePath   string
	Error      error
}

// DownloadFilesParallel downloads multiple files concurrently with rate limiting
func (s *S3Client) DownloadFilesParallel(ctx context.Context, files map[string]string, maxConcurrency int) []DownloadResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	sem := make(chan struct{}, maxConcurrency)
	results := make([]DownloadResult, 0, len(files))
	resultChan := make(chan DownloadResult, len(files))
	var wg sync.WaitGroup

	for objectName, filePath := range files {
		wg.Add(1)
		go func(name, path string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			err := s.DownloadFile(ctx, name, path)
			resultChan <- DownloadResult{
				ObjectName: name,
				FilePath:   path,
				Error:      err,
			}
		}(objectName, filePath)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// PrintResults prints operation results with summary
func (s *S3Client) PrintResults(results interface{}) {
	switch v := results.(type) {
	case []UploadResult:
		s.printUploadResults(v)
	case []DownloadResult:
		s.printDownloadResults(v)
	}
}

func (s *S3Client) printUploadResults(results []UploadResult) {
	successCount := 0
	failureCount := 0

	fmt.Println("\n=== Upload Results ===")
	for _, result := range results {
		if result.Error != nil {
			failureCount++
			fmt.Printf("✗ %s: %v\n", result.ObjectName, result.Error)
		} else {
			successCount++
			fmt.Printf("✓ %s\n", result.ObjectName)
		}
	}
	fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, failureCount)
}

func (s *S3Client) printDownloadResults(results []DownloadResult) {
	successCount := 0
	failureCount := 0

	fmt.Println("\n=== Download Results ===")
	for _, result := range results {
		if result.Error != nil {
			failureCount++
			fmt.Printf("✗ %s: %v\n", result.ObjectName, result.Error)
		} else {
			successCount++
			fmt.Printf("✓ %s -> %s\n", result.ObjectName, result.FilePath)
		}
	}
	fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, failureCount)
}
