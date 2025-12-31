# Sealos S3 Integration Client

English | [中文](README_ZH.md)

A production-grade Go S3 integration client for interacting with Sealos Object Storage service.

## Features

- ✅ Full S3 operations support (upload, download, list, delete)
- ✅ Automatic retry mechanism (exponential backoff)
- ✅ Concurrent operations support (with rate limiting)
- ✅ Presigned URL generation
- ✅ Production-grade error handling
- ✅ Detailed logging
- ✅ Connection pool optimization

## Quick Start

### Install Dependencies

```bash
go mod download
```

### Basic Usage

```bash
# Run demo
go run . -cmd demo

# Upload file
go run . -cmd upload -object myfile.txt -file ./myfile.txt

# Download file
go run . -cmd download -object myfile.txt -file ./downloaded.txt

# List objects
go run . -cmd list -prefix "documents/"

# Delete object
go run . -cmd delete -object myfile.txt

# Generate presigned URL
go run . -cmd presigned -object myfile.txt
```

## Configuration

### Command Line Arguments

```bash
-endpoint string      S3 endpoint
-key string          Access key
-secret string       Secret key
-bucket string       Bucket name
-object string       Object name
-file string         File path
-prefix string       List prefix
-external            Use external endpoint
```

## Endpoints

- **Internal Endpoint** (Recommended): `object-storage.objectstorage-system.svc.cluster.local`

  - For applications within Sealos platform
  - No traffic charges
  - Low latency

- **External Endpoint**: `objectstorageapi.hzh.sealos.run`
  - For external network access
  - May incur traffic charges

## Project Structure

```
sealos-s3-client/
├── main.go           # Main program and CLI interface
├── config.go         # Configuration management
├── client.go         # S3 client core
├── operations.go     # Basic operations (upload, download, etc.)
├── concurrent.go     # Concurrent operations
├── go.mod            # Go module definition
└── README.md         # This file
```

## Code Examples

### Initialize Client

```go
package main

import (
    "context"
    "log"
    "os"
)

func main() {
    logger := log.New(os.Stdout, "[S3] ", log.LstdFlags)

    config := LoadConfig(
        "object-storage.objectstorage-system.svc.cluster.local",
        "your-access-key",
        "your-secret-key",
        "your-bucket-name",
    )

    client, err := NewS3Client(config, logger)
    if err != nil {
        logger.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    // Use client for operations
}
```

### Upload File

```go
err := client.UploadFile(ctx, "myfile.txt", "./local-file.txt")
if err != nil {
    logger.Fatalf("Upload failed: %v", err)
}
```

### Download File

```go
err := client.DownloadFile(ctx, "myfile.txt", "./downloaded.txt")
if err != nil {
    logger.Fatalf("Download failed: %v", err)
}
```

### List Objects

```go
objects, err := client.ListObjects(ctx, "prefix/")
if err != nil {
    logger.Fatalf("List failed: %v", err)
}

for _, obj := range objects {
    println(obj)
}
```

### Concurrent Upload

```go
files := map[string]string{
    "file1.txt": "./local1.txt",
    "file2.txt": "./local2.txt",
    "file3.txt": "./local3.txt",
}

results := client.UploadFilesParallel(ctx, files, 10)
client.PrintResults(results)
```

### Generate Presigned URL

```go
import "time"

url, err := client.GeneratePresignedURL(ctx, "myfile.txt", 24*time.Hour)
if err != nil {
    logger.Fatalf("Failed to generate URL: %v", err)
}

println("Download URL:", url)
```

## Error Handling

The client automatically handles the following errors:

- **Network Errors**: Connection timeout, connection reset, etc. (auto-retry)
- **Temporary Errors**: Service temporarily unavailable (auto-retry)
- **Permanent Errors**: Insufficient permissions, object not found, etc. (immediate return)

Retry Strategy:

- Maximum 3 retries
- Exponential backoff: 1s, 2s, 4s

## Performance Optimization

### Connection Pool

The client automatically configures connection pool:

- Max idle connections: 100
- Max idle connections per host: 10
- Idle connection timeout: 90 seconds

### Concurrency Limit

Concurrent operations default to 10 concurrent tasks, adjustable via parameter:

```go
results := client.UploadFilesParallel(ctx, files, 20) // 20 concurrent
```

### Large File Handling

minio-go automatically handles large file multipart uploads:

- Files > 64MB are automatically segmented
- Configurable segment size and concurrency

## Best Practices

1. **Credential Management**

   - Never hardcode credentials
   - Use environment variables or config files
   - Rotate access keys regularly

2. **Error Handling**

   - Always check returned errors
   - Distinguish between temporary and permanent errors
   - Log detailed error information

3. **Resource Management**

   - Close files and connections promptly
   - Use defer to ensure resource release
   - Monitor memory usage

4. **Performance**

   - Use internal endpoint to avoid traffic charges
   - Set reasonable concurrency for concurrent operations
   - Monitor S3 operation latency

5. **Security**
   - Use presigned URLs instead of exposing credentials
   - Verify uploaded file integrity
   - Regularly review access logs

## Troubleshooting

### Connection Failed

```
failed to verify S3 connection: failed to check bucket existence
```

**Solution**:

- Check if endpoint is correct
- Check network connection
- Verify credentials are correct

### Permission Error

```
upload failed: AccessDenied
```

**Solution**:

- Verify Access Key and Secret Key
- Check Bucket permissions
- Ensure credentials have sufficient permissions

### Timeout Error

```
upload failed: i/o timeout
```

**Solution**:

- Check network connection
- Increase timeout duration
- Check file size

## License

MIT

## References

- [Sealos Official Documentation](https://docs.sealos.io)
- [MinIO Go SDK](https://github.com/minio/minio-go)
- [AWS S3 API Reference](https://docs.aws.amazon.com/s3/)
