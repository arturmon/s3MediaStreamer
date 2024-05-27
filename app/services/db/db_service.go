package db

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type DBRepository interface {
	ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error)
	Connect(_ *logs.Logger) error
	Ping(ctx context.Context) error
	Close(_ context.Context) error
	ExecInTransaction(ctx context.Context, sql string, args ...interface{}) error
	QueryInTransaction(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	GetConnectionString() string
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DBService struct {
	dbRepository DBRepository
}

func NewDBService(dbRepository DBRepository) *DBService {
	return &DBService{dbRepository: dbRepository}
}

func (s *DBService) ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error) {
	return s.dbRepository.ExecuteSelectQuery(ctx, selectBuilder)
}

func (s *DBService) Connect(logs *logs.Logger) error {
	return s.dbRepository.Connect(logs)
}

func (s *DBService) Ping(ctx context.Context) error {
	return s.dbRepository.Ping(ctx)
}

func (s *DBService) Close(ctx context.Context) error {
	return s.dbRepository.Close(ctx)
}

func (s *DBService) ExecInTransaction(ctx context.Context, sql string, args ...interface{}) error {
	return s.dbRepository.ExecInTransaction(ctx, sql, args)
}

func (s *DBService) QueryInTransaction(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return s.dbRepository.QueryInTransaction(ctx, sql, args)
}

func (s *DBService) GetConnectionString() string {
	return s.dbRepository.GetConnectionString()
}
func (s *DBService) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.dbRepository.Begin(ctx)
}
