package s3

import (
	"context"
	"io"
	"os"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/minio/minio-go/v7"
)

// Handler implements the HandlerS3 interface.
type Handler struct {
	handler *HandlerFromS3
}

// NewS3Handler creates a new instance of S3Handler.
func NewS3Handler(cfg *config.Config, logger *logging.Logger) (*Handler, error) {
	s3Handler, err := NewClientS3(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	return &Handler{
		handler: s3Handler,
	}, nil
}

// UploadFilesS3 Implement the HandlerS3 interface methods using the underlying s3 package.
func (h *Handler) UploadFilesS3(ctx context.Context, upload *UploadS3) error {
	return h.handler.UploadFilesS3(ctx, upload)
}

func (h *Handler) DownloadFilesS3(ctx context.Context, name string) (string, error) {
	return h.handler.DownloadFilesS3(ctx, name)
}

func (h *Handler) ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error) {
	return h.handler.ListObjectS3(ctx)
}

func (h *Handler) DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error {
	return h.handler.DeleteObjectS3(ctx, object)
}

func (h *Handler) FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error) {
	return h.handler.FindObjectFromVersion(ctx, s3tag)
}

func (h *Handler) DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error {
	return h.handler.DownloadFilesS3Stream(ctx, name, callback)
}

func (h *Handler) CleanTemplateFile(fileName string) error {
	return h.handler.CleanTemplateFile(fileName)
}
func (h *Handler) OpenTemplateFile(fileName string) (*os.File, error) {
	return h.handler.OpenTemplateFile(fileName)
}
