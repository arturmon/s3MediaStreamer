package s3

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type HandlerS3 interface {
	UploadFilesS3(ctx context.Context, upload *UploadS3) error
	DownloadFilesS3(ctx context.Context, name string) ([]byte, error)
	ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error)
	DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error
	FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error)
}
