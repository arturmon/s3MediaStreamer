package s3

import (
	"context"
	"io"
	"os"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/logging"

	"github.com/minio/minio-go/v7"
)

type HandlerS3 interface {
	UploadFilesS3(ctx context.Context, upload *UploadS3) error
	DownloadFilesS3(ctx context.Context, name string) (string, error)
	ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error)
	DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error
	FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error)
	DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error
	CleanTemplateFile(fileName string) error
	OpenTemplateFile(fileName string) (*os.File, error)
	Ping(ctx context.Context) error
}

type HandlerFromS3 struct {
	cfg       *config.Config
	logger    *logging.Logger
	s3Handler *minio.Client
}
