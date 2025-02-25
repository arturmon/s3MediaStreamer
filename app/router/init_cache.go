package router

import (
	"context"
	"s3MediaStreamer/app/internal/app"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"strconv"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/go-redis/redis/v8"
)

// Initialize cache and return cache URL and TTL.
func initCache(ctx context.Context, app *app.App) (*persist.RedisStore, time.Duration) {
	cacheURL, err := InitCacheURL(ctx, app)
	if err != nil {
		app.Logger.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	ttl := time.Duration(app.Cfg.Storage.Caching.Expiration) * time.Hour
	return cacheURL, ttl
}

// InitCacheURL initializes the Redis cache store and returns a persist.RedisStore instance.
func InitCacheURL(ctx context.Context, app *app.App) (*persist.RedisStore, error) {
	setDB := 1
	redisClient := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     app.Cfg.Storage.Caching.Address,
		Password: app.Cfg.Storage.Caching.Password,
		DB:       setDB,
	})
	// Create logs.LoggerMessageConnect
	logFields := []model.LogField{
		{Key: "TypeConnect", Value: "Redis", Mask: ""},
		{Key: "DB", Value: strconv.Itoa(setDB), Mask: ""},
		{Key: "Addr", Value: app.Cfg.Storage.Caching.Address, Mask: ""},
		{Key: "Password", Value: app.Cfg.Storage.Caching.Password, Mask: "password"},
	}
	loggerMsg := logs.NewLoggerMessageConnect(logFields)

	// Ping Redis to ensure the connection is working
	if err := redisClient.Ping(ctx).Err(); err != nil {
		app.Logger.Slog().Error("(Redis: Auth User) Failed to connect", "connection", loggerMsg.MaskFields())
		return nil, err
	}
	// Log successful connection
	app.Logger.Slog().Info("(Redis: Auth User) Successfully to connect", "connection", loggerMsg.MaskFields())
	redisStore := persist.NewRedisStore(redisClient)
	return redisStore, nil
}
