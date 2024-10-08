package connect

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/go-redis/redis/v8"
)

// InitRedis initializes the Redis client.
func InitRedis(ctx context.Context, cfg *model.Config, logger *logs.Logger, dbIndex int) (*redis.Client, error) {
	logger.Info("Starting redis Connection...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Storage.Caching.Address,
		Password: cfg.Storage.Caching.Password,
		DB:       dbIndex,
	})
	// Ping the Redis server to ensure the connection is established
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Errorf("(Redis: Auth User) Failed to connect redis at %s, errors: %v", cfg.Storage.Caching.Address, err)
		return nil, err
	}

	// Log successful connection
	logger.Infof("(Redis: Auth User) Successfully connected to Redis at %s using DB index %d", cfg.Storage.Caching.Address, dbIndex)

	return redisClient, nil
}
