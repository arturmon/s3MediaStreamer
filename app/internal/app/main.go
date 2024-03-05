package app

import (
	"context"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/client/model"
	consulelection "skeleton-golange-application/app/pkg/consulelection"
	consulservice "skeleton-golange-application/app/pkg/consulservice"
	"skeleton-golange-application/app/pkg/interfaces"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/otel"
	"skeleton-golange-application/app/pkg/s3"
	"skeleton-golange-application/app/pkg/web/gin"
)

// App represents the main application struct.
type App struct {
	Cfg            *config.Config
	Logger         *logging.Logger
	Storage        *model.DBConfig
	Gin            *gin.WebApp
	AMQPClient     *amqp.MessageClient
	S3             s3.HandlerS3
	LeaderElection *consulelection.Election
	ConsulService  *consulservice.Service
	tracer         *otel.Provider
	AppName        string
}

// NewAppInit initializes a new App instance.
func NewAppInit(ctx context.Context, cfg *config.Config, logger *logging.Logger, appName, version string) (*App, error) {
	tracer, err := initializeTracer(ctx, cfg, logger, appName, version)
	if err != nil {
		return nil, err
	}

	storage, err := initializeStorage(cfg, logger)
	if err != nil {
		return nil, err
	}

	myGin, err := initializeGin(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	s3client, err := initializeS3(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	amqpClient := initializeAMQPClient(cfg, logger)

	logger.Info("Start register consul lieder election ...")

	s := initializeConsulService(appName, cfg, logger)
	leaderElection := initializeConsulElection(appName, logger, s)
	// Return a new App instance with all initialized components.

	return &App{
		Cfg:            cfg,
		Logger:         logger,
		Storage:        storage,
		Gin:            myGin,
		AMQPClient:     amqpClient,
		S3:             s3client,
		LeaderElection: leaderElection,
		ConsulService:  s,
		tracer:         &tracer,
		AppName:        appName,
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
