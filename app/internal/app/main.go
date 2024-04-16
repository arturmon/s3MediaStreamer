package app

import (
	"context"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/amqp"
	"s3MediaStreamer/app/pkg/client/repository"
	"s3MediaStreamer/app/pkg/interfaces"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/pkg/otel"
	"s3MediaStreamer/app/pkg/s3"
	"s3MediaStreamer/app/pkg/web/gin"
	"s3MediaStreamer/app/services"
	"s3MediaStreamer/app/services/consul_election"
	"s3MediaStreamer/app/services/consul_service"
)

// App represents the main application struct.
type App struct {
	Cfg            *config.Config
	Logger         *logging.Logger
	Storage        *repository.DBConfig
	Gin            *gin.WebApp
	AMQPClient     *amqp.MessageClient
	S3             s3.HandlerS3
	LeaderElection consul_election.ConsulElection
	ConsulService  consul_service.ConsulService
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

	s, leaderElection, err := services.InitServices(appName, cfg, logger)
	if err != nil {
		return nil, err
	}

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
func (a *App) GetStorage() (*repository.DBConfig, error) {
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
