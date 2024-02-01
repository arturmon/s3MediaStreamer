package s3

import (
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/minio/minio-go/v7"
)

type UploadS3 struct {
	ObjectName  string `json:"object_name" example:"Title name"`
	FilePath    string `json:"file_path" example:"File path"`
	ContentType string `json:"content_type" example:"Content Type"`
}

type HandlerFromS3 struct {
	cfg       *config.Config
	logger    *logging.Logger
	s3Handler *minio.Client
}
