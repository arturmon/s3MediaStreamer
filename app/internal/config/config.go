package config

import (
	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var AppHealth = false

// Album album represents data about a record album.
type Album struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" pg:"default:now()"`
	Title       string    `json:"title" bson:"title"`
	Artist      string    `json:"artist" bson:"artist"`
	Price       float64   `json:"price" bson:"price"`
	Code        string    `json:"code" bson:"code"`
	Description string    `json:"description" bson:"description"`
	Completed   bool      `json:"completed" bson:"completed"`
}

type User struct {
	Id       uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid"`
	Name     string    `json:"name" bson:"name"`
	Email    string    `json:"email" bson:"email"`
	Password []byte    `json:"-" bson:"password"`
}

type Config struct {
	//IsDebug bool `env:"IS_DEBUG" env-default:"false"`
	//IsDevelopment bool `env:"IS_DEV" env-default:"false"`
	Listen struct {
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
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Print("gather config")

		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			helpText := "The Art of Development - Monolith Notes System"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
