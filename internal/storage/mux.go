package storage

import (
	"context"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/storage/pool"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBMux struct {
	DBWriter *pgxpool.Pool
	DBReader *pgxpool.Pool
}

type IDBMux interface {
	PingWriter(ctx context.Context) error
	PingReader(context.Context) error
	Close()
	BeginWrite() (DB, error)
	BeginRead() (DB, error)
}

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DBTx interface {
	DB

	Rollback(ctx context.Context) error
}

func NewDBMux(ctx context.Context, config internal.Config) (*DBMux, error) {
	pools, err := pool.NewPool(ctx, pool.PoolConfig{
		DBName:         config.DB.DBName,
		Username:       config.DB.Username,
		Password:       config.DB.Password,
		SSLMode:        config.DB.SSLMode,
		WriterEndpoint: config.DB.Writer,
		ReaderEndpoint: config.DB.Reader,
	})
	if err != nil {
		return nil, err
	}
	return &DBMux{
		DBWriter: pools.Writer(),
		DBReader: pools.Reader(),
	}, nil
}

func (m *DBMux) PingWriter(ctx context.Context) error {
	return m.DBWriter.Ping(ctx)
}

func (m *DBMux) PingReader(ctx context.Context) error {
	return m.DBReader.Ping(ctx)
}

func (m *DBMux) Close() {
	m.DBWriter.Close()
	m.DBReader.Close()
}

func (m *DBMux) BeginWrite() (DB, error) {
	return m.DBWriter, nil
}

func (m *DBMux) BeginRead() (DB, error) {
	return m.DBReader, nil
}
