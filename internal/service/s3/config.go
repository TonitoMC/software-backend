package s3

import "os"

// S3Config holds configuration for S3 client
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool
}

// NewS3Config creates a new S3Config by reading from environment variables
// with default values.
func NewS3Config() *S3Config {
	return &S3Config{
		Endpoint:        getEnv("S3_ENDPOINT", "http://minio:9000"),
		AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "minioadmin"),
		SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "minioadmin"),
		BucketName:      getEnv("S3_BUCKET", "patient-docs"),
		Region:          getEnv("AWS_REGION", "us-east-1"),
		UseSSL:          getEnv("S3_USE_SSL", "false") == "true",
	}
}

// getEnv is a helper function to get environment variables with a default.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
