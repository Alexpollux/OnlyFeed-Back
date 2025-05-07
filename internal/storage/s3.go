package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
	"os"
)

var s3Client *s3.Client
var s3Bucket string
var s3Region string

func InitS3() error {
	s3Bucket = os.Getenv("AWS_BUCKET_NAME")
	s3Region = os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("chargement config AWS: %w", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	return nil
}

func UploadToS3(file multipart.File, filename string, contentType string, folder string) (string, error) {
	key := fmt.Sprintf("%s/%s", folder, filename)

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload échoué: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3Bucket, s3Region, key)
	return publicURL, nil
}

func DeleteFromS3(key string) error {
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("erreur suppression S3 : %w", err)
	}
	return nil
}
