package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/logging"
)

// @title			Sceleton Golang Application API
// @version		0.0.1
// @description	This is a sample server Petstore server.
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http
// @host      localhost:10000
// @BasePath	/v1
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @securityDefinitions.basic	BasicAuth

func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)
	myApp, err := app.NewAppInit(cfg, &logger)
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(&logger)
	}
	// Start consuming messages
	messages, err := myApp.GetMessageClient().Consume(context.Background())
	if err != nil {
		logger.Fatal("Failed to start consuming messages:", err)
	}
	// Call PreInit with the AMQPClient instance and the config
	err = amqp.PostInit(myApp.GetMessageClient(), cfg)
	if err != nil {
		logger.Fatal("Failed to pre-initialize AMQP:", err)
	}

	logger.Info("Running Application")
	myApp.Gin.Run() // The app will run
	// Wait for the background goroutine to finish
	<-messages
	logger.Info("Background goroutine finished.")
}
