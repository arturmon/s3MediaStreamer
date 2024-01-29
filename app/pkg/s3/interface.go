package s3

import (
	"context"
	"github.com/minio/minio-go/v7"
	"skeleton-golange-application/app/internal/config"
	"sync"

	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/logging"
)

type HandlerS3 interface {
	NewClientS3(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*HandlerFromS3, error)
	InitS3(ctx context.Context) error
	UploadFilesS3(upload *model.UploadS3, ctx context.Context) error
	DownloadFilesS3(ctx context.Context, name string) ([]byte, error)
	MonitoringDirectoryS3(ctx context.Context, wg *sync.WaitGroup) (minio.ObjectInfo, error)
}

type HandlerFromS3 struct {
	cfg       *config.Config
	logger    *logging.Logger
	s3Handler *minio.Client
}
