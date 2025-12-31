package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds S3 connection configuration
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// LoadEnvFile loads environment variables from .env file
func LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	_ = LoadEnvFile(".env")

	useSSL := os.Getenv("S3_USE_SSL") != "false"

	return &Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		Bucket:    os.Getenv("S3_BUCKET"),
		UseSSL:    useSSL,
	}
}

// LoadConfigWithOverrides loads config from .env and applies CLI overrides
func LoadConfigWithOverrides(endpoint, accessKey, secretKey, bucket string) *Config {
	config := LoadConfig()

	if endpoint != "" {
		config.Endpoint = endpoint
	}
	if accessKey != "" {
		config.AccessKey = accessKey
	}
	if secretKey != "" {
		config.SecretKey = secretKey
	}
	if bucket != "" {
		config.Bucket = bucket
	}

	return config
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required (set S3_ENDPOINT)")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("access key is required (set S3_ACCESS_KEY)")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret key is required (set S3_SECRET_KEY)")
	}
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required (set S3_BUCKET)")
	}
	return nil
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
