package cashing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type CachingInterface interface {
	CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error
	CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error)
	GetTrackInCache(ctx context.Context, key string, result interface{}) error
	SetTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	UpdateTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteTrackInCache(ctx context.Context, key string) error
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

func (c *CachingRepository) GetTrackInCache(ctx context.Context, key string, result interface{}) error {
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Data not found in cache
		return fmt.Errorf("cache miss")
	} else if err != nil {
		return err
	}

	// Unmarshal JSON data into the result object
	return json.Unmarshal([]byte(data), result)
}

func (c *CachingRepository) SetTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.redisClient.Set(ctx, key, data, ttl).Err()
}

// UpdateTrackInCache updates existing track data in Redis if the track data has changed.
func (c *CachingRepository) UpdateTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Delete existing cache first
	err := c.DeleteTrackInCache(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete existing cache: %w", err)
	}

	// Set the updated value in cache
	return c.SetTrackInCache(ctx, key, value, ttl)
}

func (c *CachingRepository) DeleteTrackInCache(ctx context.Context, key string) error {
	return c.redisClient.Del(ctx, key).Err()
}
