package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connectionURI string) (storage.Storage, error) {
	pool, err := pgxpool.Connect(ctx, connectionURI)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Postgres: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Create(ctx context.Context, userID, data string) (string, time.Time, error) {
	postID := myid.New()
	now := time.Now()
	_, err := p.pool.Exec(ctx, "INSERT INTO post VALUES ($1, $2, $3, $4, $5)", userID, postID, data, now, now)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error creating: %w", err)
	}

	return postID, now, nil
}

func (p *Postgres) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	row := p.pool.QueryRow(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1 AND post_id = $2", userID, postID)
	record := &storage.Record{}
	err := row.Scan(&record.UserID, &record.PostID, &record.Data, &record.CreatedAt, &record.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, storage.ErrPostDoesNotExist
	} else if err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return record, nil
}

func (p *Postgres) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
	rows, err := p.pool.Query(ctx, "SELECT user_id, post_id, data, created_at, updated_at FROM post WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	res := []*storage.Record{}
	for rows.Next() {
		record := &storage.Record{}
		rows.Scan(&record.UserID, &record.PostID, &record.Data, &record.CreatedAt, &record.UpdatedAt)
		res = append(res, record)
	}

	return res, nil
}

func (p *Postgres) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	if _, err := p.Read(ctx, userID, postID); err != nil {
		return time.Time{}, fmt.Errorf("error reading: %w", err)
	}

	now := time.Now()
	if _, err := p.pool.Exec(ctx, "UPDATE post SET data = $1, updated_at = $2 WHERE user_id = $3 AND post_id = $4", data, now, userID, postID); err != nil {
		return time.Time{}, fmt.Errorf("error updating: %w", err)
	}

	return now, nil
}

func (p *Postgres) Delete(ctx context.Context, userID, postID string) error {
	if _, err := p.pool.Exec(ctx, "DELETE FROM post WHERE user_id = $1 AND post_id = $2", userID, postID); err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
