package app

import (
	"context"
	"s3MediaStreamer/app/inits"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/gin-gonic/gin"
)

// App represents the main application struct.
type App struct {
	Cfg     *model.Config
	Logger  *logs.Logger
	REST    *gin.Engine
	Service *inits.Service
	AppName string
}

// NewAppInit initializes a new App instance.
func NewAppInit(ctx context.Context, cfg *model.Config, logger *logs.Logger, appName, version string) (*App, error) {
	myGin := initializeGin(ctx, cfg, logger)

	logger.Info("Start init Services ...")

	service, err := inits.InitServices(ctx, appName, version, cfg, logger)

	if err != nil {
		return nil, err
	}

	return &App{
		Cfg:     cfg,
		Logger:  logger,
		REST:    myGin,
		Service: service,
		AppName: appName,
	}, nil
}
