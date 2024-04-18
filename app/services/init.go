package services

import (
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/services/consul_election"
	"s3MediaStreamer/app/services/consul_service"
)

func InitServices(appName string, cfg *config.Config, logger *logging.Logger) (consul_service.ConsulService, *consul_election.Election, error) {
	logger.Info("Starting initialize the consul...")
	service := consul_service.NewService(appName, cfg, logger)
	logger.Info("Register services consul...")
	service.Start()
	logger.Info("Starting initialize the consul election...")
	leaderElection := consul_election.NewElection(appName, logger, service)
	return service, leaderElection, nil
}
