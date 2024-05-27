package s3

import (
	"context"
	"io"
	"os"
	"s3MediaStreamer/app/model"

	"github.com/minio/minio-go/v7"
)

type S3Repository interface {
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

type S3DBRepository interface {
	GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error)
	AddS3Version(ctx context.Context, trackID, version string) error
	DeleteS3Version(ctx context.Context, version string) error
}

type S3Service struct {
	s3Repository   S3Repository
	s3DBRepository S3DBRepository
}

func NewS3Service(fileRepository S3Repository, s3DBRepository S3DBRepository) *S3Service {
	return &S3Service{
		s3Repository:   fileRepository,
		s3DBRepository: s3DBRepository,
	}
}

func (s *S3Service) UploadFilesS3(ctx context.Context, upload *model.UploadS3) error {
	return s.s3Repository.UploadFilesS3(ctx, upload)
}

func (s *S3Service) DownloadFilesS3(ctx context.Context, name string) (string, error) {
	return s.s3Repository.DownloadFilesS3(ctx, name)
}

func (s *S3Service) ListObjectS3(ctx context.Context) ([]minio.ObjectInfo, error) {
	return s.s3Repository.ListObjectS3(ctx)
}

func (s *S3Service) DeleteObjectS3(ctx context.Context, object *minio.ObjectInfo) error {
	return s.s3Repository.DeleteObjectS3(ctx, object)
}

func (s *S3Service) FindObjectFromVersion(ctx context.Context, s3tag string) (minio.ObjectInfo, error) {
	return s.s3Repository.FindObjectFromVersion(ctx, s3tag)
}

func (s *S3Service) DownloadFilesS3Stream(ctx context.Context, name string, callback func(io.Reader) error) error {
	return s.s3Repository.DownloadFilesS3Stream(ctx, name, callback)
}
func (s *S3Service) CleanTemplateFile(fileName string) error {
	return s.s3Repository.CleanTemplateFile(fileName)
}

func (s *S3Service) OpenTemplateFile(fileName string) (*os.File, error) {
	return s.s3Repository.OpenTemplateFile(fileName)
}

func (s *S3Service) Ping(ctx context.Context) error {
	return s.s3Repository.Ping(ctx)
}

func (s *S3Service) GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error) {
	return s.s3DBRepository.GetS3VersionByTrackID(ctx, trackID)
}
func (s *S3Service) AddS3Version(ctx context.Context, trackID, version string) error {
	return s.s3DBRepository.AddS3Version(ctx, trackID, version)
}
func (s *S3Service) DeleteS3Version(ctx context.Context, version string) error {
	return s.s3DBRepository.DeleteS3Version(ctx, version)
}
