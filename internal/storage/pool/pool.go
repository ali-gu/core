package pool

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "golang.org/x/crypto/x509roots/fallback"
)

const maxConnectionLifetime = 1 * time.Hour

//go:embed certs/supabase-root-2021-ca.pem
var supabaseRootCAPEM []byte

var supabaseRootCAPool = func() *x509.CertPool {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(supabaseRootCAPEM) {
		panic("pool: failed to parse embedded Supabase root CA")
	}
	return pool
}()

type PoolConfig struct {
	DBName         string
	Username       string
	Password       string
	SSLMode        internal.SSLMode
	WriterEndpoint internal.DBEndpointConfig
	ReaderEndpoint internal.DBEndpointConfig
}

type Pool struct {
	config PoolConfig

	readerPool *pgxpool.Pool
	writerPool *pgxpool.Pool
}

func (p *Pool) beforeConnect(config *pgx.ConnConfig, endpoint internal.DBEndpointConfig) error {
	config.Host = endpoint.Host
	config.Port = endpoint.Port
	config.Database = p.config.DBName
	config.User = p.config.Username
	config.Password = p.config.Password

	config.Fallbacks = nil
	switch p.config.SSLMode {
	case internal.SSLModeRequired:
		config.TLSConfig = &tls.Config{
			ServerName: endpoint.Host,
			RootCAs:    supabaseRootCAPool,
		}
	default:
		config.TLSConfig = nil
	}
	return nil
}

func (p *Pool) Writer() *pgxpool.Pool {
	return p.writerPool
}

func (p *Pool) Reader() *pgxpool.Pool {
	return p.readerPool
}

func NewPool(ctx context.Context, config PoolConfig) (*Pool, error) {
	pool := &Pool{
		config: config,
	}

	writerConfig, err := buildConfig(ctx, pool, config.WriterEndpoint)
	if err != nil {
		return nil, err
	}
	pool.writerPool, err = pgxpool.NewWithConfig(ctx, writerConfig)
	if err != nil {
		return nil, err
	}

	if config.ReaderEndpoint == config.WriterEndpoint {
		pool.readerPool = pool.writerPool
		return pool, nil
	}

	readerConfig, err := buildConfig(ctx, pool, config.ReaderEndpoint)
	if err != nil {
		return nil, err
	}
	pool.readerPool, err = pgxpool.NewWithConfig(ctx, readerConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func buildConfig(_ context.Context, pool *Pool, endpoint internal.DBEndpointConfig) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}
	config.MaxConnLifetime = maxConnectionLifetime

	config.MaxConnIdleTime = 30 * time.Second
	config.MinConns = 0
	config.MaxConns = 5

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	config.BeforeConnect = func(ctx context.Context, config *pgx.ConnConfig) error {
		return pool.beforeConnect(config, endpoint)
	}

	return config, nil
}
