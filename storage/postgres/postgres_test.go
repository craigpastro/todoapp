package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/kelseyhightower/envconfig"
)

const data = "some data"

var (
	ctx context.Context
	db  storage.Storage
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
	tracer, _ := instrumentation.NewTracer(ctx, instrumentation.TracerConfig{Enabled: false})

	var err error
	db, err = New(ctx, tracer, config.PostgresURI)
	if err != nil {
		fmt.Println("error initializing Postgres")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	record, err := db.Read(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	if record.UserID != userID {
		t.Errorf("wrong userID: got '%s', but wanted '%s'", record.UserID, userID)
	}

	if record.PostID != postID {
		t.Errorf("wrong postID: got '%s', but wanted '%s'", record.PostID, postID)
	}

	if record.Data != data {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, data)
	}
}

func TestReadNotExists(t *testing.T) {
	userID := myid.New()
	if _, err := db.Read(ctx, userID, "1"); err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestReadAll(t *testing.T) {
	userID := myid.New()
	db.Create(ctx, userID, "data 1")
	db.Create(ctx, userID, "data 2")
	records, err := db.ReadAll(ctx, userID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("wrong number of records: got '%d', but wanted '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	newData := "new data"
	db.Update(ctx, userID, postID, newData)
	record, _ := db.Read(ctx, userID, postID)

	if record.Data != newData {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, newData)
	}

	if record.CreatedAt.After(record.UpdatedAt) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestUpdateNotExists(t *testing.T) {
	userID := myid.New()
	if _, err := db.Update(ctx, userID, "1", "new data"); err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestDelete(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)

	if err := db.Delete(ctx, userID, postID); err != nil {
		t.Errorf("error not nil: %v", err)
	}

	// Now try to read the deleted record; it should not exist.
	if _, err := db.Read(ctx, userID, postID); !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestDeleteNotExists(t *testing.T) {
	userID := myid.New()
	postID := myid.New()

	if err := db.Delete(ctx, userID, postID); err != nil {
		t.Errorf("error not nil: %v", err)
	}
}
