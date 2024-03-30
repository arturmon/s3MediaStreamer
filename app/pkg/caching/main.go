package caching

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// Assuming you have a Redis client configured
var redisClient *redis.Client

// InitRedis initializes the Redis client.
func InitRedis(ctx context.Context, redisAddr, redisPassword string) *CachingStruct {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // Use default DB
	})

	// Ping the Redis server to ensure the connection is established
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil
	}

	return &CachingStruct{
		redisClient: redisClient,
	}
}

// CachePasswordVerificationInRedis caches the result of the password verification in Redis.
func (c *CachingStruct) CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error {
	// Convert the verification result (bool) to a string to store it in Redis.
	verificationResult := strconv.FormatBool(verificationSuccess)
	return redisClient.Set(ctx, "password_verification:"+userID, verificationResult, 24*time.Hour).Err()
}

// CheckPasswordVerificationInRedis checks if the result of the password verification is cached in Redis.
func (c *CachingStruct) CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error) {
	result, err := redisClient.Get(ctx, "password_verification:"+userID).Result()
	if err == redis.Nil {
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
