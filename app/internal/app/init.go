package app

import (
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/amqp"
	"s3MediaStreamer/app/pkg/client/repository"
	"s3MediaStreamer/app/pkg/consulelection"
	"s3MediaStreamer/app/pkg/consulservice"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/pkg/otel"
	"s3MediaStreamer/app/pkg/s3"
	"s3MediaStreamer/app/pkg/web/gin"
	"time"

	"context"
)

func initializeTracer(ctx context.Context, cfg *config.Config, logger *logging.Logger, appName, version string) (otel.Provider, error) {
	config := otel.ProviderConfig{
		JaegerEndpoint: cfg.AppConfig.OpenTelemetry.JaegerEndpoint + "/api/traces",
		ServiceName:    appName,
		ServiceVersion: version,
		Environment:    cfg.AppConfig.OpenTelemetry.Environment,
		Cfg:            cfg,
		Logger:         logger,
		Disabled:       cfg.AppConfig.OpenTelemetry.TracingEnabled,
	}
	tracer, err := otel.InitProvider(ctx, config)
	if err != nil {
		return otel.Provider{}, err
	}
	return tracer, nil
}

func initializeStorage(cfg *config.Config, logger *logging.Logger) (*repository.DBConfig, error) {
	// Initialize the database storage.
	logger.Info("Starting initialize the storage...")
	storage, err := repository.NewDBConfig(cfg, logger)
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
	return storage, nil
}

func initializeGin(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*gin.WebApp, error) {
	logger.Info("Starting initialize the Gin...")
	// Initialize the Gin web framework.
	myGin, err := gin.NewAppUseGin(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize Gin:", err)
		logger.Fatal(err)
		return nil, err
	}
	return myGin, nil
}

func initializeS3(ctx context.Context, cfg *config.Config, logger *logging.Logger) (s3.HandlerS3, error) {
	s3client, s3err := s3.NewClientS3(ctx, cfg, logger)
	if s3err != nil {
		logger.Error("Failed to initialize S3:", s3err)
		logger.Fatal(s3err)
		return nil, s3err
	}
	return s3client, nil
}

func initializeAMQPClient(cfg *config.Config, logger *logging.Logger) *amqp.MessageClient {
	// Create an AMQP client if it's enabled in the configuration.
	var amqpClient *amqp.MessageClient
	var err error
	logger.Info("Starting initialize the amqp...")
	for {
		amqpClient, err = amqp.NewAMQPClient(cfg.MessageQueue.SubQueueName, cfg, logger)
		if err == nil {
			break // If successful, break out of the loop
		}

		logger.Error("Failed to initialize MQ:", err)
		time.Sleep(retryWaitTimeSeconds * time.Second) // Wait before retrying
	}
	return amqpClient
}

func initializeConsulService(appName string, cfg *config.Config, logger *logging.Logger) *consulservice.Service {
	logger.Info("Starting initialize the consul...")
	s := consulservice.NewService(appName, cfg, logger)
	logger.Info("Register service consul...")
	s.Start()
	return s
}

func initializeConsulElection(appName string, logger *logging.Logger, s *consulservice.Service) *consulelection.Election {
	leaderElection := consulelection.NewElection(appName, logger, s)
	return leaderElection
}
