package config

import (
	"log"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// getConfigManager returns a singleton instance of the configuration manager.
func getConfigManager() *Manager {
	return &Manager{}
}

// Manager is responsible for managing the application's configuration.
type Manager struct {
	instance *model.Config
	once     sync.Once
}

// GetConfig returns the singleton instance of the configuration.
func GetConfig() *model.Config {
	cfgManager := getConfigManager()

	cfgManager.once.Do(func() {
		log.Print("gathering config")

		cfgManager.instance = &model.Config{}
	})

	if err := cleanenv.ReadConfig("conf/application.yml", cfgManager.instance); err != nil {
		helpText := "Stream Player S3"
		help, _ := cleanenv.GetDescription(cfgManager.instance, &helpText)
		log.Printf(help)
		log.Printf("Error reading environment variables: %v", err)
	}
	go func() {
		time.Sleep(sleepDurationSeconds * time.Second) // sleep for 5 seconds
		err := cleanenv.UpdateEnv(cfgManager.instance)
		if err != nil {
			log.Printf("Error update environment variables: %v", err)
		}
	}()
	return cfgManager.instance
}

// PrintAllDefaultEnvs prints the help text containing all the default environment variables.
func PrintAllDefaultEnvs(logger *logs.Logger) {
	cfg := &model.Config{}
	helpText := "Stream Player S3"
	help, _ := cleanenv.GetDescription(cfg, &helpText)
	// Print the help text containing all the default environment variables
	logger.Debug(help)
}
