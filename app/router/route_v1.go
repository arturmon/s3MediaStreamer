package router

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/handlers"
	"s3MediaStreamer/app/internal/app"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"strconv"
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

	// Initialize Redis cache store
	cacheURL, ttl := initCache(ctx, app)

	v1 := app.REST.Group("/v1")

	// Users routes
	initUserRoutes(v1.Group("/users"), allHandlers)

	// Tracks routes
	initTrackRoutes(v1.Group("/tracks"), allHandlers, cacheURL, ttl, app.Cfg.Storage.Caching.Enabled)

	// Swagger docs
	initSwaggerRoutes(v1.Group("/swagger"))

	// Audio routes
	initAudioRoutes(v1.Group("/audio"), allHandlers)

	// Playlist routes
	initPlaylistRoutes(v1.Group("/playlist"), allHandlers, cacheURL, ttl, app.Cfg.Storage.Caching.Enabled)
}

// User-related routes
func initUserRoutes(users *gin.RouterGroup, allHandlers *handlers.Handlers) {
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

// Track-related routes
func initTrackRoutes(tracks *gin.RouterGroup, allHandlers *handlers.Handlers, cacheURL *persist.RedisStore, ttl time.Duration, cacheEnabled bool) {
	if cacheEnabled {
		tracks.GET("", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Track.GetAllTracks)
		tracks.GET("/:code", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Track.GetTrackByID)
	} else {
		tracks.GET("", allHandlers.Track.GetAllTracks)
		tracks.GET("/:code", allHandlers.Track.GetTrackByID)
	}
}

// Audio routes
func initAudioRoutes(audio *gin.RouterGroup, allHandlers *handlers.Handlers) {
	audio.GET("/stream/:segment", allHandlers.Audio.StreamM3U)
	audio.GET("/:playlist_id", allHandlers.Audio.Audio)
}

// Playlist-related routes
func initPlaylistRoutes(playlist *gin.RouterGroup, allHandlers *handlers.Handlers, cacheURL *persist.RedisStore, ttl time.Duration, cacheEnabled bool) {
	playlist.POST("/create", allHandlers.Playlist.CreatePlaylist)

	playlist.DELETE("/:playlist_id", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.DeletePlaylist))
	playlist.POST("/:playlist_id/:track_id", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.AddToPlaylist))

	if cacheEnabled {
		playlist.GET("/:playlist_id", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.ListTracksFromPlaylist))
		playlist.GET("/get", cache.CacheByRequestURI(cacheURL, ttl), allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.ListPlaylists))
	} else {
		playlist.GET("/:playlist_id", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.ListTracksFromPlaylist))
		playlist.GET("/get", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.ListPlaylists))
	}

	playlist.POST("/:playlist_id/tracks", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.AddTracksToPlaylist))
	playlist.DELETE("/:playlist_id/:track_id", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.RemoveFromPlaylist))
	playlist.DELETE("/:playlist_id/clear", allHandlers.Wrapper.WrapWithUserCheck(allHandlers.Playlist.ClearPlaylist))
}

// Swagger routes
func initSwaggerRoutes(swagger *gin.RouterGroup) {
	swagger.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

	app.REST.Use(app.Service.ACL.ExtractUserRole(app.Logger))
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
