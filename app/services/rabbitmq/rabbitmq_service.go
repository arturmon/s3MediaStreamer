package rabbitmq

import (
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/s3"
	"s3MediaStreamer/app/services/tags"
	"s3MediaStreamer/app/services/track"
)

type MessageRepository interface {
}

type MessageService struct {
	logger  *logs.Logger
	storage db.DBRepository
	s3      s3.S3Service
	track   track.TrackService
	tags    tags.TagsService
}

func NewMessageService(logger *logs.Logger,
	storage db.DBRepository,
	s3 s3.S3Service,
	track track.TrackService,
	tags tags.TagsService) *MessageService {
	return &MessageService{
		logger,
		storage,
		s3,
		track,
		tags,
	}
}
