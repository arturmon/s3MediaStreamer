package gin

import (
	"context"
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"skeleton-golange-application/app/pkg/s3"

	"github.com/gin-contrib/cors"

	"github.com/casbin/casbin/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gin-gonic/gin"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

type AppInterface interface {
	Run()
}

type WebApp struct {
	cfg           *config.Config
	logger        *logging.Logger
	router        *gin.Engine
	storage       *model.DBConfig
	healthMetrics *monitoring.HealthMetrics
	metrics       *monitoring.Metrics
	enforcer      *casbin.Enforcer
	S3            *s3.HandlerFromS3
}

func NewAppUseGin(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*WebApp, error) {
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

	router.Use(LoggingMiddleware(LoggingMiddlewareAdapter(logger)))

	logger.Info("prometheus initializing")
	p := ginPrometheus.NewPrometheus("gin")
	p.Use(router)

	logger.Info("storage initializing")
	storage, err := model.NewDBConfig(cfg, logger)
	if err != nil {
		return nil, err
	}

	reg := prometheus.NewRegistry()
	metrics := monitoring.NewMetrics(reg)

	enforcer, err := GetEnforcer(storage)
	if err != nil {
		return nil, err
	}

	if saveErr := enforcer.SavePolicy(); saveErr != nil {
		return nil, saveErr
	}

	// Initialize session
	logger.Info("session initializing")
	initSession(ctx, router, cfg, logger)

	healthMetrics := monitoring.NewHealthMetrics()
	go monitoring.PingStorage(context.Background(), storage.Operations, healthMetrics)

	s3client, s3err := (&s3.HandlerFromS3{}).NewClientS3(cfg, logger)
	if s3err != nil {
		logger.Error("Failed to initialize S3:", s3err)
		logger.Fatal(s3err)
		return nil, s3err
	}
	err = s3client.InitS3(ctx)
	if err != nil {
		return nil, err
	}

	return &WebApp{
		cfg:           cfg,
		logger:        logger,
		router:        router,
		storage:       storage,
		healthMetrics: healthMetrics,
		metrics:       metrics,
		enforcer:      enforcer,
		S3:            s3client,
	}, nil
}

func (a *WebApp) Run(ctx context.Context) {
	a.startHTTP(ctx)
}

func (a *WebApp) startHTTP(ctx context.Context) {
	a.logger.Info("start HTTP")
	a.setupStaticFiles()

	// Routes
	a.setupSystemRoutes()
	// Group: v1
	a.setupAppRoutesV1()

	a.logger.Info("view Casbin Policies:")
	policies := a.enforcer.GetPolicy()
	for _, p := range policies {
		a.logger.Infof("Policy: %v", p)
	}
	a.logger.Info("application completely initialized, ...started")
	a.logger.Infof("The service is ready to listen and serve on %s:%s.", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	// Start server

	server := a.startServer()
	// Wait for context cancellation to stop the server gracefully
	<-ctx.Done()

	// Shutdown the server gracefully
	a.shutdownServer(server)

	a.logger.Info("Application stopped")
}

func (a *WebApp) setupStaticFiles() {
	a.router.StaticFile("/favicon.ico", "./favicon.ico")
}

func (a *WebApp) setupSystemRoutes() {
	a.logger.Info("heartbeat metric initializing")
	a.router.GET("/health", func(c *gin.Context) {
		monitoring.HealthGET(c, a.healthMetrics) // Pass the healthMetrics to HealthGET.
	})
	a.router.GET("/ping", Ping)
	a.router.GET("/job/status", JobStatus)

	a.router.Use(ExtractUserRole(a.logger))
	a.router.Use(NewAuthorizerWithRoleExtractor(a.enforcer, a.logger, func(c *gin.Context) string {
		if role, ok := c.Get("userRole"); ok {
			return role.(string)
		}
		return "anonymous" // Default role
	}))
}

func (a *WebApp) startServer() *http.Server {
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	server := &http.Server{
		Addr:              connectionString,
		Handler:           a.router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error:", err)
		}
	}()

	return server
}

func (a *WebApp) shutdownServer(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.logger.Fatal("HTTP server shutdown error:", err)
	}
}

func LogWithLogrusf(logger *logging.Logger, format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
