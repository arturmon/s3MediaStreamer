package router

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/handlers"
	"s3MediaStreamer/app/internal/app"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(ctx context.Context, app *app.App, allHandlers *handlers.Handlers) {
	setupStaticFiles(app)
	setupSystemRoutes(ctx, app, allHandlers)

	// Initialize the Redis cache store
	cacheURL, err := InitCacheURL(ctx, app)
	if err != nil {
		app.Logger.Fatalf("Failed to initialize Redis cache: %v", err)
		return
	}
	ttl := time.Duration(app.Cfg.Storage.Caching.Expiration) * time.Hour

	v1 := app.REST.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", allHandlers.User.Register)
			users.OPTIONS("/login", HandleOptions)
			users.POST("/login", allHandlers.User.Login)
			users.GET("/me", allHandlers.User.User)
			users.POST("/delete", allHandlers.User.DeleteUser)
			users.POST("/logout", allHandlers.User.Logout)
			users.POST("/refresh", allHandlers.User.RefreshTokenHandler)
			otp := users.Group("/otp")
			{
				otp.POST("/generate", allHandlers.Otp.GenerateOTP)
				otp.POST("/verify", allHandlers.Otp.VerifyOTP)
				otp.POST("/validate", allHandlers.Otp.ValidateOTP)
				otp.POST("/disable", allHandlers.Otp.DisableOTP)
			}
		}
		tracks := v1.Group("/tracks")
		{
			if app.Cfg.Storage.Caching.Enabled {
				tracks.GET("", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Track.GetAllTracks)
				tracks.GET("/:code", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Track.GetTrackByID)
			} else {
				tracks.GET("", allHandlers.Track.GetAllTracks)
				tracks.GET("/:code", allHandlers.Track.GetTrackByID)
			}
		}
		app.Logger.Info("swagger docs initializing")
		swagger := v1.Group("/swagger")
		{
			swagger.GET("", func(c *gin.Context) {
				c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			})
			swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
		audio := v1.Group("/audio")
		{
			audio.GET("/stream/:segment", allHandlers.Audio.StreamM3U)
			audio.GET("/:playlist_id", allHandlers.Audio.Audio)
		}
		playlist := v1.Group("/playlist")
		{
			playlist.POST("/create", allHandlers.Playlist.CreatePlaylist)
			playlist.DELETE("/:playlist_id", allHandlers.Playlist.DeletePlaylist)
			playlist.POST("/:playlist_id/:track_id", allHandlers.Playlist.AddToPlaylist)
			// Conditionally add caching middleware based on configuration
			if app.Cfg.Storage.Caching.Enabled {
				playlist.GET("/:playlist_id", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Playlist.ListTracksFromPlaylist)
				playlist.GET("/get", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Playlist.ListPlaylists)
			} else {
				playlist.GET("/:playlist_id", allHandlers.Playlist.ListTracksFromPlaylist)
				playlist.GET("/get", allHandlers.Playlist.ListPlaylists)
			}
			playlist.POST("/:playlist_id", allHandlers.Playlist.SetFromPlaylist)
			playlist.DELETE("/:playlist_id/:track_id", allHandlers.Playlist.RemoveFromPlaylist)
			playlist.DELETE("/:playlist_id/clear", allHandlers.Playlist.ClearPlaylist)
		}
	}
}

func setupStaticFiles(app *app.App) {
	app.REST.StaticFile("/favicon.ico", "./favicon.ico")
}

func setupSystemRoutes(ctx context.Context, app *app.App, allHandlers *handlers.Handlers) {
	sessionName := app.Cfg.Session.SessionName
	app.REST.Use(sessions.Sessions(sessionName, app.Service.InitRepo.InitConnect.SessionStore))

	app.Logger.Info("heartbeat metric initializing")
	app.REST.GET("/health/liveness", func(c *gin.Context) {
		allHandlers.Health.LivenessGET(c, app.Service.Health) // Pass the healthMetrics to HealthGET.
	})
	app.REST.GET("/health/readiness", func(c *gin.Context) {
		allHandlers.Health.ReadinessGET(c, app.Service.Health) // Pass the healthMetrics to HealthGET.
	})

	app.REST.GET("/job/status", allHandlers.Job.JobStatus)

	app.REST.Use(cors.New(handlers.ConfigCORS(app.Cfg.AppConfig.Web.CorsAllowOrigins)))

	app.REST.Use(app.Service.ACL.ExtractUserRole(ctx, app.Logger))
	app.REST.Use(app.Service.ACL.NewAuthorizerWithRoleExtractor(app.Service.ACL.AccessControl, app.Logger, func(c *gin.Context) string {
		if role, ok := c.Get("userRole"); ok {
			return role.(string)
		}
		return "anonymous" // Default role
	}))
}

func HandleOptions(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// InitCacheUrl initializes the Redis cache store and returns a persist.RedisStore instance.
func InitCacheURL(ctx context.Context, app *app.App) (*persist.RedisStore, error) {
	setDB := 1
	redisClient := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     app.Cfg.Storage.Caching.Address,
		Password: app.Cfg.Storage.Caching.Password,
		DB:       setDB,
	})

	// Ping Redis to ensure the connection is working
	if err := redisClient.Ping(ctx).Err(); err != nil {
		app.Logger.Errorf("(Redis: caching URL) Failed to connect redis at %s, errors: %v", app.Cfg.Storage.Caching.Address, err)
		return nil, err
	}
	// Log successful connection
	app.Logger.Infof("(Redis: caching URL) Successfully connected to Redis at %s using DB index %d", app.Cfg.Storage.Caching.Address, setDB)

	redisStore := persist.NewRedisStore(redisClient)
	return redisStore, nil
}
