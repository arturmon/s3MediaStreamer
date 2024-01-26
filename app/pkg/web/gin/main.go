package gin

import (
	"context"
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/app/pkg/monitoring"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"

	"github.com/casbin/casbin/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gin-gonic/gin"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

type AppInterface interface {
	Run()
}

type WebApp struct {
	cfg           *config.Config
	logger        *logging.Logger
	router        *gin.Engine
	storage       *model.DBConfig
	healthMetrics *monitoring.HealthMetrics
	metrics       *monitoring.Metrics
	enforcer      *casbin.Enforcer
}

func NewAppUseGin(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*WebApp, error) {
	logger.Info("router initializing")
	// Gin instance
	switch cfg.AppConfig.GinMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	logger.Info("setup CORS")
	router.Use(cors.New(ConfigCORS()))
	router.Use(gin.Recovery())

	router.Use(LoggingMiddleware(LoggingMiddlewareAdapter(logger)))

	logger.Info("prometheus initializing")
	p := ginPrometheus.NewPrometheus("gin")
	p.Use(router)

	logger.Info("storage initializing")
	storage, err := model.NewDBConfig(cfg, logger)
	if err != nil {
		return nil, err
	}

	reg := prometheus.NewRegistry()
	metrics := monitoring.NewMetrics(reg)

	enforcer, err := GetEnforcer(storage)
	if err != nil {
		return nil, err
	}

	if saveErr := enforcer.SavePolicy(); saveErr != nil {
		return nil, saveErr
	}

	// Initialize session
	logger.Info("session initializing")
	initSession(ctx, router, cfg, logger)

	healthMetrics := monitoring.NewHealthMetrics()
	go monitoring.PingStorage(context.Background(), storage.Operations, healthMetrics)

	return &WebApp{
		cfg:           cfg,
		logger:        logger,
		router:        router,
		storage:       storage,
		healthMetrics: healthMetrics,
		metrics:       metrics,
		enforcer:      enforcer,
	}, nil
}

func (a *WebApp) Run(ctx context.Context) {
	a.startHTTP(ctx)
}

func (a *WebApp) startHTTP(ctx context.Context) {
	a.logger.Info("start HTTP")
	a.setupStaticFiles()

	// Routes
	a.setupSystemRoutes()
	// Group: v1
	a.setupAppRoutesV1()

	a.logger.Info("view Casbin Policies:")
	policies := a.enforcer.GetPolicy()
	for _, p := range policies {
		a.logger.Infof("Policy: %v", p)
	}
	a.logger.Info("application completely initialized, ...started")
	a.logger.Infof("The service is ready to listen and serve on %s:%s.", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	// Start server

	server := a.startServer()
	// Wait for context cancellation to stop the server gracefully
	<-ctx.Done()

	// Shutdown the server gracefully
	a.shutdownServer(server)

	a.logger.Info("Application stopped")
}

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
			tracks.DELETE("/deleteAll", a.GetDeleteAll)
			tracks.DELETE("/delete/:code", a.GetDeleteByID)
			tracks.POST("/add", a.PostTracks)
			tracks.PATCH("/update", a.UpdateTrack)
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
			audio.POST("/upload", a.PostFiles)
		}
		playlist := v1.Group("/playlist")
		{
			playlist.POST("/:playlist_id/add/track/:track_id", a.AddToPlaylist)
			playlist.DELETE("/:playlist_id/remove/track/:track_id", a.RemoveFromPlaylist)
			playlist.DELETE("/:playlist_id/clear", a.ClearPlaylist)
			playlist.POST("/create", a.CreatePlaylist)
			playlist.DELETE("/delete/:id", a.DeletePlaylist)
			playlist.POST("/:playlist_id/set", a.SetFromPlaylist)
		}

		player := v1.Group("/player")
		{
			player.GET("/play/:playlist_id", a.Play)
		}
	}
}

func (a *WebApp) setupStaticFiles() {
	a.router.StaticFile("/favicon.ico", "./favicon.ico")
}

func (a *WebApp) setupSystemRoutes() {
	a.logger.Info("heartbeat metric initializing")
	a.router.GET("/health", func(c *gin.Context) {
		monitoring.HealthGET(c, a.healthMetrics) // Pass the healthMetrics to HealthGET.
	})
	a.router.GET("/ping", Ping)
	a.router.GET("/job/status", JobStatus)

	a.router.Use(ExtractUserRole(a.logger))
	a.router.Use(NewAuthorizerWithRoleExtractor(a.enforcer, a.logger, func(c *gin.Context) string {
		if role, ok := c.Get("userRole"); ok {
			return role.(string)
		}
		return "anonymous" // Default role
	}))
}

func (a *WebApp) startServer() *http.Server {
	connectionString := fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
	server := &http.Server{
		Addr:              connectionString,
		Handler:           a.router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error:", err)
		}
	}()

	return server
}

func (a *WebApp) shutdownServer(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.logger.Fatal("HTTP server shutdown error:", err)
	}
}

func LogWithLogrusf(logger *logging.Logger, format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
