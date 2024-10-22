package app

import (
	"context"
	"fmt"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginPrometheus "github.com/penglongli/gin-metrics/ginmetrics"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const shutdownTimeout = 5 * time.Second
const ReadHeaderTimeout = 5 * time.Second

func initializeGin(_ context.Context, cfg *model.Config, logger *logs.Logger) *gin.Engine {
	logger.Info("router initializing")
	// Gin instance
	switch cfg.AppConfig.Web.Mode {
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

	router.Use(otelgin.Middleware("s3MediaStreamer", otelgin.WithGinFilter(excludePathsFromTracing(excludedPaths))))

	config := sloggin.Config{
		WithSpanID:         cfg.AppConfig.Web.Debug.WithSpanID,
		WithTraceID:        cfg.AppConfig.Web.Debug.WithTraceID,
		WithRequestBody:    cfg.AppConfig.Web.Debug.WithRequestBody,
		WithResponseBody:   cfg.AppConfig.Web.Debug.WithResponseBody,
		WithRequestHeader:  cfg.AppConfig.Web.Debug.WithRequestHeader,
		WithResponseHeader: cfg.AppConfig.Web.Debug.WithResponseHeader,
	}

	if config.WithSpanID || config.WithTraceID || config.WithRequestBody || config.WithResponseBody || config.WithRequestHeader || config.WithResponseHeader {
		router.Use(sloggin.New(logger.WithGroup("http")))
		router.Use(sloggin.NewWithConfig(logger.Slog(), config))
	}

	logger.Info("prometheus initializing")

	p := ginPrometheus.GetMonitor()
	p.SetMetricPath("/metrics")
	p.Use(router)

	return router
}

func (a *App) startHTTP(ctx context.Context) {
	a.Logger.Info("start HTTP")

	a.Logger.Debug("view Casbin Policies:")
	a.Service.ACL.GetAllPolicies(a.Logger)

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
			a.Logger.Fatalf("HTTP server error: %s", err)
		}
	}()

	return server
}

func (a *App) shutdownServer(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.Logger.Fatalf("HTTP server shutdown error: %s", err)
	}
}
