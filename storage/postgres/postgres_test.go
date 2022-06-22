package postgres

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/stretchr/testify/require"
)

const data = "some data"

var (
	db storage.Storage
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	pool, err := CreatePool(ctx, Config{URI: "postgres://postgres:password@127.0.0.1:5432/postgres"})
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS post (
		user_id TEXT NOT NULL,
		post_id TEXT NOT NULL,
		data TEXT,
		created_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ,
		PRIMARY KEY (user_id, post_id)
	)`); err != nil {
		log.Fatalf("error creating table: %v\n", err)
	}

	db = New(pool, telemetry.NewNoopTracer())

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.UserID, created.UserID, "got '%s', want '%s'", record.UserID, userID)
	require.Equal(t, record.PostID, created.PostID, "got '%s', want '%s'", record.PostID, created.PostID)
	require.Equal(t, record.Data, data, "got '%s', want '%s'", record.Data, data)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Read(ctx, userID, "1")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestReadAll(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	_, err := db.Create(ctx, userID, "data 1")
	require.NoError(t, err)
	_, err = db.Create(ctx, userID, "data 2")
	require.NoError(t, err)

	records, err := db.ReadAll(ctx, userID)
	require.NoError(t, err)

	require.Len(t, records, 2, "got '%d', want '%d'", len(records), 2)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	newData := "new data"
	_, err = db.Update(ctx, userID, created.PostID, newData)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.Data, newData, "got '%s', want '%s'")
	require.True(t, record.CreatedAt.Before(record.UpdatedAt))
}

func TestUpdateNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Update(ctx, userID, "1", "new data")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)

	err := db.Delete(ctx, userID, created.PostID)
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, created.PostID)
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
