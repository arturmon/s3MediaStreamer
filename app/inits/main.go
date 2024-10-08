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
	S3Repo      *repoS3.Repository
	PgRepo      *repoDB.Client
}

type Service struct {
	InitRepo        *initRepo
	ConsulService   *consul.Service
	ConsulElection  *consul.ElService
	ConsulKV        *consul.KVService
	AuthCache       *cashing.CachingService
	S3Storage       *s3.Service
	TracingProvider *otel.Provider
	MetricsMonitor  *monitoring.Metrics
	AccessControl   *auth.Service
	Audio           *audio.Service
	Track           *track.Service
	ACL             *acl.Service
	Storage         *db.Service
	Health          *health.Service
	Message         *rabbitmq.Service
	Tags            *tags.Service
	User            *user.Service
	Playlist        *playlist.Service
	Session         *session.Service
	OTP             *otp.Service
	Tree            *tree.Service
	mDNS            *mdns.Service
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
