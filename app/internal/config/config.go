package config

import (
	"skeleton-golange-application/app/pkg/logging"
	"sync"
	"time"

	"github.com/bojanz/currency"

	"github.com/google/uuid"
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

// Album represents data about a record album.
type Album struct {
	ID          uuid.UUID       `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	UpdatedAt   time.Time       `json:"updated_at" bson:"updated_at" pg:"default:now()" swaggerignore:"true"`
	Title       string          `json:"title" bson:"title" example:"Title name"`
	Artist      string          `json:"artist" bson:"artist" example:"Artist name"`
	Price       currency.Amount `json:"price" bson:"price" example:"{Number: 1.10, Currency: EUR}" swaggertype:"string,string"`
	Code        string          `json:"code" bson:"code" example:"I001"`
	Description string          `json:"description" bson:"description" example:"A short description of the application"`
	Sender      string          `json:"sender" bson:"sender" example:"amqp or rest"`
	CreatorUser uuid.UUID       `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
}

// User represents user account information.
// @Description User account information
// @Description with: user _id, name, email, password
type User struct {
	ID       uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	Name     string    `json:"name" bson:"name" example:"Artur"`
	Email    string    `json:"email" bson:"email" example:"aaaa@aaaa.com"`
	Password []byte    `json:"-" bson:"password"  example:"1111"`
	Role     string    `json:"role" bson:"role"  example:"-"`
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
		LogLevel string `env:"LOG_LEVEL" env-default:"info"`   // debug, release
		GinMode  string `env:"GIN_MODE" env-default:"release"` // debug, test, release
	}
	Storage struct {
		Type     string `env:"STORAGE_TYPE" env-default:"postgresql"` // mongodb, postgresql
		Username string `env:"STORAGE_USERNAME" env-default:"root"`
		Password string `env:"STORAGE_PASSWORD" env-default:"1qazxsw2"`
		Host     string `env:"STORAGE_HOST" env-default:"localhost"`
		Port     string `env:"STORAGE_PORT" env-default:"5432"` // 5432 postgresql, 27017 mongodb
		Database string `env:"STORAGE_DATABASE" env-default:"db_issue_album"`
		// Mongo use
		Collections      string `env:"STORAGE_COLLECTIONS" env-default:"col_issues"`
		CollectionsUsers string `env:"STORAGE_COLLECTIONS_USERS" env-default:"col_users"`
	}
	// MessageQueue
	MessageQueue struct {
		Enable        bool   `env:"MQ_ENABLE" env-default:"true"`
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
	Session struct {
		SessionStorageType string `env:"SESSION_STORAGE_TYPE" env-default:"postgres"` // cookie, memory, memcached,
		// mongo, postgres
		SessionName string `env:"SESSION_COOKIES_SESSION_NAME" env-default:"gin-session"`
		Cookies     struct {
			SessionSecretKey string `env:"SESSION_COOKIES_SESSION_SECRET_KEY" env-default:"sdfgerfsd3543g"`
		}
		Memcached struct {
			MemcachedHost string `env:"SESSION_MEMCACHED_HOST" env-default:"localhost"`
			MemcachedPort string `env:"SESSION_MEMCACHED_PORT" env-default:"11211"`
		}
		Mongodb struct {
			MongoHost     string `env:"SESSION_MONGO_HOST" env-default:"localhost"`
			MongoPort     string `env:"SESSION_MONGO_PORT" env-default:"27017"`
			MongoDatabase string `env:"SESSION_MONGO_DATABASE" env-default:"session"`
			MongoUser     string `env:"SESSION_MONGO_USERNAME" env-default:"root"`
			MongoPass     string `env:"SESSION_MONGO_PASSWORD" env-default:"1qazxsw2"`
		}
		Postgresql struct {
			PostgresqlHost     string `env:"SESSION_POSTGRESQL_HOST" env-default:"localhost"`
			PostgresqlPort     string `env:"SESSION_POSTGRESQL_PORT" env-default:"5432"`
			PostgresqlDatabase string `env:"SESSION_POSTGRESQL_DATABASE" env-default:"session"`
			PostgresqlUser     string `env:"SESSION_POSTGRESQL_USER" env-default:"root"`
			PostgresqlPass     string `env:"SESSION_POSTGRESQL_PASS" env-default:"1qazxsw2"`
		}
	}
}

// GetConfig returns the singleton instance of the configuration.
func GetConfig() *Config {
	cfgManager := getConfigManager()

	cfgManager.once.Do(func() {
		log.Info("gathering config")

		cfgManager.instance = &Config{}
	})

	if err := cleanenv.ReadEnv(cfgManager.instance); err != nil {
		helpText := "The Art of Development - Monolith Notes System"
		help, _ := cleanenv.GetDescription(cfgManager.instance, &helpText)
		log.Debug(help)
		log.Errorf("Error reading environment variables: %v", err)
	}

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
