package gin

import (
	"context"
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
}

func NewAppUseGin(cfg *config.Config, logger *logging.Logger) (*WebApp, error) {
	logger.Info("router initializing")

	// Gin instance
	gin.SetMode(cfg.AppConfig.GinMode)
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
	m := monitoring.NewMetrics(reg)

	healthMetrics := monitoring.NewHealthMetrics()
	go monitoring.PingStorage(context.Background(), storage.Operations, healthMetrics)

	return &WebApp{
		cfg:           cfg,
		logger:        logger,
		router:        router,
		storage:       storage,
		healthMetrics: healthMetrics,
		metrics:       m,
	}, nil
}

func (a *WebApp) Run() {
	a.startHTTP()
}

func (a *WebApp) startHTTP() {
	a.logger.Info("start HTTP")
	// Routes
	a.logger.Info("heartbeat metric initializing")
	a.router.GET("/health", func(c *gin.Context) {
		monitoring.HealthGET(c, a.healthMetrics) // Pass the healthMetrics to HealthGET.
	})
	// Group: v1
	v1 := a.router.Group("/v1")

	v1.POST("/users/register", a.Register)
	v1.POST("/users/login", a.Login)
	v1.GET("/user", a.User)
	v1.POST("/users/delete", a.DeleteUser)
	v1.POST("users/logout", a.Logout)
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

	a.router.GET("/ping", Ping)

	a.logger.Info("application completely initialized and started")
	a.logger.Info("The service is ready to listen and serve.")
	// Start server
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	if err := a.router.Run(connectionString); err != nil {
		a.logger.Fatal(err)
	}
}
