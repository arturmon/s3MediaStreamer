package gin

import (
	"context"
	"database/sql"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const sessionMaxAge = 60 * 60 * 24

func initSession(ctx context.Context, router *gin.Engine, cfg *config.Config, logger *logging.Logger) {
	var store sessions.Store

	// Initialize session
	switch cfg.Session.SessionStorageType {
	case "cookie":
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	case "memory":
		store = memstore.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	case "memcached":
		memcachedURL := cfg.Session.Memcached.MemcachedHost + ":" + cfg.Session.Memcached.MemcachedPort
		store = memcached.NewStore(memcache.New(memcachedURL), "", []byte(cfg.Session.Cookies.SessionSecretKey))
	case "mongo":
		mongoURL := "mongodb://" + cfg.Session.Mongodb.MongoUser + ":" + cfg.Session.Mongodb.MongoPass +
			"@" + cfg.Session.Mongodb.MongoHost + ":" + cfg.Session.Mongodb.MongoPort
		mongoOptions := options.Client().ApplyURI(mongoURL)
		client, err := mongo.NewClient(mongoOptions)
		if err != nil {
			logger.Errorf("Error creating Mongo store: %v", err)
		} else {
			connectErr := client.Connect(ctx)
			if connectErr != nil {
				logger.Errorf("Error creating Mongo store: %v", connectErr)
			} else {
				c := client.Database(cfg.Session.Mongodb.MongoDatabase).Collection("sessions")
				store = mongodriver.NewStore(c, 3600, true, []byte(cfg.Session.Cookies.SessionSecretKey))
			}
		}
	case "postgres":
		var postgresURL string
		postgresURL = fmt.Sprintf("%s:%s", cfg.Session.Postgresql.PostgresqlHost, cfg.Session.Postgresql.PostgresqlPort)
		postgresURL = fmt.Sprintf("postgresql://%s:%s@%s", cfg.Session.Postgresql.PostgresqlUser, cfg.Session.Postgresql.PostgresqlPass, postgresURL)
		postgresURL = fmt.Sprintf("%s/%s", postgresURL, cfg.Session.Postgresql.PostgresqlDatabase)
		postgresURL = postgresURL + "?sslmode=disable"
		db, err := sql.Open("postgres", postgresURL)
		if err != nil {
			logger.Errorf("Error creating Postgres store: %v", err)
		}
		store, err = postgres.NewStore(db, []byte(cfg.Session.Cookies.SessionSecretKey))
		if err != nil {
			logger.Errorf("Error creating Postres store: %v", err)
		}
	default:
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	}

	logger.Infof("session use storage: %s", cfg.Session.SessionStorageType)
	store.Options(sessions.Options{MaxAge: sessionMaxAge}) // expire in a day
	sessionName := cfg.Session.SessionName
	router.Use(sessions.Sessions(sessionName, store))
}

func countSession(c *gin.Context) {
	session := sessions.Default(c)
	var count int
	v := session.Get("count")
	if v == nil {
		count = 0
	} else {
		if val, ok := v.(int); ok {
			count = val
			count++
		} else {
			logrus.Errorf("Error converting session value to int")
			return
		}
	}
	session.Set("count", count)

	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
	}
}
