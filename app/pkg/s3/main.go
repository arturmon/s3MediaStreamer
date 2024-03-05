package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/logging"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClientS3(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*HandlerFromS3, error) {
	logger.Info("S3 initializing...")

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

func (h *HandlerFromS3) DownloadFilesS3(ctx context.Context, name string) (string, error) {
	// Extract the object name after the last "/"
	objectName := filepath.Base(name)
	tempDir := os.TempDir()
	h.logger.Debugln("Temporary directory:", tempDir)
	fullFilePath := tempDir + "/" + objectName

	err := h.s3Handler.FGetObject(ctx, h.cfg.AppConfig.S3.BucketName, objectName, fullFilePath, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	return fullFilePath, nil
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

func (h *HandlerFromS3) DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
		VersionID:        object.VersionID,
	}
	err := h.s3Handler.RemoveObject(ctx, h.cfg.AppConfig.S3.BucketName, object.Key, opts)
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

func (h *HandlerFromS3) DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error {
	// Extract the object name after the last "/"
	objectName := filepath.Base(name)

	object, err := h.s3Handler.GetObject(ctx, h.cfg.AppConfig.S3.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	return callback(object)
}

func (h *HandlerFromS3) CleanTemplateFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("error deleting file %s: %w", fileName, err)
	}
	h.logger.Debugf("File %s deleted successfully\n", fileName)
	return nil
}

func (h *HandlerFromS3) OpenTemplateFile(fileName string) (*os.File, error) {
	_, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	// defer f.Close()
	return f, nil
}

func (h *HandlerFromS3) Ping(ctx context.Context) error {
	_, err := h.s3Handler.ListBuckets(ctx)
	return err
}
