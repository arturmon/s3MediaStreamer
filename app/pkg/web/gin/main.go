package gin

import (
	"context"
	"fmt"
	"net/http"
	conf "s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/caching"
	"s3MediaStreamer/app/pkg/client/model"
	"s3MediaStreamer/app/pkg/logging"
	"s3MediaStreamer/app/pkg/monitoring"
	"s3MediaStreamer/app/pkg/s3"

	ginPrometheus "github.com/penglongli/gin-metrics/ginmetrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/gin-contrib/cors"

	"github.com/casbin/casbin/v2"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type AppInterface interface {
	Run()
}

type WebApp struct {
	cfg         *conf.Config
	logger      *logging.Logger
	router      *gin.Engine
	storage     *model.DBConfig
	metrics     *monitoring.Metrics
	enforcer    *casbin.Enforcer
	S3          s3.HandlerS3
	redisClient *caching.CachingStruct
}

func NewAppUseGin(ctx context.Context, cfg *conf.Config, logger *logging.Logger) (*WebApp, error) {
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

	s3Handler, s3err := s3.NewClientS3(ctx, cfg, logger)
	if s3err != nil {
		logger.Error("Failed to initialize S3:", s3err)
		logger.Fatal(s3err)
		return nil, s3err
	}

	// Init caching
	logger.Info("redis initializing")
	redisClient := caching.InitRedis(
		ctx,
		cfg.Storage.Caching.Address,
		cfg.Storage.Caching.Password)
	if redisClient == nil && !cfg.Storage.Caching.Enabled {
		logger.Info("redis is NOT initializing or disabled !!!")
	}

	return &WebApp{
		cfg:         cfg,
		logger:      logger,
		router:      router,
		storage:     storage,
		metrics:     metrics,
		enforcer:    enforcer,
		S3:          s3Handler,
		redisClient: redisClient,
	}, nil
}

func (a *WebApp) Run(ctx context.Context, hcw *monitoring.HealthCheckWrapper) {
	a.startHTTP(ctx, hcw)
}

func (a *WebApp) startHTTP(ctx context.Context, hcw *monitoring.HealthCheckWrapper) {
	a.logger.Info("start HTTP")
	a.setupStaticFiles()

	// Routes
	a.setupSystemRoutes(ctx, hcw)
	// Group: v1
	a.setupAppRoutesV1()

	a.logger.Debug("view Casbin Policies:")
	policies := a.enforcer.GetPolicy()
	var logMessage string
	for _, p := range policies {
		logMessage += fmt.Sprintf("Policy: %v\n", p)
	}
	a.logger.Debugf(logMessage)
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

func (a *WebApp) setupSystemRoutes(ctx context.Context, hcw *monitoring.HealthCheckWrapper) {
	_, span := otel.Tracer("").Start(ctx, "setupSystemRoutes")
	defer span.End()
	a.logger.Info("heartbeat metric initializing")
	a.router.GET("/health/liveness", func(c *gin.Context) {
		monitoring.LivenessGET(c, hcw) // Pass the healthMetrics to HealthGET.
	})
	a.router.GET("/health/readiness", func(c *gin.Context) {
		monitoring.ReadinessGET(c, hcw) // Pass the healthMetrics to HealthGET.
	})

	a.router.GET("/job/status", JobStatus)

	a.router.Use(ExtractUserRole(ctx, a.logger))
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
