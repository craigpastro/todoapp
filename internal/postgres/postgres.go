package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/exp/slog"
)

//go:embed migrations/*
var fs embed.FS

func New(connString string, migrate bool) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connString)
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

	if migrate {
		if err := Migrate(connString); err != nil {
			return nil, err
		}
	}

	return pool, nil
}

func MustNew(connString string, migrate bool) *pgxpool.Pool {
	pool, err := New(connString, migrate)
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

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose error: %w", err)
	}

	return nil
}
