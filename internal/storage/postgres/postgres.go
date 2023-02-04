package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/internal/gen/db"
	"github.com/craigpastro/crudapp/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("internal/storage/postgres")

type Postgres struct {
	queries *db.Queries
	db      *sql.DB // still needed for streaming read all
}

var _ storage.Storage = (*Postgres)(nil)

func New(sqlDB *sql.DB) *Postgres {
	return &Postgres{
		queries: db.New(sqlDB),
		db:      sqlDB,
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
	ctx, span := tracer.Start(ctx, "postgres.Create")
	defer span.End()

	post, err := p.queries.Create(ctx, db.CreateParams{
		UserID: userID,
		PostID: ulid.Make().String(),
		Data:   data,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating: %w", err)
	}

	return &storage.Record{
		UserID:    post.UserID,
		PostID:    post.PostID,
		Data:      post.Data,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (p *Postgres) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := tracer.Start(ctx, "postgres.Read")
	defer span.End()

	post, err := p.queries.Read(ctx, db.ReadParams{
		UserID: userID,
		PostID: postID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrPostDoesNotExist
		}
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return &storage.Record{
		UserID:    post.UserID,
		PostID:    post.PostID,
		Data:      post.Data,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (p *Postgres) ReadAll(ctx context.Context, userID string) (storage.RecordIterator, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadAll")
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
	ctx, span := tracer.Start(ctx, "postgres.Upsert")
	defer span.End()

	post, err := p.queries.Upsert(ctx, db.UpsertParams{
		UserID: userID,
		PostID: postID,
		Data:   data,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating: %w", err)
	}

	return &storage.Record{
		UserID:    post.UserID,
		PostID:    post.PostID,
		Data:      post.Data,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (p *Postgres) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := tracer.Start(ctx, "postgres.Delete")
	defer span.End()

	err := p.queries.Delete(ctx, db.DeleteParams{
		UserID: userID,
		PostID: postID,
	})
	if err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
