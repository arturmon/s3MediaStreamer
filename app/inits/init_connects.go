package inits

import (
	"context"
	"s3MediaStreamer/app/connect"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
)

func initConnects(ctx context.Context, cfg *model.Config, logger *logs.Logger) (*initConnect, error) {
	logger.Info("Starting initialize the connect...")
	logger.Info("Starting initialize the connect cashing user auth...")
	cashingDB, err := connect.InitRedis(ctx, cfg, logger, 0)
	if err != nil && !cfg.Storage.Caching.Enabled {
		logger.Info("redis is NOT initializing or disabled !!!")
	}
	rabbitCon, err := connect.NewRabbitMQConnection(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	s3client, err := connect.NewClientS3(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	pgclient, err := connect.NewDBConfig(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	sessionclient, err := connect.InitSession(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("Complete connect initialize.")
	return &initConnect{
		cashingDB:    cashingDB,
		RabbitCon:    rabbitCon,
		s3Client:     s3client,
		pgClient:     pgclient,
		SessionStore: sessionclient,
	}, nil
}
