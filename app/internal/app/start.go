package app

import (
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/health"
	"time"

	"context"
)

func (a *App) Start(ctx context.Context) {
	go a.Service.ConsulElection.Init()
	a.handleHealthCheckResults(ctx, a.Service.Health)

	a.Logger.Info("🚀 Running Application...")
	a.Run(ctx)

	a.Logger.Info("Application stopped")
}

// StartPprofServer запускает сервер pprof для профилирования.
func StartPprofServer(logger *logs.Logger) {
	server := &http.Server{
		Addr:              "localhost:6060",
		Handler:           nil,
		ReadHeaderTimeout: PPOFReadHeaderTimeout * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err.Error())
	}

	logger.Info("Use endpoint ppof http://localhost:6060/debug/pprof/")
}

func (a *App) Run(ctx context.Context) {
	a.startHTTP(ctx)
}

func (a *App) handleHealthCheckResults(ctx context.Context, healthCheckWrapper *health.Service) {
	resultChan := make(chan bool)
	healthCheckWrapper.CheckMonitoring(ctx, resultChan)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // exit goroutine when context is canceled
			case isHealthy := <-resultChan:
				if !isHealthy && a.Service.ConsulElection.IsLeader() {
					// Trigger ReElection if components are not healthy
					err := a.Service.ConsulElection.ReElection(a.Service.ConsulElection.GetElectionClient())
					if err != nil {
						return
					}
				}
			}
		}
	}()
}
