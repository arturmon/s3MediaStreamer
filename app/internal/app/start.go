package app

import (
	"net/http"
	"skeleton-golange-application/app/pkg/amqp"
	consul_election "skeleton-golange-application/app/pkg/consulelection"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"time"

	"context"
)

func (a *App) Start(ctx context.Context) {
	a.startAMQPConsumers(ctx)
	go a.LeaderElection.Election.Init()
	healthCheckWrapper := a.startHealthChecks()
	a.handleHealthCheckResults(ctx, healthCheckWrapper)

	a.Logger.Info("🚀 Running Application...")
	a.Gin.Run(ctx, healthCheckWrapper)
	a.Logger.Info("Application stopped")
}

// StartPprofServer запускает сервер pprof для профилирования.
func StartPprofServer(logger *logging.Logger) {
	server := &http.Server{
		Addr:              "localhost:6060",
		Handler:           nil,
		ReadHeaderTimeout: PPOFReadHeaderTimeout * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	logger.Println("Use endpoint ppof http://localhost:6060/debug/pprof/")
}

func (a *App) startAMQPConsumers(ctx context.Context) {
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err := amqp.ConsumeMessagesWithPool(ctx, *a.Logger, a.GetMessageClient(), numWorkers, workerDone); err != nil {
			// Handle error
			a.Logger.Fatal(err)
		}
	}()
}

func (a *App) startHealthChecks() *monitoring.HealthCheckWrapper {
	healthMetrics := monitoring.NewHealthMetrics()
	healthCheckWrapper := monitoring.NewHealthCheckWrapper(healthMetrics, a.Storage.Operations, a.AMQPClient, a.S3, a.Logger)
	healthCheckWrapper.StartHealthChecks()
	return healthCheckWrapper
}

func (a *App) handleHealthCheckResults(ctx context.Context, healthCheckWrapper *monitoring.HealthCheckWrapper) {
	resultChan := make(chan bool)
	healthCheckWrapper.CheckMonitoring(ctx, resultChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return // exit goroutine when context is canceled
			case isHealthy := <-resultChan:
				if !isHealthy && a.LeaderElection.Election.IsLeader() {
					// Trigger ReElection if components are not healthy
					consul_election.ReElection(a.LeaderElection.Election)
				}
			}
		}
	}()
}
