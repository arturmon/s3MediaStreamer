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
	logger.Info("Starting S3 connection setup...")
	// Check that AccessKeyID and SecretAccessKey are not empty
	if cfg.AppConfig.S3.AccessKeyID == "" || cfg.AppConfig.S3.SecretAccessKey == "" {
		err := errors.New("AccessKeyID or SecretAccessKey is empty")
		logger.Errorf("Configuration error: %v", err)
		return nil, err
	}

	minioClient, err := minio.New(cfg.AppConfig.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AppConfig.S3.AccessKeyID, cfg.AppConfig.S3.SecretAccessKey, ""),
		Secure: cfg.AppConfig.S3.UseSSL,
		Region: cfg.AppConfig.S3.Location,
	})
	if err != nil {
		logger.Fatalf("Failed to create MinIO client: %v", err)
		return nil, err
	}

	// Verify connection by listing buckets
	logger.Info("Verifying S3 connection by listing buckets...")
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logger.Errorf("Failed to list S3 buckets: %v", err)
		return nil, err
	}

	logger.Infof("Successfully connected to S3 endpoint: %s", cfg.AppConfig.S3.Endpoint)
	return minioClient, nil
}
