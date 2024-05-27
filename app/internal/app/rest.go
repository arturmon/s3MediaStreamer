package app

import (
	"context"
	"fmt"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/health"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginPrometheus "github.com/penglongli/gin-metrics/ginmetrics"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const shutdownTimeout = 5 * time.Second
const ReadHeaderTimeout = 5 * time.Second

func initializeGin(_ context.Context, cfg *model.Config, logger *logs.Logger) (*gin.Engine, error) {
	logger.Info("router initializing")
	// Gin instance
	switch cfg.AppConfig.GinMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	logger.Info("setup CORS")
	router.Use(cors.New(ConfigCORS()))
	router.Use(gin.Recovery())

	router.Use(otelgin.Middleware("s3MediaStreamer"))
	router.Use(LoggingMiddleware(LoggingMiddlewareAdapter(logger)))

	logger.Info("prometheus initializing")

	p := ginPrometheus.GetMonitor()
	p.SetMetricPath("/metrics")
	p.Use(router)

	return router, nil
}

func (a *App) startHTTP(ctx context.Context, hcw *health.HealthCheckService) {
	a.Logger.Info("start HTTP")

	a.Logger.Debug("view Casbin Policies:")
	//policies := a.Service.AccessControl.GetPolicy()
	var logMessage string
	//for _, p := range policies {
	//	logMessage += fmt.Sprintf("Policy: %v\n", p)
	//}
	a.Logger.Debugf(logMessage)
	a.Logger.Info("application completely initialized, ...started")
	a.Logger.Infof("The services is ready to listen and serve on %s:%s.", a.Cfg.Listen.BindIP, a.Cfg.Listen.Port)
	// Start server

	server := a.startServer()
	// Wait for context cancellation to stop the server gracefully
	<-ctx.Done()

	// Shutdown the server gracefully
	a.shutdownServer(server)

	a.Logger.Info("Application stopped")
}

func (a *App) startServer() *http.Server {
	connectionString := fmt.Sprintf("%s:%s", a.Cfg.Listen.BindIP, a.Cfg.Listen.Port)
	server := &http.Server{
		Addr:              connectionString,
		Handler:           a.REST,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("HTTP server error:", err)
		}
	}()

	return server
}

func (a *App) shutdownServer(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.Logger.Fatal("HTTP server shutdown error:", err)
	}
}

func LogWithLogrusf(logger *logs.Logger, format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
