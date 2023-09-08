package gin

import (
	"context"
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/casbin/casbin/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

const shutdownTimeout = 5 * time.Second
const ReadHeaderTimeout = 5 * time.Second

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

	enforcer, err := GetEnforcer(cfg, storage)
	if err != nil {
		return nil, err
	}

	initErr := initRoles(enforcer)
	if initErr != nil {
		return nil, initErr
	}
	if saveErr := enforcer.SavePolicy(); saveErr != nil {
		return nil, saveErr
	}

	// Initialize session
	logger.Info("session initializing")
	initSession(ctx, router, cfg, logger)

	healthMetrics := monitoring.NewHealthMetrics()
	go monitoring.PingStorage(context.Background(), storage.Operations, healthMetrics)

	return &WebApp{
		cfg:           cfg,
		logger:        logger,
		router:        router,
		storage:       storage,
		healthMetrics: healthMetrics,
		metrics:       metrics,
		enforcer:      enforcer,
	}, nil
}

func (a *WebApp) Run(ctx context.Context) {
	a.startHTTP(ctx)
}

func (a *WebApp) startHTTP(ctx context.Context) {
	a.logger.Info("start HTTP")
	// Routes
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

	// Group: v1
	v1 := a.router.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", a.Register)
			users.OPTIONS("/login", handleOptions)
			users.POST("/login", a.Login)
			users.GET("/me", a.User)
			users.POST("/delete", a.DeleteUser)
			users.POST("/logout", a.Logout)
			users.POST("/refresh", a.refreshTokenHandler)
			otp := users.Group("/otp")
			{
				otp.POST("/generate", a.GenerateOTP)
				otp.POST("/verify", a.VerifyOTP)
				otp.POST("/validate", a.ValidateOTP)
				otp.POST("/disable", a.DisableOTP)
			}
		}
		albums := v1.Group("/albums")
		{
			albums.GET("", a.GetAllAlbums)
			albums.GET("/:code", a.GetAlbumByID)
			albums.DELETE("/deleteAll", a.GetDeleteAll)
			albums.DELETE("/delete/:code", a.GetDeleteByID)
			albums.POST("/add", a.PostAlbums)
			albums.POST("/update", a.UpdateAlbum)
		}
		a.logger.Info("swagger docs initializing")
		swagger := v1.Group("/swagger")
		{
			swagger.GET("", func(c *gin.Context) {
				c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			})
			swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	}
	a.logger.Info("view Casbin Policies:")
	policies := a.enforcer.GetPolicy()
	for _, p := range policies {
		a.logger.Infof("Policy: %v", p)
	}
	a.logger.Info("application completely initialized, ...started")
	a.logger.Infof("The service is ready to listen and serve on %s:%s.", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	// Start server
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	server := &http.Server{
		Addr:              connectionString,
		Handler:           a.router,
		ReadHeaderTimeout: ReadHeaderTimeout, // Set the ReadHeaderTimeout here
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error:", err)
		}
	}()

	// Wait for context cancellation to stop the server gracefully
	<-ctx.Done()

	// Shutdown the server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		a.logger.Fatal("HTTP server shutdown error:", err)
	}

	a.logger.Info("Application stopped")
}

func LogWithLogrusf(logger *logging.Logger, format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
