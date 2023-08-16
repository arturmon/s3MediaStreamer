package config

import (
	"skeleton-golange-application/app/pkg/logging"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

// getConfigManager returns a singleton instance of the configuration manager.
func getConfigManager() *configManager {
	return &configManager{}
}

// ConfigManager is responsible for managing the application's configuration.
type configManager struct {
	instance *Config
	once     sync.Once
}

var (
	cfgManager = getConfigManager()
)

// Album represents data about a record album.
type Album struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" pg:"default:now()" swaggerignore:"true"`
	Title       string    `json:"title" bson:"title" example:"Title name"`
	Artist      string    `json:"artist" bson:"artist" example:"Artist name"`
	Price       float64   `json:"price" bson:"price" example:"111.111"`
	Code        string    `json:"code" bson:"code" example:"I001"`
	Description string    `json:"description" bson:"description" example:"A short description of the application"`
	Completed   bool      `json:"completed" bson:"completed" example:"false"`
}

// User represents user account information.
// @Description User account information
// @Description with: user _id, name, email, password
type User struct {
	Id       uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	Name     string    `json:"name" bson:"name" example:"Artur"`
	Email    string    `json:"email" bson:"email" example:"aaaa@aaaa.com"`
	Password []byte    `json:"-" bson:"password"  example:"1111"`
}

// Config represents the application's configuration.
type Config struct {
	// AppHealth stores the status of the application's health.
	AppHealth bool
	Listen    struct {
		BindIP string `env:"BIND_IP" env-default:"0.0.0.0"`
		Port   string `env:"PORT" env-default:"10000"`
	}
	AppConfig struct {
		LogLevel string `env:"LOG_LEVEL" env-default:"info"`
		// debug, release
		GinMode string `env:"GIN_MODE" env-default:"release"`
	}
	// STORAGE_TYPE: mongodb, postgresql
	// 5432 postgresql, 27017 mongodb
	Storage struct {
		Type     string `env:"STORAGE_TYPE" env-default:"postgresql"`
		Username string `env:"STORAGE_USERNAME" env-default:"root"`
		Password string `env:"STORAGE_PASSWORD" env-default:"1qazxsw2"`
		Host     string `env:"STORAGE_HOST" env-default:"localhost"`
		Port     string `env:"STORAGE_PORT" env-default:"5432"`
		Database string `env:"STORAGE_DATABASE" env-default:"db_issue_album"`
		// Mongo use
		Collections      string `env:"STORAGE_COLLECTIONS" env-default:"col_issues"`
		CollectionsUsers string `env:"STORAGE_COLLECTIONS_USERS" env-default:"col_users"`
	}
	// MessageQueue
	MessageQueue struct {
		Enable        bool   `env:"MQ_ENABLE" env-default:"false"`
		SubRoutingKey string `env:"MQ_ROUTING_KEY" env-default:"sub-routing-key"`
		SubQueueName  string `env:"MQ_QUEUE_NAME" env-default:"sub_queue"`
		PubExchange   string `env:"MQ_EXCHANGE" env-default:"pub-exchange"`
		PubRoutingKey string `env:"MQ_ROUTING_KEY" env-default:"pub-routing-key"`
		PubQueueName  string `env:"MQ_QUEUE_NAME" env-default:"pub_queue"`
		User          string `env:"MQ_USER" env-default:"user"`
		Pass          string `env:"MQ_PASS" env-default:"password"`
		Broker        string `env:"MQ_BROKER" env-default:"localhost"`
		BrokerPort    int    `env:"MQ_BROKER_PORT" env-default:"5672"`
	}
}

// GetConfig returns the singleton instance of the configuration.
func GetConfig() *Config {
	cfgManager := getConfigManager()

	cfgManager.once.Do(func() {
		log.Info("gather config")

		cfgManager.instance = &Config{}

		if err := cleanenv.ReadEnv(cfgManager.instance); err != nil {
			helpText := "The Art of Development - Monolith Notes System"
			help, _ := cleanenv.GetDescription(cfgManager.instance, &helpText)
			log.Debug(help)
			log.Fatal(err)
		}
	})

	return cfgManager.instance
}

// PrintAllDefaultEnvs prints the help text containing all the default environment variables.
func PrintAllDefaultEnvs(logger *logging.Logger) {
	cfg := &Config{}
	helpText := "The Art of Development - Monolith Notes System"
	help, _ := cleanenv.GetDescription(cfg, &helpText)
	// Print the help text containing all the default environment variables
	logger.Debug(help)
}

// GetAppHealth returns the value of AppHealth.
func GetAppHealth() bool {
	return cfgManager.instance.AppHealth
}

// SetAppHealth sets the value of AppHealth.
func SetAppHealth(health bool) {
	cfgManager.instance.AppHealth = health
}
