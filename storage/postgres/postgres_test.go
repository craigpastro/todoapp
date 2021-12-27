package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

const data = "some data"

var (
	ctx      context.Context
	postgres *Postgres
)

type Config struct {
	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`
}

func TestMain(m *testing.M) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fmt.Println("error reading config", err)
		os.Exit(1)
	}

	ctx = context.Background()
	pool, err := pgxpool.Connect(ctx, config.PostgresURI)
	if err != nil {
		fmt.Println("error initializing postgres")
		os.Exit(1)
	}

	pool.Exec(ctx, `DROP TABLE IF EXISTS post`)
	pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS post (
		user_id TEXT NOT NULL,
		post_id TEXT NOT NULL,
		data TEXT,
		created_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ,
		PRIMARY KEY (user_id, post_id)
	)`)

	postgres = &Postgres{pool: pool}

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	userID := myid.New()
	postID, _, _ := postgres.Create(ctx, userID, data)
	record, err := postgres.Read(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %s", err)
	}

	if record.UserID != userID {
		t.Errorf("wrong userID. got '%s', want '%s'", record.UserID, userID)
	}

	if record.PostID != postID {
		t.Errorf("wrong postID. got '%s', want '%s'", record.PostID, postID)
	}

	if record.Data != data {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, data)
	}
}

func TestReadAll(t *testing.T) {
	userID := myid.New()
	postgres.Create(ctx, userID, "data 1")
	postgres.Create(ctx, userID, "data 2")
	records, err := postgres.ReadAll(ctx, userID)

	if err != nil {
		t.Errorf("error not nil: %s", err)
	}

	if len(records) != 2 {
		t.Errorf("wrong number of records. got '%d', want '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	userID := myid.New()
	postID, _, _ := postgres.Create(ctx, userID, data)
	newData := "new data"
	postgres.Update(ctx, userID, postID, newData)
	record, _ := postgres.Read(ctx, userID, postID)

	if record.Data != newData {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, newData)
	}

	if record.CreatedAt.After(record.UpdatedAt) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestDelete(t *testing.T) {
	userID := myid.New()
	postID, _, _ := postgres.Create(ctx, userID, data)
	err := postgres.Delete(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %s", err)
	}

	// Now try to read the delete record; it should not exist.
	_, err = postgres.Read(ctx, userID, postID)
	if !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error. got '%v', want '%v'", err, storage.ErrPostDoesNotExist)
	}
}
