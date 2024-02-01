package main

import (
	"context"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/internal/jobs"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/logging"
	_ "skeleton-golange-application/app/pkg/web/gin"

	_ "github.com/joho/godotenv/autoload"
)

// @title			Sceleton Golang Application API
// @version		0.0.1
// @description	This is a sample server Petstore server.
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http https
// @host      localhost:10000
// @BasePath	/v1
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @securityDefinitions.basic	BasicAuth
// @authorizationurl http://localhost:10000/v1/users/login
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.GetConfig()
	logger := logging.GetLogger(cfg.AppConfig.LogLevel, cfg.AppConfig.LogType)
	logger.Info("config initialize")
	logger.Info("logger initialize")
	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(&logger)
	}
	myApp, err := app.NewAppInit(cfg, &logger)
	if err != nil {
		logger.Error("Failed to initialize the new my app:", err)
	}

	logger.Info("Starting initialize the job runner...")
	err = jobs.InitJob(myApp)
	if err != nil {
		logger.Error("Failed to initialize the job runner:", err)
	}

	app.HandleSignals(ctx, logger, cancel)
	/*
		go func() {
			if err := amqp.ConsumeMessages(ctx, logger, myApp.GetMessageClient()); err != nil {
				logger.Fatal(err)
			}
		}()

	*/

	// Specify the number of workers in the pool
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err = amqp.ConsumeMessagesWithPool(ctx, logger, myApp.GetMessageClient(), numWorkers, workerDone); err != nil {
			// Handle error
			logger.Fatal(err)
		}
	}()

	logger.Info("🚀 Running Application...")
	myApp.Gin.Run(ctx) // The app will run
	if err != nil {
		logger.Fatal(err)
	}
	// Wait for the context to be cancelled before exiting
	<-ctx.Done()
	logger.Info("Application stopped")
}
