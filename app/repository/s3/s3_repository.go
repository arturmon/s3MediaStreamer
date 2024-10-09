package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/minio/minio-go/v7"
)

type RepositoryInterface interface {
	UploadFilesS3(ctx context.Context, upload *model.UploadS3) error
	DownloadFilesS3(ctx context.Context, name string) (string, error)
	ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error)
	DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error
	FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error)
	DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error
	CleanTemplateFile(fileName string) error
	OpenTemplateFile(fileName string) (*os.File, error)
	Ping(ctx context.Context) error
}

type Repository struct {
	cfg      *model.Config
	logger   *logs.Logger
	s3Client *minio.Client
}

func NewS3Repository(cfg *model.Config, logger *logs.Logger, client *minio.Client) *Repository {
	logger.Info("Starting S3 repository...")
	return &Repository{
		cfg:      cfg,
		logger:   logger,
		s3Client: client,
	}
}

func (h *Repository) UploadFilesS3(ctx context.Context, upload *model.UploadS3) error {
	info, err := h.s3Client.FPutObject(
		ctx, h.cfg.AppConfig.S3.BucketName,
		upload.ObjectName,
		upload.FilePath,
		minio.PutObjectOptions{ContentType: upload.ContentType})
	if err != nil {
		h.logger.Fatal(err.Error())
	}

	h.logger.Infof("Successfully uploaded %s of size %d\n", upload.ObjectName, info.Size)
	return nil
}

func (h *Repository) DownloadFilesS3(ctx context.Context, name string) (string, error) {
	// Extract the object name after the last "/"
	objectName := filepath.Base(name)
	tempDir := os.TempDir()
	h.logger.Debug("Temporary directory: ", tempDir)
	fullFilePath := tempDir + "/" + objectName

	err := h.s3Client.FGetObject(ctx, h.cfg.AppConfig.S3.BucketName, objectName, fullFilePath, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	return fullFilePath, nil
}

func (h *Repository) ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error) {
	opts := minio.ListObjectsOptions{
		Recursive:    true,
		WithMetadata: true,
		WithVersions: true,
	}
	var objects []minio.ObjectInfo

	for object := range h.s3Client.ListObjects(ctx, h.cfg.AppConfig.S3.BucketName, opts) {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func (h *Repository) DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
		VersionID:        object.VersionID,
	}
	err := h.s3Client.RemoveObject(ctx, h.cfg.AppConfig.S3.BucketName, object.Key, opts)
	if err != nil {
		return err
	}

	return nil
}

func (h *Repository) FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error) {
	objects, err := h.ListObjectS3(ctx)
	if err != nil {
		h.logger.Errorf("Error listing objects from S3: %s", err.Error())

		return minio.ObjectInfo{}, err
	}

	for _, object := range objects {
		if object.VersionID == s3tag {
			return object, nil
		}
	}

	// Object not found, return an error
	h.logger.Errorf("Object with version %s not found", s3tag)
	return minio.ObjectInfo{}, fmt.Errorf("object not found")
}

func (h *Repository) DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error {
	// Extract the object name after the last "/"
	objectName := filepath.Base(name)

	object, err := h.s3Client.GetObject(ctx, h.cfg.AppConfig.S3.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer func(object *minio.Object) {
		err = object.Close()
		if err != nil {

		}
	}(object)

	return callback(object)
}

func (h *Repository) CleanTemplateFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("error deleting file %s: %w", fileName, err)
	}
	h.logger.Debugf("File %s deleted successfully\n", fileName)
	return nil
}

func (h *Repository) OpenTemplateFile(fileName string) (*os.File, error) {
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

func (h *Repository) Ping(ctx context.Context) error {
	_, err := h.s3Client.ListBuckets(ctx)
	return err
}
