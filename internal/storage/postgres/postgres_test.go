package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
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

	read, err := db.Read(ctx, userID, created.GetPostId())
	require.NoError(t, err)

	require.Equal(t, created.GetUserId(), read.GetUserId())
	require.Equal(t, created.GetPostId(), read.GetPostId())
	require.Equal(t, created.GetData(), read.GetData())
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

	post1, err := db.Create(ctx, userID, "data 1")
	require.NoError(t, err)

	post2, err := db.Create(ctx, userID, "data 2")
	require.NoError(t, err)

	posts, _, err := db.ReadAll(ctx, userID)
	require.NoError(t, err)

	require.Len(t, posts, 2)

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(pb.Post{}),
		cmpopts.IgnoreFields(pb.Post{}, "CreatedAt", "UpdatedAt"),
	}

	require.True(t, cmp.Equal(post1, posts[0], opts...))
	require.True(t, cmp.Equal(post2, posts[1], opts...))
}

func TestUpsert(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	post, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	time.Sleep(time.Millisecond) // just in case

	newData := "new data"
	updatedPost, err := db.Upsert(ctx, userID, post.GetPostId(), newData)
	require.NoError(t, err)

	require.Equal(t, updatedPost.Data, newData, "got '%s', want '%s'")
	require.True(t, post.GetCreatedAt().AsTime().Before(updatedPost.GetUpdatedAt().AsTime()))
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()

	post, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	err = db.Delete(ctx, userID, post.GetPostId())
	require.NoError(t, err)

	// Now try to read the deleted post; it should not exist.
	_, err = db.Read(ctx, userID, post.GetPostId())
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	userID := ulid.Make().String()
	postID := ulid.Make().String()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
