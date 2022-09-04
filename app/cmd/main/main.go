package main

import (
	log "github.com/sirupsen/logrus"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	_ "skeleton-golange-application/app/docs"

	//https://github.com/swaggo/gin-swagger

	_ "github.com/joho/godotenv/autoload"
	"skeleton-golange-application/app/internal/app"
)

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("Starting the service...")
	log.Print("config initializing")
	cfg := config.GetConfig()
	log.Print("logger initializing")
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)

	a, err := app.NewApp(cfg, &logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("Running Application")
	a.Run()
}
