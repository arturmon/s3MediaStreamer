package db

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error)
	Connect(_ *logs.Logger) error
	Ping(ctx context.Context) error
	Close(_ context.Context) error
	ExecInTransaction(ctx context.Context, sql string, args []interface{}) error
	QueryInTransaction(ctx context.Context, sql string, args []interface{}) (pgx.Rows, error)
	GetConnectionString() string
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	dbRepository Repository
}

func NewDBService(dbRepository Repository) *Service {
	return &Service{dbRepository: dbRepository}
}

func (s *Service) ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error) {
	return s.dbRepository.ExecuteSelectQuery(ctx, selectBuilder)
}

func (s *Service) Connect(logs *logs.Logger) error {
	return s.dbRepository.Connect(logs)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.dbRepository.Ping(ctx)
}

func (s *Service) Close(ctx context.Context) error {
	return s.dbRepository.Close(ctx)
}

func (s *Service) ExecInTransaction(ctx context.Context, sql string, args ...interface{}) error {
	return s.dbRepository.ExecInTransaction(ctx, sql, args)
}

func (s *Service) QueryInTransaction(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return s.dbRepository.QueryInTransaction(ctx, sql, args)
}

func (s *Service) GetConnectionString() string {
	return s.dbRepository.GetConnectionString()
}
func (s *Service) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.dbRepository.Begin(ctx)
}
