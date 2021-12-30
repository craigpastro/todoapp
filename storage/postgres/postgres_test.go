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
	ctx context.Context
	db  *Postgres
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
		fmt.Println("error initializing Postgres")
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

	db = &Postgres{pool: pool}

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	record, err := db.Read(ctx, userID, postID)

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

func TestReadNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Read(ctx, userID, "1")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("wanted '%s', but got '%s'", storage.ErrPostDoesNotExist, err)
	}
}

func TestReadAll(t *testing.T) {
	userID := myid.New()
	db.Create(ctx, userID, "data 1")
	db.Create(ctx, userID, "data 2")
	records, err := db.ReadAll(ctx, userID)

	if err != nil {
		t.Errorf("error not nil: %s", err)
	}

	if len(records) != 2 {
		t.Errorf("wrong number of records. got '%d', want '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	newData := "new data"
	db.Update(ctx, userID, postID, newData)
	record, _ := db.Read(ctx, userID, postID)

	if record.Data != newData {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, newData)
	}

	if record.CreatedAt.After(record.UpdatedAt) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestUpdateNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Update(ctx, userID, "1", "new data")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("wanted ErrPostDoesNotExist, but got: %s", err)
	}
}

func TestDelete(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	err := db.Delete(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %s", err)
	}

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, postID)
	if !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error. got '%v', want '%v'", err, storage.ErrPostDoesNotExist)
	}
}
