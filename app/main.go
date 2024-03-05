package main

import (
	"context"
	_ "net/http/pprof"
	_ "s3MediaStreamer/app/docs"
	"s3MediaStreamer/app/internal/app"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/internal/jobs"
	"s3MediaStreamer/app/pkg/logging"
	_ "s3MediaStreamer/app/pkg/web/gin"

	_ "github.com/joho/godotenv/autoload"
)

// @title			S3 Media Streamer Application API
// @version		0.0.1
// @description	This is a s3 media streamer server.
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http https
// @host      s3streammedia.com
// @BasePath	/v1
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @securityDefinitions.basic	BasicAuth
// @authorizationurl http://s3streammedia.com/v1/users/login
func main() {
	// debug.SetMemoryLimit(2048)
	version := "0.0.1"
	buildTime := "0000-00-00 UTC"
	appName := "s3MediaStreamer"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.GetConfig()
	logger := logging.GetLogger(cfg.AppConfig.LogLevel, cfg.AppConfig.LogType, cfg.AppConfig.LogGelfServer, cfg.AppConfig.LogGelfServerType, appName)
	logger.Printf("App Version: %s Build Time: %s\n", version, buildTime)
	logger.Info("config initialize")
	logger.Info("logger initialize")
	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(&logger)
		go app.StartPprofServer(&logger)
	}
	myApp, err := app.NewAppInit(ctx, cfg, &logger, appName, version)
	if err != nil {
		logger.Error("Failed to initialize the new my app:", err)
	}

	logger.Info("Starting initialize the job runner...")
	err = jobs.InitJob(myApp)
	if err != nil {
		logger.Error("Failed to initialize the job runner:", err)
	}

	app.HandleSignals(ctx, logger, cancel)

	myApp.Start(ctx)
}
