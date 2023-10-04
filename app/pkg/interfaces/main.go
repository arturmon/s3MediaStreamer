package interfaces

import (
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/web/gin"
)

type AppInterface interface {
	GetStorage() (*model.DBConfig, error)
	GetLogger() *logging.Logger
	GetCfg() *config.Config
	GetGin() (*gin.WebApp, error)
	GetMessageClient() *amqp.MessageClient
}
