package inits

import (
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	repoCashing "s3MediaStreamer/app/repository/cashing"
	repoDB "s3MediaStreamer/app/repository/postgres"
	repoS3 "s3MediaStreamer/app/repository/s3"
)

func initRepos(cfg *model.Config, logger *logs.Logger, conn *initConnect) *initRepo {
	logger.Info("Starting initialize the repository...")
	cashingRepo := repoCashing.InitRedisRepository(logger, conn.cashingDB)
	s3Repo := repoS3.NewS3Repository(cfg, logger, conn.s3Client)
	pgRepo := repoDB.InitDBRepository(cfg, logger, conn.pgClient)
	logger.Info("Complete repository initialize.")
	return &initRepo{
		InitConnect: conn,
		CashingRepo: cashingRepo,
		S3Repo:      s3Repo,
		PgRepo:      pgRepo,
	}
}
