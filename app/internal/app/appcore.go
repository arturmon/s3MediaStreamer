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

// App represents the main application struct.
type App struct {
	cfg        *config.Config
	logger     *logging.Logger
	storage    *model.DBConfig
	Gin        *gin.WebApp
	amqpClient *amqp.MessageClient
}

// NewAppInit initializes a new App instance.
func NewAppInit(cfg *config.Config, logger *logging.Logger) (*App, error) {
	// Initialize the database storage
	logger.Info("Starting initialize the storage...")
	storage, err := model.NewDBConfig(cfg)
	if err != nil {
		logger.Error("Failed to initialize the storage:", err)
		return nil, err
	}

	ctx := context.Background()
	// Start monitoring the database storage
	go monitoring.PingStorage(ctx, storage.Operations)

	// Initialize the Gin web framework
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

	// Return a new App instance with all initialized components
	return &App{
		cfg:        cfg,
		logger:     logger,
		storage:    storage,
		Gin:        myGin,
		amqpClient: amqpClient,
	}, nil
}

// GetStorage returns the initialized database storage instance.
func (a *App) GetStorage() (*model.DBConfig, error) {
	return a.storage, nil
}

// GetLogger returns the logger instance used in the application.
func (a *App) GetLogger() *logging.Logger {
	return a.logger
}

// GetCfg returns the application's configuration.
func (a *App) GetCfg() *config.Config {
	return a.cfg
}

// GetGin returns the initialized Gin web framework instance.
func (a *App) GetGin() (*gin.WebApp, error) {
	return a.Gin, nil
}

// GetMessageClient returns the initialized AMQP client instance.
func (a *App) GetMessageClient() *amqp.MessageClient {
	return a.amqpClient
}

// Ensure that App implements the interfaces.AppInterface interface.
var _ interfaces.AppInterface = &App{}
