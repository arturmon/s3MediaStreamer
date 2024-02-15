package config

import (
	"fmt"
	"skeleton-golange-application/app/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

// getConfigManager returns a singleton instance of the configuration manager.
func getConfigManager() *Manager {
	return &Manager{}
}

// Manager is responsible for managing the application's configuration.
type Manager struct {
	instance *Config
	once     sync.Once
}

// GetConfig returns the singleton instance of the configuration.
func GetConfig() *Config {
	cfgManager := getConfigManager()

	cfgManager.once.Do(func() {
		log.Info("gathering config")

		cfgManager.instance = &Config{}
	})

	if err := cleanenv.ReadEnv(cfgManager.instance); err != nil {
		helpText := fmt.Sprintf("Stream Player S3")
		help, _ := cleanenv.GetDescription(cfgManager.instance, &helpText)
		log.Debug(help)
		log.Errorf("Error reading environment variables: %v", err)
	}

	return cfgManager.instance
}

// PrintAllDefaultEnvs prints the help text containing all the default environment variables.
func PrintAllDefaultEnvs(logger *logging.Logger) {
	cfg := &Config{}
	helpText := fmt.Sprintf("Stream Player S3")
	help, _ := cleanenv.GetDescription(cfg, &helpText)
	// Print the help text containing all the default environment variables
	logger.Debug(help)
}
