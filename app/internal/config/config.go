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

// Album album represents data about a record album.
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

type User struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Email    string             `json:"email" bson:"email"`
	Password []byte             `json:"-" bson:"password"`
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
	// STORAGE_TYPE: mongodb, postgresql
	Storage struct {
		Type     string `env:"STORAGE_TYPE" env-default:"mongodb"`
		Username string `env:"STORAGE_USERNAME" env-default:"root"`
		Password string `env:"STORAGE_PASSWORD" env-default:"1qazxsw2"`
		Host     string `env:"STORAGE_HOST" env-default:"localhost"`
		Port     string `env:"STORAGE_PORT" env-default:"27017"`
		// posdtresq 'db_issue_album'
		Database string `env:"STORAGE_DATABASE" env-default:"db_issue_album"`
		// Mongo use
		Collections      string `env:"STORAGE_COLLECTIONS" env-default:"col_issues"`
		CollectionsUsers string `env:"STORAGE_COLLECTIONS_USERS" env-default:"col_users"`
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
