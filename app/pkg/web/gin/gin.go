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

func NewAppUseGin(cfg *config.Config, logger *logging.Logger) (*WebApp, error) {
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
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	logger.Info("prometheus initializing")
	p := ginPrometheus.NewPrometheus("gin")
	p.Use(router)

	logger.Info("storage initializing")
	storage, err := model.NewDBConfig(cfg)
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
	initSession(router, cfg)

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

	// Group: v1
	v1 := a.router.Group("/v1")
	v1.Use(ExtractUserRole(a.logger))
	v1.Use(NewAuthorizerWithRoleExtractor(a.enforcer, a.logger, func(c *gin.Context) string {
		if role, ok := c.Get("userRole"); ok {
			return role.(string)
		}
		return "anonymous" // Default role
	}))

	v1.POST("/users/register", a.Register)
	v1.OPTIONS("/users/login", handleOptions)
	v1.POST("/users/login", a.Login)
	v1.GET("/users/me", a.User)
	v1.POST("/users/delete", a.DeleteUser)
	v1.POST("/users/logout", a.Logout)
	v1.GET("/albums", a.GetAllAlbums)
	v1.GET("/albums/:code", a.GetAlbumByID)
	v1.POST("/album", a.PostAlbums)
	v1.POST("/album/update", a.UpdateAlbum)
	v1.DELETE("/albums/deleteAll", a.GetDeleteAll)
	v1.DELETE("/albums/delete/:code", a.GetDeleteByID)

	a.logger.Info("swagger docs initializing")
	v1.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	a.logger.Info("view Casbin Policies:")
	policies := a.enforcer.GetPolicy()
	for _, p := range policies {
		a.logger.Infof("Policy: %v", p)
	}
	a.logger.Info("application completely initialized and started")
	a.logger.Info("The service is ready to listen and serve.")
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
