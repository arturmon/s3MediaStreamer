package app

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

type App struct {
	cfg        *config.Config
	logger     *logging.Logger
	router     *gin.Engine
	httpServer *http.Server
	storage    *model.DBConfig
}

func NewApp(config *config.Config, logger *logging.Logger) (App, error) {
	logger.Println("router initializing")

	// Gin instance
	router := gin.New()
	logger.Println("setup CORS")
	router.Use(CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	logger.Println("prometheus initializing")
	p := ginPrometheus.NewPrometheus("gin")
	p.Use(router)

	storage, err := model.NewDBConfig(config)
	if err != nil {
		return App{}, err
	}
	ctx := context.Background()
	go monitoring.PingStorage(ctx, storage.Operations)

	return App{
		cfg:     config,
		logger:  logger,
		router:  router,
		storage: storage,
	}, nil
}

func (a *App) Run() {
	a.startHTTP()
}

func (a *App) startHTTP() {

	a.logger.Info("start HTTP")
	// Routes
	a.logger.Println("heartbeat metric initializing")
	a.router.GET("/health", monitoring.HealthGET)
	// Group: v1
	v1 := a.router.Group("/v1")
	{
		v1.GET("/ping", Ping)
		v1.POST("/register", a.Register)
		v1.POST("/login", a.Login)
		v1.GET("/user", a.User)
		v1.POST("/deleteUser", a.DeleteUser)
		v1.POST("/logout", a.Logout)
		v1.GET("/albums", a.GetAllAlbums)
		v1.GET("/albums/:code", a.GetAlbumByID)
		v1.POST("/album", a.PostAlbums)
		v1.DELETE("/albums/deleteAll", a.GetDeleteAll)
		v1.DELETE("/albums/delete/:code", a.GetDeleteByID)
		a.logger.Println("swagger docs initializing")
		v1.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	a.logger.Println("application completely initialized and started")
	a.logger.Info("The service is ready to listen and serve.")
	// Start server
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	if err := a.router.Run(connectionString); err != nil {
		a.logger.Fatal(err)
	}
}
