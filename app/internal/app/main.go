package app

import (
	"context"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/interfaces"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"skeleton-golange-application/app/pkg/s3"
	"skeleton-golange-application/app/pkg/web/gin"
)

// App represents the main application struct.
type App struct {
	Cfg        *config.Config
	Logger     *logging.Logger
	Storage    *model.DBConfig
	Gin        *gin.WebApp
	AMQPClient *amqp.MessageClient
	S3         s3.HandlerS3
}

// NewAppInit initializes a new App instance.
func NewAppInit(cfg *config.Config, logger *logging.Logger) (*App, error) {
	healthMetrics := monitoring.NewHealthMetrics()
	// Initialize the database storage.
	logger.Info("Starting initialize the storage...")
	storage, err := model.NewDBConfig(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize the storage:", err)
		return nil, err
	}

	// Initialize DBOperations interface within the storage.
	err = storage.Operations.Connect(logger) // Initialize the storage's Operations field
	if err != nil {
		logger.Error("Failed to connect to the database:", err)
		return nil, err
	}

	ctx := context.Background()
	// Start monitoring the database storage with a ticker
	go func() {
		// Call PingStorage in a goroutine
		monitoring.PingStorage(ctx, storage.Operations, healthMetrics)
	}()

	logger.Info("Starting initialize the Gin...")
	// Initialize the Gin web framework.
	myGin, err := gin.NewAppUseGin(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize Gin:", err)
		logger.Fatal(err)
		return nil, err
	}
	s3client, s3err := s3.NewClientS3(ctx, cfg, logger)
	if s3err != nil {
		logger.Error("Failed to initialize S3:", s3err)
		logger.Fatal(s3err)
		return nil, s3err
	}

	// Create an AMQP client if it's enabled in the configuration.
	var amqpClient *amqp.MessageClient
	logger.Info("Starting initialize the amqp...")
	amqpClient, err = amqp.NewAMQPClient(cfg.MessageQueue.SubQueueName, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize MQ:", err)
	}

	// Return a new App instance with all initialized components.
	return &App{
		Cfg:        cfg,
		Logger:     logger,
		Storage:    storage,
		Gin:        myGin,
		AMQPClient: amqpClient,
		S3:         s3client,
	}, nil
}

// GetStorage returns the initialized database storage instance.
func (a *App) GetStorage() (*model.DBConfig, error) {
	return a.Storage, nil
}

// GetLogger returns the logger instance used in the application.
func (a *App) GetLogger() *logging.Logger {
	return a.Logger
}

// GetCfg returns the application's configuration.
func (a *App) GetCfg() *config.Config {
	return a.Cfg
}

// GetGin returns the initialized Gin web framework instance.
func (a *App) GetGin() (*gin.WebApp, error) {
	return a.Gin, nil
}

// GetMessageClient returns the initialized AMQP client instance.
func (a *App) GetMessageClient() *amqp.MessageClient {
	return a.AMQPClient
}

func (a *App) GetS3Client() (s3.HandlerS3, error) {
	return a.S3, nil
}

// Ensure that App implements the interfaces.AppInterface interface.
var _ interfaces.AppInterface = &App{}
