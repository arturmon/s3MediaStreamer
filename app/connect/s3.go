package connect

import (
	"context"
	"errors"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClientS3(ctx context.Context, cfg *model.Config, logger *logs.Logger) (*minio.Client, error) {
	logger.Info("Starting S3 Connection...")
	// Check that AccessKeyID and SecretAccessKey are not empty
	if cfg.AppConfig.S3.AccessKeyID == "" || cfg.AppConfig.S3.SecretAccessKey == "" {
		return nil, errors.New("AccessKeyID or SecretAccessKey is empty")
	}

	minioClient, err := minio.New(cfg.AppConfig.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AppConfig.S3.AccessKeyID, cfg.AppConfig.S3.SecretAccessKey, ""),
		Secure: cfg.AppConfig.S3.UseSSL,
		Region: cfg.AppConfig.S3.Location,
	})
	if err != nil {
		logger.Fatalln(err)
		return nil, err
	}

	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logger.Fatalln("Failed to list buckets:", err)
	}

	logger.Printf("S3 %v connected.\n", cfg.AppConfig.S3.Endpoint)

	return minioClient, nil
}
