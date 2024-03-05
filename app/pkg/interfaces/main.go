package interfaces

import (
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/amqp"
	"s3MediaStreamer/app/pkg/client/model"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/pkg/s3"
	"s3MediaStreamer/app/pkg/web/gin"
)

type AppInterface interface {
	GetStorage() (*model.DBConfig, error)
	GetLogger() *logging.Logger
	GetCfg() *config.Config
	GetGin() (*gin.WebApp, error)
	GetMessageClient() *amqp.MessageClient
	GetS3Client() (s3.HandlerS3, error)
}
