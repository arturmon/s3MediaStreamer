package postgres

import (
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Pool             *pgxpool.Pool
	ConnectionString string
}

func InitDBRepository(_ *model.Config, logger *logs.Logger, pgClient *Client) *Client {
	logger.Info("Starting DB repository...")
	return &Client{
		Pool: pgClient.Pool,
	}
}
