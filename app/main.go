package main

import (
	"context"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/logging"

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
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)
	logger.Info("config initialize")
	logger.Info("logger initialize")
	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(&logger)
	}
	myApp, err := app.NewAppInit(cfg, &logger)
	if err != nil {
		logger.Fatal(err)
	}

	app.HandleSignals(ctx, logger, cancel)
	amqp.ConsumeMessages(ctx, logger, myApp.GetMessageClient())

	// Call PostInit with the AMQPClient instance and the config
	if myApp.GetMessageClient() != nil {
		err = amqp.PostInit(myApp.GetMessageClient(), cfg) // Use the existing variable name "err"
		if err != nil {
			logger.Fatal("Failed to pre-initialize AMQP:", err)
		}
	}

	logger.Info("Running Application")
	myApp.Gin.Run(ctx) // The app will run
	if err != nil {
		logger.Fatal(err)
	}
	// Wait for the context to be cancelled before exiting
	<-ctx.Done()
	logger.Info("Application stopped")
}
