package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/internal/jobs"
	"skeleton-golange-application/app/pkg/amqp"
	"skeleton-golange-application/app/pkg/consul"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	_ "skeleton-golange-application/app/pkg/web/gin"
	"time"

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
	// debug.SetMemoryLimit(2048)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.GetConfig()
	logger := logging.GetLogger(cfg.AppConfig.LogLevel, cfg.AppConfig.LogType)
	logger.Printf("App Version: %s Build Time: %s\n", Version, BuildTime)
	logger.Info("config initialize")
	logger.Info("logger initialize")
	if cfg.AppConfig.LogLevel == "debug" {
		config.PrintAllDefaultEnvs(&logger)
		go func() {
			server := &http.Server{
				Addr:              "localhost:6060",
				Handler:           nil,
				ReadHeaderTimeout: PPOFReadHeaderTimeout * time.Second,
			}
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal(err)
			}
			logger.Println("Use endpoint ppof http://localhost:6060/debug/pprof/")
		}()
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

	// Specify the number of workers in the pool
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err = amqp.ConsumeMessagesWithPool(ctx, logger, myApp.GetMessageClient(), numWorkers, workerDone); err != nil {
			// Handle error
			logger.Fatal(err)
		}
	}()

	go myApp.LeaderElection.Init()

	// Start monitoring the database storage with a ticker
	healthMetrics := monitoring.NewHealthMetrics()
	// Call PingStorage in a goroutine

	healthCheckWrapper := monitoring.NewHealthCheckWrapper(healthMetrics, myApp.Storage.Operations, myApp.AMQPClient, myApp.S3, &logger)
	healthCheckWrapper.StartHealthChecks()

	resultChan := make(chan bool)
	healthCheckWrapper.CheckMonitoring(ctx, resultChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return // exit goroutine when context is canceled
			case isHealthy := <-resultChan:
				if !isHealthy && myApp.LeaderElection.IsLeader() {
					// Trigger ReElection if components are not healthy
					consul.ReElection(myApp.LeaderElection)
				}
			}
		}
	}()

	logger.Info("🚀 Running Application...")
	myApp.Gin.Run(ctx, healthCheckWrapper) // The app will run
	if err != nil {
		logger.Fatal(err)
	}
	// Wait for the context to be cancelled before exiting
	<-ctx.Done()
	myApp.LeaderElection.Stop()
	err = consul.DeregisterService(myApp.Consul, myApp.AppName+"-"+consul.GetHostname())
	if err != nil {
		logger.Errorf("Error deregistering service: %v", err)
	}
	logger.Info("Application stopped")
}

var (
	Version   = "0.0.1"
	BuildTime = "0000-00-00 UTC"
)

const PPOFReadHeaderTimeout = 10
