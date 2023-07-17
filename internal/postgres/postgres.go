package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/cenkalti/backoff"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

//go:embed migrations/*
var fs embed.FS

var tracer = otel.Tracer("internal/postgres")

type Config struct {
	ConnString        string
	Migrate           bool
	MigrateConnString string
}

type queryTracer struct{}

func (queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = tracer.Start(ctx, data.SQL)
	return ctx
}

func (queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	trace.SpanFromContext(ctx).End()
}

func New(cfg *Config) (*pgxpool.Pool, error) {
	if cfg.Migrate {
		if err := Migrate(cfg.MigrateConnString); err != nil {
			return nil, err
		}
	}

	pgxCfg, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Add tracing to queries
	pgxCfg.ConnConfig.Tracer = queryTracer{}

	pgxCfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		_, err := conn.Exec(ctx, "set role todoapp_user")
		if err != nil {
			panic(err)
		}
		return true
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres: %w", err)
	}

	err = backoff.Retry(func() error {
		if err = pool.Ping(context.Background()); err != nil {
			slog.Info("waiting for Postgres")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres: %w", err)
	}

	return pool, nil
}

func MustNew(cfg *Config) *pgxpool.Pool {
	pool, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return pool
}

func Migrate(connString string) error {
	goose.SetBaseFS(fs)

	db, err := goose.OpenDBWithDriver("pgx", connString)
	if err != nil {
		return fmt.Errorf("goose error: %w", err)
	}

	err = backoff.Retry(func() error {
		if err = db.Ping(); err != nil {
			slog.Info("waiting for Postgres")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return fmt.Errorf("goose error: error connecting to Postgres: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose error: %w", err)
	}

	return nil
}
