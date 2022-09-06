package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/mongodb"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"
)

type App struct {
	cfg         *config.Config
	logger      *logging.Logger
	router      *gin.Engine
	httpServer  *http.Server
	mongoClient *mongo.Client
}

func NewApp(config *config.Config, logger *logging.Logger) (App, error) {
	logger.Println("router initializing")

	// Gin instance
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	logger.Println("prometheus initializing")
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	mongoConfig := mongodb.NewMongoConfig(
		config.Storage.MongoDB.Username, config.Storage.MongoDB.Password,
		config.Storage.MongoDB.Host, config.Storage.MongoDB.Port, config.Storage.MongoDB.Database, config.Storage.MongoDB.Collections,
	)
	// mongoClient setup
	mongoClient, err := mongodb.GetMongoClient(mongoConfig)
	if err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()
	go monitoring.PingStorage(ctx, mongoClient, config)

	return App{
		cfg:         config,
		logger:      logger,
		router:      router,
		mongoClient: mongoClient,
	}, nil
}

func (a *App) Run() {
	a.startHTTP()
}

func (a *App) startHTTP() {

	a.logger.Info("start HTTP")
	// Routes

	a.router.GET("/ping", Ping)
	a.router.GET("/albums", a.GetAllAlbums)
	a.router.GET("/albums/:code", a.GetAlbumByID)
	a.router.POST("/albums", a.PostAlbums)
	a.router.GET("/albums/deleteAll", a.GetDeleteAll)
	a.router.GET("/albums/delete/:code", a.GetDeleteByID)

	a.logger.Println("heartbeat metric initializing")
	a.router.GET("/health", monitoring.HealthGET)

	a.logger.Println("swagger docs initializing")
	a.router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Group: v1
	v1 := a.router.Group("/v1")
	{
		v1.GET("/ping", Ping)
		a.logger.Println("heartbeat metric initializing")
		v1.GET("/health", monitoring.HealthGET)
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	a.router.Use(CORSMiddleware())
	a.logger.Println("application completely initialized and started")
	a.logger.Info("The service is ready to listen and serve.")
	// Start server
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	if err := a.router.Run(connectionString); err != nil {
		a.logger.Fatal(err)
	}
}
