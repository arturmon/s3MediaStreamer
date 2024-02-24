package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (a *WebApp) setupAppRoutesV1() {
	v1 := a.router.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", a.Register)
			users.OPTIONS("/login", handleOptions)
			users.POST("/login", a.Login)
			users.GET("/me", a.User)
			users.POST("/delete", a.DeleteUser)
			users.POST("/logout", a.Logout)
			users.POST("/refresh", a.refreshTokenHandler)
			otp := users.Group("/otp")
			{
				otp.POST("/generate", a.GenerateOTP)
				otp.POST("/verify", a.VerifyOTP)
				otp.POST("/validate", a.ValidateOTP)
				otp.POST("/disable", a.DisableOTP)
			}
		}
		tracks := v1.Group("/tracks")
		{
			tracks.GET("", a.GetAllTracks)
			tracks.GET("/:code", a.GetTrackByID)
		}
		a.logger.Info("swagger docs initializing")
		swagger := v1.Group("/swagger")
		{
			swagger.GET("", func(c *gin.Context) {
				c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			})
			swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
		audio := v1.Group("/audio")
		{
			audio.GET("/stream/:segment", a.StreamM3U)
			audio.GET("/:playlist_id", a.Audio)
		}
		playlist := v1.Group("/playlist")
		{
			playlist.POST("/:playlist_id/add/track/:track_id", a.AddToPlaylist)
			playlist.DELETE("/:playlist_id/remove/track/:track_id", a.RemoveFromPlaylist)
			playlist.DELETE("/:playlist_id/clear", a.ClearPlaylist)
			playlist.POST("/create", a.CreatePlaylist)
			playlist.DELETE("/delete/:id", a.DeletePlaylist)
			playlist.POST("/:playlist_id/set", a.SetFromPlaylist)
			playlist.GET(":playlist_id/get", a.ListTracksFromPlaylist)
		}

		player := v1.Group("/player")
		{
			player.GET("/play/:playlist_id", a.Play)
		}
	}
}
