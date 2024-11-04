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
	var err error

	logger.Info("Initializing session store...")
	var logFields []model.LogField
	// Initialize session
	switch cfg.Session.SessionStorageType {
	case "cookie":
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
		logFields = []model.LogField{
			{Key: "TypeConnect", Value: "Cookie Session", Mask: ""},
			{Key: "Secret", Value: cfg.Session.Cookies.SessionSecretKey, Mask: "password"},
		}
	case "memory":
		store = memstore.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
		logFields = []model.LogField{
			{Key: "TypeConnect", Value: "Memory Session", Mask: ""},
			{Key: "Secret", Value: cfg.Session.Cookies.SessionSecretKey, Mask: "password"},
		}
	case "memcached":
		memcachedURL := cfg.Session.Memcached.MemcachedHost + ":" + cfg.Session.Memcached.MemcachedPort
		store = memcached.NewStore(memcache.New(memcachedURL), "", []byte(cfg.Session.Cookies.SessionSecretKey))
		logFields = []model.LogField{
			{Key: "TypeConnect", Value: "Memcached Session", Mask: ""},
			{Key: "Addr", Value: memcachedURL, Mask: ""},
			{Key: "Secret", Value: cfg.Session.Cookies.SessionSecretKey, Mask: "password"},
		}
	case "mongo":
		mongoURL := "mongodb://" + cfg.Session.Mongodb.MongoUser + ":" + cfg.Session.Mongodb.MongoPass +
			"@" + cfg.Session.Mongodb.MongoHost + ":" + cfg.Session.Mongodb.MongoPort
		mongoOptions := options.Client().ApplyURI(mongoURL)
		client, mongoErr := mongo.Connect(ctx, mongoOptions) // Use Connect instead of NewClient
		logFields = []model.LogField{
			{Key: "TypeConnect", Value: "Mongo Session", Mask: ""},
			{Key: "MongoDB", Value: cfg.Session.Mongodb.MongoDatabase, Mask: ""},
			{Key: "Host", Value: cfg.Session.Mongodb.MongoHost, Mask: ""},
			{Key: "User", Value: cfg.Session.Mongodb.MongoUser, Mask: ""},
			{Key: "Port", Value: cfg.Session.Mongodb.MongoPort, Mask: ""},
		}
		if mongoErr != nil {
			logger.Errorf("Failed to create Mongo client: %v", mongoErr)
			err = mongoErr
		} else {
			c := client.Database(cfg.Session.Mongodb.MongoDatabase).Collection("sessions")
			store = mongodriver.NewStore(c, mongodriverMaxIdle, true, []byte(cfg.Session.Cookies.SessionSecretKey))
		}
	case "postgres":
		sslMode := "sslmode=disable"
		dsn := url.URL{
			Scheme:   "postgresql",
			User:     url.UserPassword(cfg.Session.Postgresql.PostgresqlUser, cfg.Session.Postgresql.PostgresqlPass),
			Host:     net.JoinHostPort(cfg.Session.Postgresql.PostgresqlHost, cfg.Session.Postgresql.PostgresqlPort),
			Path:     cfg.Session.Postgresql.PostgresqlDatabase,
			RawQuery: sslMode, // This enables SSL/TLS
		}
		logFields = []model.LogField{
			{Key: "TypeConnect", Value: "Postgres Session", Mask: ""},
			{Key: "DB", Value: cfg.Session.Postgresql.PostgresqlDatabase, Mask: ""},
			{Key: "Other", Value: sslMode, Mask: ""},
			{Key: "Addr", Value: net.JoinHostPort(cfg.Session.Postgresql.PostgresqlHost, cfg.Session.Postgresql.PostgresqlPort), Mask: ""},
			{Key: "User", Value: cfg.Session.Postgresql.PostgresqlUser, Mask: ""},
			{Key: "Password", Value: cfg.Session.Postgresql.PostgresqlPass, Mask: "password"},
		}
		db, postgresErr := sql.Open("postgres", dsn.String())
		if postgresErr != nil {
			logger.Errorf("Failed to create Postgres database connection: %v", postgresErr)
			err = postgresErr
		}
		db.SetMaxOpenConns(SetMaxOpenConns)
		db.SetMaxIdleConns(SetMaxIdleConns)
		store, postgresErr = postgres.NewStore(db, []byte(cfg.Session.Cookies.SessionSecretKey))
		if postgresErr != nil {
			logger.Errorf("Failed to create Postgres store: %v", postgresErr)
			err = postgresErr
		}

	default:
		logger.Warnf("Unknown session storage type '%s', defaulting to cookie store", cfg.Session.SessionStorageType)
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	}

	loggerMsg := logs.NewLoggerMessageConnect(logFields)

	store.Options(sessions.Options{
		MaxAge:   sessionMaxAge,
		Path:     "/", // Set the cookie path to "/"
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	if err != nil {
		logger.Slog().Error("(Session) Failed to connect", "connection", loggerMsg.MaskFields())
		return store, err
	}

	// logger.Info("Session store initialized successfully")
	logger.Slog().Info("(Session) Successfully to connect", "connection", loggerMsg.MaskFields())
	return store, nil
}
