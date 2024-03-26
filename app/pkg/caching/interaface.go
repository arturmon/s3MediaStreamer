package caching

import (
	"github.com/go-redis/redis/v8"
)

type CachingStruct struct {
	redisClient *redis.Client
}
