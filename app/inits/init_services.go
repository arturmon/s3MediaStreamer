package inits

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/acl"
	"s3MediaStreamer/app/services/audio"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/cashing"
	"s3MediaStreamer/app/services/consul"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/health"
	"s3MediaStreamer/app/services/mdns"
	"s3MediaStreamer/app/services/monitoring"
	"s3MediaStreamer/app/services/otel"
	"s3MediaStreamer/app/services/otp"
	"s3MediaStreamer/app/services/playlist"
	"s3MediaStreamer/app/services/rabbitmq"
	"s3MediaStreamer/app/services/s3"
	session "s3MediaStreamer/app/services/session"
	"s3MediaStreamer/app/services/tags"
	"s3MediaStreamer/app/services/track"
	"s3MediaStreamer/app/services/tree"
	"s3MediaStreamer/app/services/user"

	"github.com/prometheus/client_golang/prometheus"
)

func initServices(ctx context.Context,
	appName,
	version string,
	cfg *model.Config,
	logger *logs.Logger,
	repo *initRepo) (*Service, error) {
	logger.Info("Starting initialize the service...")

	consulService := consul.NewService(appName, cfg, logger)
	consulService.Start()
	leaderElectionService := consul.NewElection(appName, logger, *consulService)
	consulKV := consul.NewConsulKVService(cfg, logger, *consulService)

	cashingService := cashing.NewCachingService(repo.CashingRepo)
	tagsService := tags.NewTagsService()
	s3Service := s3.NewS3Service(repo.S3Repo, repo.PgRepo)
	tracingService, err := otel.InitializeTracer(ctx, cfg, logger, appName, version)
	if err != nil {
		return nil, err
	}
	registry := prometheus.NewRegistry()
	metrics := monitoring.NewMetrics(registry)
	// metricsMonitorService := monitoring.NewMonitoringService()

	accessControlService := auth.NewAuthService(repo.PgRepo)
	treeService := tree.NewTreeService()
	trackService := track.NewTrackService(repo.PgRepo, treeService)
	aclService, err := acl.NewACLService()
	if err != nil {
		return nil, err
	}
	storageService := db.NewDBService(repo.PgRepo)

	healthMetrics := health.NewHealthMetrics()
	healthService := health.NewHealthCheckWrapper(healthMetrics, repo.PgRepo, repo.InitConnect.RabbitCon, repo.S3Repo, logger)
	healthService.StartHealthChecks()

	sessionService := session.NewSessionHandler()
	messageService := rabbitmq.NewMessageService(logger, repo.PgRepo, *s3Service, *trackService, *tagsService)

	userService := user.NewUserService(repo.PgRepo, *sessionService, *cashingService, logger, *accessControlService, cfg)
	playlistService := playlist.NewPlaylistService(repo.PgRepo, repo.PgRepo, *sessionService, *accessControlService, *userService, logger, treeService)
	audioService := audio.NewAudioService(*trackService, *s3Service, *playlistService, logger)
	otpService := otp.NewOTPService(*userService, cfg)
	// Initialize mDNS service if enabled
	var mDNSService *mdns.Service
	if cfg.AppConfig.MDNS.Enabled {
		mDNSService = mdns.NewMDNSService(appName, cfg.Listen.Port, logger)
		mDNSService.Start()
	}
	logger.Info("Complete service initialize.")
	return &Service{
		InitRepo:        repo,
		ConsulService:   consulService,
		ConsulElection:  leaderElectionService,
		ConsulKV:        consulKV,
		AuthCache:       cashingService,
		S3Storage:       s3Service,
		TracingProvider: tracingService,
		MetricsMonitor:  metrics,
		AccessControl:   accessControlService,
		Audio:           audioService,
		Track:           trackService,
		ACL:             aclService,
		Storage:         storageService,
		Health:          healthService,
		Message:         messageService,
		Tags:            tagsService,
		User:            userService,
		Playlist:        playlistService,
		Session:         sessionService,
		OTP:             otpService,
		Tree:            treeService,
		mDNS:            mDNSService,
	}, nil
}
