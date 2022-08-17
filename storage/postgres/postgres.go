package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

type Postgres struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

type Config struct {
	URL string
}

func New(pool *pgxpool.Pool, tracer trace.Tracer) *Postgres {
	return &Postgres{
		pool:   pool,
		tracer: tracer,
	}
}

func CreatePool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	err := backoff.Retry(func() error {
		var err error
		pool, err = pgxpool.Connect(ctx, config.URL)
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error initializing Postgres: %w", err)
	}

	return pool, nil
}

func (p *Postgres) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.Create")
	defer span.End()

	postID := myid.New()
	now := time.Now()
	if _, err := p.pool.Exec(ctx, "INSERT INTO post VALUES ($1, $2, $3, $4, $5)", userID, postID, data, now, now); err != nil {
		return nil, fmt.Errorf("error creating: %w", err)
	}

	return &storage.Record{
		UserID:    userID,
		PostID:    postID,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Postgres) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.Read")
	defer span.End()

	row := p.pool.QueryRow(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1 AND post_id = $2", userID, postID)
	var record storage.Record
	if err := row.Scan(&record.UserID, &record.PostID, &record.Data, &record.CreatedAt, &record.UpdatedAt); errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrPostDoesNotExist
	} else if err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return &record, nil
}

func (p *Postgres) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.ReadAll")
	defer span.End()

	rows, err := p.pool.Query(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	var res []*storage.Record
	for rows.Next() {
		var record storage.Record
		if err := rows.Scan(&record.UserID, &record.PostID, &record.Data, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning: %w", err)
		}
		res = append(res, &record)
	}

	return res, nil
}

func (p *Postgres) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.Update")
	defer span.End()

	if _, err := p.Read(ctx, userID, postID); errors.Is(err, storage.ErrPostDoesNotExist) {
		return time.Time{}, err
	} else if err != nil {
		return time.Time{}, fmt.Errorf("error updating: %w", err)
	}

	now := time.Now()
	if _, err := p.pool.Exec(ctx, "UPDATE post SET data = $1, updated_at = $2 WHERE user_id = $3 AND post_id = $4", data, now, userID, postID); err != nil {
		return time.Time{}, fmt.Errorf("error updating: %w", err)
	}

	return now, nil
}

func (p *Postgres) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := p.tracer.Start(ctx, "postgres.Delete")
	defer span.End()

	if _, err := p.pool.Exec(ctx, "DELETE FROM post WHERE user_id = $1 AND post_id = $2", userID, postID); err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
