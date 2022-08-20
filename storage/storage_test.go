package storage_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/mongodb"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const data = "some data"

type storageTest struct {
	name      string
	storage   storage.Storage
	container testcontainers.Container
}

func TestStorage(t *testing.T) {
	storageTests := []storageTest{
		newMemory(),
		newMongoDB(t),
		newPostgres(t),
	}

	for _, test := range storageTests {
		t.Run(test.name, func(t *testing.T) {
			testRead(t, test.storage)
			testReadNotExists(t, test.storage)
			testReadAll(t, test.storage)
			testUpdate(t, test.storage)
			testUpdateNotExists(t, test.storage)
			testDelete(t, test.storage)
			testDeleteNotExists(t, test.storage)

			if test.container != nil {
				if err := test.container.Terminate(context.Background()); err != nil {
					log.Println(err)
				}
			}
		})
	}
}

func newMemory() storageTest {
	return storageTest{
		name:    "memory",
		storage: memory.New(telemetry.NewNoopTracer()),
	}
}

func newMongoDB(t *testing.T) storageTest {
	ctx := context.Background()
	logger := telemetry.Must(telemetry.NewLogger(telemetry.LoggerConfig{}))
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env:          map[string]string{"MONGO_INITDB_ROOT_USERNAME": "mongodb", "MONGO_INITDB_ROOT_PASSWORD": "password"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017/tcp")
	require.NoError(t, err)

	coll, err := mongodb.CreateCollection(ctx, mongodb.Config{URL: fmt.Sprintf("mongodb://mongodb:password@localhost:%s", port.Port())}, logger)
	require.NoError(t, err)

	return storageTest{
		name:      "mongodb",
		storage:   mongodb.New(coll, tracer),
		container: container,
	}
}

func newPostgres(t *testing.T) storageTest {
	ctx := context.Background()
	logger := telemetry.Must(telemetry.NewLogger(telemetry.LoggerConfig{}))
	tracer := telemetry.NewNoopTracer()

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
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	pool, err := postgres.CreatePool(ctx, postgres.Config{URL: fmt.Sprintf("postgres://postgres:password@localhost:%s/postgres", port.Port())}, logger)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS post (
		user_id TEXT NOT NULL,
		post_id TEXT NOT NULL,
		data TEXT,
		created_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ,
		PRIMARY KEY (user_id, post_id)
	)`)
	require.NoError(t, err)

	return storageTest{
		name:      "postgres",
		storage:   postgres.New(pool, tracer),
		container: container,
	}
}

func testRead(t *testing.T, storage storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, err := storage.Create(ctx, userID, data)
	require.NoError(t, err)
	record, err := storage.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.UserID, created.UserID)
	require.Equal(t, record.PostID, created.PostID)
	require.Equal(t, record.Data, data)
}

func testReadNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Read(ctx, userID, "1")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testReadAll(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()

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

func testUpdate(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	time.Sleep(time.Millisecond) // just in case
	newData := "new data"
	_, err = db.Update(ctx, userID, created.PostID, newData)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.Data, newData, "got '%s', want '%s'")
	require.True(t, record.CreatedAt.Before(record.UpdatedAt))
}

func testUpdateNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Update(ctx, userID, "1", "new data")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testDelete(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)

	err := db.Delete(ctx, userID, created.PostID)
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, created.PostID)
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testDeleteNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
