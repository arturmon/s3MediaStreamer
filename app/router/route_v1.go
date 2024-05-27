package router

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/handlers"
	"s3MediaStreamer/app/internal/app"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(ctx context.Context, app *app.App, allHandlers *handlers.Handlers) {
	setupStaticFiles(app)
	setupSystemRoutes(ctx, app, allHandlers)

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
			otp := users.Group("/otp_handler")
			{
				otp.POST("/generate", allHandlers.Otp.GenerateOTP)
				otp.POST("/verify", allHandlers.Otp.VerifyOTP)
				otp.POST("/validate", allHandlers.Otp.ValidateOTP)
				otp.POST("/disable", allHandlers.Otp.DisableOTP)
			}
		}
		tracks := v1.Group("/tracks")
		{
			tracks.GET("", allHandlers.Track.GetAllTracks)
			tracks.GET("/:code", allHandlers.Track.GetTrackByID)
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
			playlist.GET("/:playlist_id", allHandlers.Playlist.ListTracksFromPlaylist)
			playlist.POST("/:playlist_id", allHandlers.Playlist.SetFromPlaylist)
			playlist.DELETE("/:playlist_id/:track_id", allHandlers.Playlist.RemoveFromPlaylist)
			playlist.DELETE("/:playlist_id/clear", allHandlers.Playlist.ClearPlaylist)
			playlist.GET("/get", allHandlers.Playlist.ListPlaylists)
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
	app.REST.GET("/health_handler/liveness", func(c *gin.Context) {
		allHandlers.Health.LivenessGET(c, app.Service.Health) // Pass the healthMetrics to HealthGET.
	})
	app.REST.GET("/health_handler/readiness", func(c *gin.Context) {
		allHandlers.Health.ReadinessGET(c, app.Service.Health) // Pass the healthMetrics to HealthGET.
	})

	app.REST.GET("/job/status", allHandlers.Job.JobStatus)

	app.REST.Use(app.Service.Acl.ExtractUserRole(ctx, app.Logger))
	app.REST.Use(app.Service.Acl.NewAuthorizerWithRoleExtractor(app.Service.Acl.AccessControl, app.Logger, func(c *gin.Context) string {
		if role, ok := c.Get("userRole"); ok {
			return role.(string)
		}
		return "anonymous" // Default role
	}))
}

func HandleOptions(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}
