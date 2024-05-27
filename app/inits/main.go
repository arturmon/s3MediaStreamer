package inits

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	repoDB "s3MediaStreamer/app/repository/postgres"
	repoS3 "s3MediaStreamer/app/repository/s3"
	"s3MediaStreamer/app/services/acl"
	"s3MediaStreamer/app/services/audio"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/cashing"
	"s3MediaStreamer/app/services/consul"
	"s3MediaStreamer/app/services/db"
	"s3MediaStreamer/app/services/health"
	"s3MediaStreamer/app/services/monitoring"
	"s3MediaStreamer/app/services/otel"
	"s3MediaStreamer/app/services/otp"
	"s3MediaStreamer/app/services/playlist"
	"s3MediaStreamer/app/services/rabbitmq"
	"s3MediaStreamer/app/services/s3"
	session "s3MediaStreamer/app/services/session"
	"s3MediaStreamer/app/services/tags"
	"s3MediaStreamer/app/services/track"
	"s3MediaStreamer/app/services/user"

	"github.com/gin-contrib/sessions"
	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"github.com/streadway/amqp"
)

type initConnect struct {
	cashingDB    *redis.Client
	RabbitCon    *amqp.Connection
	s3Client     *minio.Client
	pgClient     *repoDB.Client
	SessionStore sessions.Store
}

type initRepo struct {
	InitConnect *initConnect
	CashingRepo cashing.CachingRepository
	S3Repo      *repoS3.S3Repository
	PgRepo      *repoDB.Client
}

// TODO metrics,enforcer, audio, track, acl, auth, db, health, rabbit, s3, session, tags, user
// TODO refactor , otel
// TODO monitoring
type Service struct {
	InitRepo        *initRepo
	ConsulService   *consul.ConsulService
	ConsulElection  *consul.Election
	AuthCache       *cashing.CachingService
	TagService      *tags.TagsService
	S3Storage       *s3.S3Service
	TracingProvider *otel.Provider
	MetricsMonitor  *monitoring.Metrics
	AccessControl   *auth.AuthService
	Audio           *audio.AudioService
	Track           *track.TrackService
	Acl             *acl.AclService
	Storage         *db.DBService
	Health          *health.HealthCheckService
	Message         *rabbitmq.MessageService
	Tags            *tags.TagsService
	User            *user.UserService
	Playlist        *playlist.PlaylistService
	Session         *session.SessionService
	OTP             *otp.OTPService
}

func InitServices(ctx context.Context, appName, version string, cfg *model.Config, logger *logs.Logger) (*Service, error) {

	connectSetup, err := initConnects(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	repoSetup := initRepos(cfg, logger, connectSetup)

	initService, err := initServices(ctx, appName, version, cfg, logger, repoSetup)
	if err != nil {
		return nil, err
	}
	return initService, nil
}
