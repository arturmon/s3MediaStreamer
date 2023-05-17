package main

import (
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
	_ "github.com/swaggo/gin-swagger"
	_ "skeleton-golange-application/app/docs"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/web/gin"
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
	log.Print("logger initializing")
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)
	logger.Info("Starting the service...")
	logger.Info("config initializing")
	appInstanceUseGin, err := gin.NewAppUseGin(cfg, &logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("Running Application")
	appInstanceUseGin.Run() // The app will run

}
