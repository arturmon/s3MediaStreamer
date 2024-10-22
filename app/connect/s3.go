package connect

import (
	"context"
	"errors"
	"os"
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
	// Create logs.LoggerMessageConnect
	logFields := []model.LogField{
		{Key: "TypeConnect", Value: "S3", Mask: ""},
		{Key: "AccessKeyID", Value: cfg.AppConfig.S3.AccessKeyID, Mask: ""},
		{Key: "Addr", Value: cfg.AppConfig.S3.Endpoint, Mask: ""},
		{Key: "SecretAccessKey", Value: cfg.AppConfig.S3.SecretAccessKey, Mask: "password"},
		{Key: "Region", Value: cfg.AppConfig.S3.Location, Mask: ""},
		{Key: "Secure", Value: cfg.AppConfig.S3.UseSSL, Mask: ""},
	}
	loggerMsg := logs.NewLoggerMessageConnect(logFields)

	if err != nil {
		logger.Slog().Error("(S3) Failed create MinIO client", "connection", loggerMsg.MaskFields())
		os.Exit(1)
		return nil, err
	}

	// Verify connection by listing buckets
	logger.Info("Verifying S3 connection by listing buckets...")
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logger.Slog().Error("(S3) Failed to list S3 bucket", "connection", loggerMsg.MaskFields())
		return nil, err
	}

	logger.Slog().Info("(S3) Successfully to connect", "connection", loggerMsg.MaskFields())
	return minioClient, nil
}
