package router

import (
	"context"
	"s3MediaStreamer/app/internal/app"
	"time"

	"github.com/chenyahui/gin-cache/persist"
)

// Initialize cache and return cache URL and TTL
func initCache(ctx context.Context, app *app.App) (*persist.RedisStore, time.Duration) {
	cacheURL, err := InitCacheURL(ctx, app)
	if err != nil {
		app.Logger.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	ttl := time.Duration(app.Cfg.Storage.Caching.Expiration) * time.Hour
	return cacheURL, ttl
}
