package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var AppHealth = false

// album represents data about a record album.
type Album struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	Title       string             `json:"title" bson:"title"`
	Artist      string             `json:"artist" bson:"artist"`
	Price       float64            `json:"price" bson:"price"`
	Code        string             `json:"code" bson:"code"`
	Description string             `json:"description" bson:"description"`
	Completed   bool               `json:"completed" bson:"completed"`
}

type Config struct {
	IsDebug       bool `env:"IS_DEBUG" env-default:"false"`
	IsDevelopment bool `env:"IS_DEV" env-default:"false"`
	Listen        struct {
		BindIP string `env:"BIND_IP" env-default:"0.0.0.0"`
		Port   string `env:"PORT" env-default:"10000"`
	}
	AppConfig struct {
		LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	}
	Storage struct {
		Type    string `env:"STORAGE_TYPE" env-default:"mongo"`
		MongoDB struct {
			Host        string `env:"MONGO_HOST" env-default:"localhost""`
			Port        string `env:"MONGO_PORT" env-default:"27017"`
			Database    string `env:"MONGO_DATABASE" env-default:"db_issue_album"`
			Collections string `env:"MONGO_COL" env-default:"col_issues"`
			Username    string `env:"MONGO_USERNAME" env-default:"root"`
			Password    string `env:"MONGO_PASSWORD" env-default:"1qazxsw2"`
		}
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
