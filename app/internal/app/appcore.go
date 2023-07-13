package app

import (
	"context"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/interfaces"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"skeleton-golange-application/app/pkg/web/gin"
)

type App struct {
	cfg     *config.Config
	logger  *logging.Logger
	storage *model.DBConfig
	Gin     *gin.WebApp
}

func NewAppInit(cfg *config.Config, logger *logging.Logger) (*App, error) {
	logger.Info("logger initializing")
	logger.Info("Starting the service...")
	logger.Info("config initializing")

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

	return &App{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
		Gin:     myGin,
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

var _ interfaces.AppInterface = &App{}
