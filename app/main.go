package main

import (
	"context"
	"os"
	"os/signal"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/logging"
	"syscall"

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

// @schemes http
// @host      localhost:10000
// @BasePath	/v1
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @securityDefinitions.basic	BasicAuth

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

	// Setup signal handling to gracefully stop the application
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			logger.Info("Received termination signal. Stopping the application...")
			cancel()
		case <-ctx.Done():
			// Context cancelled, exiting goroutine
		}
	}()

	// Start consuming messages
	if myApp.GetMessageClient() != nil {
		go func() {
			messages, err := myApp.GetMessageClient().Consume(ctx)
			if err != nil {
				logger.Fatal("Failed to start consuming messages:", err)
			}
			// Wait for the background goroutine to finish
			logger.Info("Waiting for the background goroutine to finish...")
			<-messages
			logger.Info("Finished consuming messages")
			cancel() // Cancel the context to stop other goroutines
		}()
	}

	// Call PreInit with the AMQPClient instance and the config
	if myApp.GetMessageClient() != nil {
		err = amqp.PostInit(myApp.GetMessageClient(), cfg)
		if err != nil {
			logger.Fatal("Failed to pre-initialize AMQP:", err)
		}
	}

	logger.Info("Running Application")
	myApp.Gin.Run() // The app will run

	// Wait for the context to be cancelled before exiting
	<-ctx.Done()
	logger.Info("Application stopped")
}
