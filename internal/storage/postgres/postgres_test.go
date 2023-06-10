package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/internal/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const data = "some data"

var db storage.Storage

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "password"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = container.Terminate(context.Background())
	}()

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	connString := fmt.Sprintf("postgres://postgres:password@%s:%s/postgres", host, port.Port())

	p := MustNew(connString, true)
	defer p.Close()

	db = p

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.UserID, created.UserID)
	require.Equal(t, record.PostID, created.PostID)
	require.Equal(t, record.Data, data)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()

	_, err := db.Read(ctx, userID, "1")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestReadAll(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()

	rec1, err := db.Create(ctx, userID, "data 1")
	require.NoError(t, err)

	rec2, err := db.Create(ctx, userID, "data 2")
	require.NoError(t, err)

	iter, err := db.ReadAll(ctx, userID)
	require.NoError(t, err)

	var record storage.Record

	require.True(t, iter.Next(ctx))
	require.NoError(t, iter.Get(&record))
	// Monotonic clock issues: see https://github.com/stretchr/testify/issues/502
	require.True(t, cmp.Equal(rec1, &record, cmpopts.IgnoreFields(storage.Record{}, "CreatedAt", "UpdatedAt")))

	require.True(t, iter.Next(ctx))
	require.NoError(t, iter.Get(&record))
	// Monotonic clock issues: see https://github.com/stretchr/testify/issues/502
	require.True(t, cmp.Equal(rec2, &record, cmpopts.IgnoreFields(storage.Record{}, "CreatedAt", "UpdatedAt")))

	require.False(t, iter.Next(ctx))
}

func TestUpsert(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	record, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	time.Sleep(time.Millisecond) // just in case
	newData := "new data"
	updatedRecord, err := db.Upsert(ctx, userID, record.PostID, newData)
	require.NoError(t, err)

	require.Equal(t, updatedRecord.Data, newData, "got '%s', want '%s'")
	require.True(t, record.CreatedAt.Before(updatedRecord.UpdatedAt))
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	created, _ := db.Create(ctx, userID, data)

	err := db.Delete(ctx, userID, created.PostID)
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, created.PostID)
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	postID := ulid.Make().String()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
