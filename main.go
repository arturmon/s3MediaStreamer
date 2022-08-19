package main

import (
	log "github.com/sirupsen/logrus"

	"skeleton-golange-application/docs"
	_ "skeleton-golange-application/docs"

	//https://github.com/swaggo/gin-swagger

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("Starting the service...")

	GetMongoClient()

	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Swagger Albums API"
	docs.SwaggerInfo.Description = "This is a sample server albums server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Gin instance
	router := gin.New()
	router.Use(gin.Logger())
	/*
		router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

			// your custom format
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		}))
	*/
	router.Use(gin.Recovery())

	// Routes
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/ping", Ping)
	router.GET("/albums", getAllAlbums)
	router.GET("/albums/:code", getAlbumByID)
	router.POST("/albums", postAlbums)
	router.GET("/albums/deleteAll", getDeleteAll)
	router.GET("/albums/delete/:code", getDeleteByID)

	router.GET("/health", HealthGET)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Group: v1
	v1 := router.Group("/v1")
	{
		v1.GET("/ping", Ping)
		v1.GET("/health", HealthGET)
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	log.Info("The service is ready to listen and serve.")
	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
