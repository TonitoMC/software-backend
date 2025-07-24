package s3

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service interface {
	UploadFile(key string, file *multipart.FileHeader) error
	GeneratePresignedURL(key string, expiry time.Duration) (string, error)
}

type s3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service() S3Service {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"")),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // Required for MinIO
		}
	})

	return &s3Service{
		client: client,
		bucket: os.Getenv("S3_BUCKET"),
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
