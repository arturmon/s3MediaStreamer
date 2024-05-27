package connect

import (
	"context"
	"database/sql"
	"encoding/gob"
	"net"
	"net/http"
	"net/url"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	sessionMaxAge      = 60 * 60 * 24
	mongodriverMaxIdle = 3600 // Define a named constant for better readability
	SetMaxOpenConns    = 10
	SetMaxIdleConns    = 5
)

func InitSession(ctx context.Context, cfg *model.Config, logger *logs.Logger) (sessions.Store, error) {
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
			return nil, err
		}
		db.SetMaxOpenConns(SetMaxOpenConns)
		db.SetMaxIdleConns(SetMaxIdleConns)
		store, err = postgres.NewStore(db, []byte(cfg.Session.Cookies.SessionSecretKey))
		if err != nil {
			logger.Errorf("Error creating Postres store: %v", err)
			return nil, err
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

	logger.Info("session initializing")

	logger.Infof("session use storage: %s", cfg.Session.SessionStorageType)
	return store, nil
}
