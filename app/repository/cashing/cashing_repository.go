package cashing

import (
	"context"
	"errors"
	"s3MediaStreamer/app/internal/logs"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type CachingInterface interface {
	CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error
	CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error)
}

type CachingRepository struct {
	redisClient *redis.Client
}

// Assuming you have a Redis client configured.
// var redisClient *redis.Client

// InitRedisRepository initializes the Redis client.
func InitRedisRepository(logger *logs.Logger, redisClient *redis.Client) *CachingRepository {
	logger.Info("Starting Redis repository...")
	return &CachingRepository{
		redisClient: redisClient,
	}
}

// CachePasswordVerificationInRedis caches the result of the password verification in Redis.
func (c *CachingRepository) CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error {
	// Convert the verification result (bool) to a string to store it in Redis.
	verificationResult := strconv.FormatBool(verificationSuccess)

	return c.redisClient.Set(ctx, "password_verification:"+userID, verificationResult, time.Duration(expiration)*time.Hour).Err()
}

// CheckPasswordVerificationInRedis checks if the result of the password verification is cached in Redis.
func (c *CachingRepository) CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error) {
	result, err := c.redisClient.Get(ctx, "password_verification:"+userID).Result()
	if errors.Is(err, redis.Nil) {
		// The result is not in the cache.
		return false, false, nil
	} else if err != nil {
		// There was an error fetching data from Redis.
		return false, false, err
	}

	// Convert the result from a string back to a bool.
	verificationSuccess, err := strconv.ParseBool(result)
	if err != nil {
		// There was an error converting the result.
		return false, false, err
	}

	// The result was found and successfully converted.
	return true, verificationSuccess, nil
}
