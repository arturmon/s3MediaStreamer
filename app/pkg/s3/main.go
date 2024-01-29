package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func (h *HandlerFromS3) NewClientS3(cfg *config.Config, logger *logging.Logger) (*HandlerFromS3, error) {
	logger.Info("S3 initializing...")
	minioClient, err := minio.New(cfg.AppConfig.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AppConfig.S3.AccessKeyID, cfg.AppConfig.S3.SecretAccessKey, ""),
		Secure: cfg.AppConfig.S3.UseSSL,
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

func (h *HandlerFromS3) InitS3(ctx context.Context) error {
	err := h.s3Handler.MakeBucket(ctx, h.cfg.AppConfig.S3.BucketName, minio.MakeBucketOptions{Region: h.cfg.AppConfig.S3.Location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := h.s3Handler.BucketExists(ctx, h.cfg.AppConfig.S3.BucketName)
		if errBucketExists == nil && exists {
			h.logger.Printf("We already own %s\n", h.cfg.AppConfig.S3.BucketName)
		} else {
			h.logger.Fatalln(err)
		}
	} else {
		h.logger.Printf("Successfully created %s\n", h.cfg.AppConfig.S3.BucketName)
	}
	// Enable versioning for the S3 bucket
	err = h.s3Handler.SetBucketVersioning(ctx, h.cfg.AppConfig.S3.BucketName, minio.BucketVersioningConfiguration{
		Status: "Enabled",
	})
	if err != nil {
		h.logger.Fatalln("Error enabling versioning:", err)
	} else {
		h.logger.Printf("Successfully enabled versioning for %s\n", h.cfg.AppConfig.S3.BucketName)
	}
	return nil
}

func (h *HandlerFromS3) UploadFilesS3(ctx context.Context, upload *model.UploadS3) error {
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
	object, err := h.s3Handler.GetObject(ctx, h.cfg.AppConfig.S3.BucketName, name, minio.GetObjectOptions{})
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
