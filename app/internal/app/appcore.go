package app

import (
	"context"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/interfaces"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"skeleton-golange-application/app/pkg/web/gin"
)

type App struct {
	cfg        *config.Config
	logger     *logging.Logger
	storage    *model.DBConfig
	Gin        *gin.WebApp
	amqpClient *amqp.MessageClient
}

func NewAppInit(cfg *config.Config, logger *logging.Logger) (*App, error) {
	logger.Info("Starting initialize the storage...")
	storage, err := model.NewDBConfig(cfg)
	if err != nil {
		logger.Error("Failed to initialize the storage:", err)
		return nil, err
	}

	ctx := context.Background()
	go monitoring.PingStorage(ctx, storage.Operations)

	myGin, err := gin.NewAppUseGin(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize Gin:", err)
		logger.Fatal(err)
		return nil, err
	}

	// Create an AMQP client
	amqpClient, err := amqp.NewAMQPClient(cfg.MessageQueue.SubQueueName, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize MQ:", err)
		logger.Fatal(err)
		return nil, err
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		storage:    storage,
		Gin:        myGin,
		amqpClient: amqpClient,
	}, nil
}

func (a *App) GetStorage() (*model.DBConfig, error) {
	return a.storage, nil
}

func (a *App) GetLogger() *logging.Logger {
	return a.logger
}
func (a *App) GetCfg() *config.Config {
	return a.cfg
}

func (a *App) GetGin() (*gin.WebApp, error) {
	return a.Gin, nil
}

func (a *App) GetMessageClient() *amqp.MessageClient {
	return a.amqpClient
}

var _ interfaces.AppInterface = &App{}
