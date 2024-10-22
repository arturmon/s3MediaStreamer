package connect

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// InitRedis initializes the Redis client.
func InitRedis(ctx context.Context, cfg *model.Config, logger *logs.Logger, dbIndex int) (*redis.Client, error) {
	logger.Info("Starting redis Connection...")
	// Create redis.Options
	redisOptions := redis.Options{
		Addr:     cfg.Storage.Caching.Address,
		Password: cfg.Storage.Caching.Password,
		DB:       dbIndex,
	}
	// Create logs.LoggerMessageConnect
	logFields := []model.LogField{
		{Key: "TypeConnect", Value: "Redis", Mask: ""},
		{Key: "DB", Value: strconv.Itoa(dbIndex), Mask: ""},
		{Key: "Addr", Value: cfg.Storage.Caching.Address, Mask: ""},
		{Key: "Password", Value: cfg.Storage.Caching.Password, Mask: "password"},
	}
	loggerMsg := logs.NewLoggerMessageConnect(logFields)

	// Use logger to log the redis.Options using formatterRedis
	logger.Info("Starting Redis Connection...")
	// Create Redis client with options
	redisClient := redis.NewClient(&redisOptions)

	// Ping the Redis server to ensure the connection is established
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Slog().Error("(Redis: Auth User) Failed to connect", "connection", loggerMsg.MaskFields())
		return nil, err
	}

	// Log successful connection
	logger.Slog().Info("(Redis: Auth User) Successfully to connect", "connection", loggerMsg.MaskFields())
	return redisClient, nil
}
