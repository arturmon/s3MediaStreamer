package main

import (
	"context"
	_ "net/http/pprof"
	_ "s3MediaStreamer/app/docs"
	"s3MediaStreamer/app/handlers"
	"s3MediaStreamer/app/internal/app"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/internal/jobs"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/router"

	_ "github.com/joho/godotenv/autoload"
)

// @title			               S3 Media Streamer Application API
// @version		                   0.0.1
// @description	                   This is a s3 media streamer server.
// @contact.name                   API Support
// @contact.url                    http://www.swagger.io/support
// @contact.email                  support@swagger.io

// @license.name                   Apache 2.0
// @license.url                    http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes                        http https
// @host                           s3streammedia.localhost
// @BasePath	                   /v1
// @externalDocs.description       OpenAPI
// @externalDocs.url               https://swagger.io/resources/open-api/
// @securityDefinitions.apikey     ApiKeyAuth
// @in                             header
// @name                           Authorization
// @description                    Enter the JWT token in the format: Bearer {token}
func main() {
	// debug.SetMemoryLimit(2048)
	version := "0.0.1"
	buildTime := "0000-00-00 UTC"
	appName := "s3MediaStreamer"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.GetConfig()
	logger := logs.GetLogger(
		cfg.AppConfig.LogLevel,
		cfg.AppConfig.LogType,
		cfg.AppConfig.LogGelfServer,
		cfg.AppConfig.LogGelfServerType,
		appName,
	)
	logger.Printf("App Version: %s Build Time: %s\n", version, buildTime)
	logger.Info("config initialize")
	logger.Infof("Logger initialized with level: %s, type: %s", cfg.AppConfig.LogLevel, cfg.AppConfig.LogType)
	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(logger)
		go app.StartPprofServer(logger)
	}
	logger.Info("Initializing application...")
	myApp, err := app.NewAppInit(ctx, cfg, logger, appName, version)
	if err != nil {
		logger.Fatalf("Failed to initialize the application: %v", err)
	}
	logger.Info("Application initialized successfully")

	// Initialize handlers
	logger.Info("Initializing handlers...")
	handler := handlers.NewHandlers(ctx, myApp)
	logger.Info("Handlers initialized successfully")

	// Initialize router
	logger.Info("Initializing router...")
	router.InitRouter(ctx, myApp, handler)
	logger.Info("Router initialized successfully")

	// Initialize job runner
	logger.Info("Initializing job runner...")
	err = jobs.InitJob(myApp)
	if err != nil {
		logger.Fatalf("Failed to initialize the job runner: %v", err)
	}
	logger.Info("Job runner initialized successfully")

	// Handle system signals
	logger.Info("Setting up signal handling...")
	app.HandleSignals(ctx, logger, cancel)
	logger.Info("Signal handling setup complete")

	// Start the application
	logger.Info("Starting the application...")
	myApp.Start(ctx)
	logger.Info("Application started successfully")
}
