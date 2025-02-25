package rabbitmq

import (
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/s3"
	"s3MediaStreamer/app/services/tags"
	"s3MediaStreamer/app/services/track"
)

type Repository interface{}

type Service struct {
	cfg     *model.Config
	logger  *logs.Logger
	storage db.Repository
	s3      s3.Service
	track   track.Service
	tags    tags.Service
}

func NewMessageService(cfg *model.Config,
	logger *logs.Logger,
	storage db.Repository,
	s3 s3.Service,
	track track.Service,
	tags tags.Service,
) *Service {
	return &Service{
		cfg,
		logger,
		storage,
		s3,
		track,
		tags,
	}
}
