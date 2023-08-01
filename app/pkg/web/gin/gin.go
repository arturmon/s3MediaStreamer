package gin

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
)

type AppInterface interface {
	Run()
}

type WebApp struct {
	cfg        *config.Config
	logger     *logging.Logger
	router     *gin.Engine
	httpServer *http.Server
	storage    *model.DBConfig
}

func NewAppUseGin(config *config.Config, logger *logging.Logger) (*WebApp, error) {
	logger.Info("router initializing")

	// Gin instance
	gin.SetMode(config.AppConfig.GinMode)
	router := gin.New()
	logger.Info("setup CORS")
	router.Use(CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	logger.Info("prometheus initializing")
	p := ginPrometheus.NewPrometheus("gin")
	p.Use(router)

	storage, err := model.NewDBConfig(config)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	go monitoring.PingStorage(ctx, storage.Operations)

	return &WebApp{
		cfg:     config,
		logger:  logger,
		router:  router,
		storage: storage,
	}, nil
}

func (a *WebApp) Run() {
	a.startHTTP()
}

func (a *WebApp) startHTTP() {

	a.logger.Info("start HTTP")
	// Routes
	a.logger.Info("heartbeat metric initializing")
	a.router.GET("/health", monitoring.HealthGET)
	// Group: v1
	v1 := a.router.Group("/v1")
	{
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
	}
	a.router.GET("/ping", Ping)

	a.logger.Info("application completely initialized and started")
	a.logger.Info("The service is ready to listen and serve.")
	// Start server
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	if err := a.router.Run(connectionString); err != nil {
		a.logger.Fatal(err)
	}
}
