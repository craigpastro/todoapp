package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var _ storage.Storage = (*Postgres)(nil)

type Postgres struct {
	db     *sql.DB
	tracer trace.Tracer
}

func New(db *sql.DB, tracer trace.Tracer) *Postgres {
	return &Postgres{
		db:     db,
		tracer: tracer,
	}
}

func CreateDB(ctx context.Context, connString string, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening Postgres: %w", err)
	}

	err = backoff.Retry(func() error {
		if err = db.Ping(); err != nil {
			logger.Info("waiting for Postgres")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres: %w", err)
	}

	return db, nil
}

func (p *Postgres) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.Create")
	defer span.End()

	postID := ulid.Make().String()
	now := time.Now()
	if _, err := p.db.ExecContext(ctx, "INSERT INTO post VALUES ($1, $2, $3, $4, $5)", userID, postID, data, now, now); err != nil {
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

	row := p.db.QueryRowContext(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1 AND post_id = $2", userID, postID)
	var record storage.Record
	if err := row.Scan(&record.UserID, &record.PostID, &record.Data, &record.CreatedAt, &record.UpdatedAt); errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrPostDoesNotExist
	} else if err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return &record, nil
}

func (p *Postgres) ReadAll(ctx context.Context, userID string) (storage.RecordIterator, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.ReadAll")
	defer span.End()

	rows, err := p.db.QueryContext(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	return &recordInterator{rows: rows}, nil
}

type recordInterator struct {
	rows *sql.Rows
}

func (i *recordInterator) Next(_ context.Context) bool {
	return i.rows.Next()
}

func (i *recordInterator) Get(dest *storage.Record) error {
	return i.rows.Scan(&dest.UserID, &dest.PostID, &dest.Data, &dest.CreatedAt, &dest.UpdatedAt)
}

func (i *recordInterator) Close(_ context.Context) {
	i.rows.Close()
}

func (p *Postgres) Upsert(ctx context.Context, userID, postID, data string) (*storage.Record, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.Upsert")
	defer span.End()

	record, err := p.Read(ctx, userID, postID)
	if err != nil {
		if errors.Is(err, storage.ErrPostDoesNotExist) {
			return nil, err
		}
		return nil, fmt.Errorf("error updating: %w", err)
	}

	now := time.Now()
	stmt := "INSERT INTO post VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, post_id) DO UPDATE SET data = $3, updated_at = $5"
	if _, err := p.db.ExecContext(ctx, stmt, userID, postID, data, now, now); err != nil {
		return nil, fmt.Errorf("error updating: %w", err)
	}

	return &storage.Record{
		UserID:    userID,
		PostID:    postID,
		Data:      data,
		CreatedAt: record.CreatedAt,
		UpdatedAt: now,
	}, nil
}

func (p *Postgres) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := p.tracer.Start(ctx, "postgres.Delete")
	defer span.End()

	if _, err := p.db.ExecContext(ctx, "DELETE FROM post WHERE user_id = $1 AND post_id = $2", userID, postID); err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
