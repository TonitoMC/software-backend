package s3

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service interface {
	UploadFile(key string, file *multipart.FileHeader) error
	GeneratePresignedURL(key string, expiry time.Duration) (string, error)
	GetObject(key string) (io.ReadCloser, error) // Add this line
}

type s3Service struct {
	client *s3.Client
	bucket string
}

// NewS3Service now takes an *S3Config as an argument.
func NewS3Service(cfg *S3Config) S3Service {
	// Use the values from the provided S3Config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"", // Session token is not used for static credentials
		)),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" { // Use cfg.Endpoint
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO
		}
	})

	return &s3Service{
		client: client,
		bucket: cfg.BucketName, // Use cfg.BucketName
	}
}

func (s *s3Service) UploadFile(key string, file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   src,
	})
	return err
}

func (s *s3Service) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)

	request, err := presigner.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (s *s3Service) GetObject(key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
