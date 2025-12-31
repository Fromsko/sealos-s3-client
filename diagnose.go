package main

import (
	"fmt"
	"os"
)

// DiagnoseCredentials provides detailed credential diagnostics
func DiagnoseCredentials() {
	fmt.Println("\n=== Credential Diagnostics ===")

	// Check environment variables
	fmt.Println("1. Environment Variables:")
	envVars := map[string]string{
		"S3_ENDPOINT":   os.Getenv("S3_ENDPOINT"),
		"S3_ACCESS_KEY": os.Getenv("S3_ACCESS_KEY"),
		"S3_SECRET_KEY": os.Getenv("S3_SECRET_KEY"),
		"S3_BUCKET":     os.Getenv("S3_BUCKET"),
	}

	for key, value := range envVars {
		if value == "" {
			fmt.Printf("  %s: (not set)\n", key)
		} else {
			fmt.Printf("  %s: %s\n", key, maskSecret(value))
		}
	}

	// Current credentials from config
	fmt.Println("\n2. Loaded Configuration:")
	config := LoadConfig()
	fmt.Printf("  Endpoint:   %s\n", config.Endpoint)
	fmt.Printf("  Access Key: %s\n", maskSecret(config.AccessKey))
	fmt.Printf("  Secret Key: %s\n", maskSecret(config.SecretKey))
	fmt.Printf("  Bucket:     %s\n", config.Bucket)

	// Endpoint information
	fmt.Println("\n3. Available Endpoints:")
	endpoints := map[string]string{
		"Internal":       "object-storage.objectstorage-system.svc.cluster.local",
		"External (API)": "objectstorageapi.hzh.sealos.run",
	}

	for name, url := range endpoints {
		fmt.Printf("  %s: %s\n", name, url)
	}

	// Recommendations
	fmt.Println("\n4. Recommendations:")
	fmt.Println("  • If using internal endpoint: ensure you're running inside Sealos cluster")
	fmt.Println("  • If using external endpoint: check if credentials are correct")
	fmt.Println("  • The static host endpoint might require different credentials")
	fmt.Println("  • Try: go run . -cmd test")
	fmt.Println("  • Or: go run . -cmd upload -object test.txt -file ./test.txt -external")

	// Troubleshooting
	fmt.Println("\n5. Troubleshooting:")
	fmt.Println("  Access Denied:")
	fmt.Println("    → Check if Access Key and Secret Key are correct")
	fmt.Println("    → Verify bucket name exists")
	fmt.Println("    → Check if credentials have permission to access bucket")
	fmt.Println("")
	fmt.Println("  Signature mismatch:")
	fmt.Println("    → Endpoint might use different authentication method")
	fmt.Println("    → Try different endpoint")
	fmt.Println("    → Check if credentials are for this specific endpoint")
	fmt.Println("")
	fmt.Println("  Connection refused:")
	fmt.Println("    → Endpoint is unreachable")
	fmt.Println("    → Check network connectivity")
	fmt.Println("    → Verify endpoint URL is correct")
}
