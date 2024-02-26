package gin

import (
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"net"
	"net/http"
	"net/url"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/gin-contrib/sessions/cookie"

	"github.com/google/uuid"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	sessionMaxAge      = 60 * 60 * 24
	mongodriverMaxIdle = 3600 // Define a named constant for better readability
	SetMaxOpenConns    = 10
	SetMaxIdleConns    = 5
)

func initSession(ctx context.Context, router *gin.Engine, cfg *config.Config, logger *logging.Logger) {
	gob.Register(uuid.UUID{})
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
		client, err := mongo.Connect(ctx, mongoOptions) // Use Connect instead of NewClient
		if err != nil {
			logger.Errorf("Error creating Mongo store: %v", err)
		} else {
			c := client.Database(cfg.Session.Mongodb.MongoDatabase).Collection("sessions")
			store = mongodriver.NewStore(c, mongodriverMaxIdle, true, []byte(cfg.Session.Cookies.SessionSecretKey))
		}
	case "postgres":
		dsn := url.URL{
			Scheme:   "postgresql",
			User:     url.UserPassword(cfg.Session.Postgresql.PostgresqlUser, cfg.Session.Postgresql.PostgresqlPass),
			Host:     net.JoinHostPort(cfg.Session.Postgresql.PostgresqlHost, cfg.Session.Postgresql.PostgresqlPort),
			Path:     cfg.Session.Postgresql.PostgresqlDatabase,
			RawQuery: "sslmode=disable", // This enables SSL/TLS
		}
		db, err := sql.Open("postgres", dsn.String())
		if err != nil {
			logger.Errorf("Error creating Postgres store: %v", err)
			return
		}
		db.SetMaxOpenConns(SetMaxOpenConns)
		db.SetMaxIdleConns(SetMaxIdleConns)
		store, err = postgres.NewStore(db, []byte(cfg.Session.Cookies.SessionSecretKey))
		if err != nil {
			logger.Errorf("Error creating Postres store: %v", err)
			return
		}
	default:
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	}
	store.Options(sessions.Options{
		MaxAge:   sessionMaxAge,
		Path:     "/", // Set the cookie path to "/"
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	sessionName := cfg.Session.SessionName
	router.Use(sessions.Sessions(sessionName, store))
	logger.Infof("session use storage: %s", cfg.Session.SessionStorageType)
}

func setSessionData(c *gin.Context, data map[string]interface{}) error {
	session := sessions.Default(c)

	// Set the values in the session
	for key, value := range data {
		session.Set(key, value)
	}

	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
		return err
	}
	return nil
}

func getSessionKey(c *gin.Context, key string) (interface{}, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return nil, errors.New("session value is nil")
	}
	return value, nil
}

func logoutSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
		return err
	}
	return nil
}
