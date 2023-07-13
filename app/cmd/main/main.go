package main

import (
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/swaggo/gin-swagger"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
)

// @title			Sceleton Golang Application API
// @version		1.0
// @description	This is a sample server Petstore server.
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath	/v1
func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)
	app, err := app.NewAppInit(cfg, &logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("Running Application")
	app.Gin.Run() // The app will run
}
