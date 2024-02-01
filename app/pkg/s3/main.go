package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClientS3(cfg *config.Config, logger *logging.Logger) (*HandlerFromS3, error) {
	logger.Info("S3 initializing...")
	minioClient, err := minio.New(cfg.AppConfig.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AppConfig.S3.AccessKeyID, cfg.AppConfig.S3.SecretAccessKey, ""),
		Secure: cfg.AppConfig.S3.UseSSL,
		Region: cfg.AppConfig.S3.Location,
	})
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("S3 %v connected.\n", cfg.AppConfig.S3.Endpoint)

	return &HandlerFromS3{
		cfg:       cfg,
		logger:    logger,
		s3Handler: minioClient,
	}, nil
}

func (h *HandlerFromS3) UploadFilesS3(ctx context.Context, upload *UploadS3) error {
	info, err := h.s3Handler.FPutObject(
		ctx, h.cfg.AppConfig.S3.BucketName,
		upload.ObjectName,
		upload.FilePath,
		minio.PutObjectOptions{ContentType: upload.ContentType})
	if err != nil {
		h.logger.Fatalln(err)
	}

	h.logger.Printf("Successfully uploaded %s of size %d\n", upload.ObjectName, info.Size)
	return nil
}

func (h *HandlerFromS3) DownloadFilesS3(ctx context.Context, name string) ([]byte, error) {
	// Extract the object name after the last "/"
	objectName := filepath.Base(name)

	object, err := h.s3Handler.GetObject(ctx, h.cfg.AppConfig.S3.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, object); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (h *HandlerFromS3) ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error) {
	opts := minio.ListObjectsOptions{
		Recursive:    true,
		WithMetadata: true,
		WithVersions: true,
	}
	var objects []minio.ObjectInfo

	for object := range h.s3Handler.ListObjects(ctx, h.cfg.AppConfig.S3.BucketName, opts) {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func (h *HandlerFromS3) DeleteObjectS3(ctx context.Context, object string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	err := h.s3Handler.RemoveObject(ctx, h.cfg.AppConfig.S3.BucketName, object, opts)
	if err != nil {
		return err
	}
	return nil
}

func (h *HandlerFromS3) FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error) {
	objects, err := h.ListObjectS3(ctx)
	if err != nil {
		return minio.ObjectInfo{}, err
	}

	for _, object := range objects {
		if object.VersionID == s3tag {
			return object, nil
		}
	}

	// Object not found, return an error
	return minio.ObjectInfo{}, fmt.Errorf("object not found")
}
