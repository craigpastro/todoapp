package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/sqlc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const data = "some data"

var q *sqlc.Queries

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

	pool := MustNew(connString, true)
	defer pool.Close()

	q = sqlc.New(pool)

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	created, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   data,
	})
	require.NoError(t, err)

	read, err := q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		PostID: created.PostID,
	})
	require.NoError(t, err)

	require.Equal(t, created.UserID, read.UserID)
	require.Equal(t, created.PostID, read.PostID)
	require.Equal(t, created.Data, read.Data)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	_, err := q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		PostID: uuid.NewString(),
	})
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestReadAll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	post1, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   "data1",
	})
	require.NoError(t, err)

	post2, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   "data2",
	})
	require.NoError(t, err)

	posts, err := q.ReadPage(ctx, sqlc.ReadPageParams{
		UserID: userID,
	})
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
	userID := uuid.NewString()

	post, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   data,
	})
	require.NoError(t, err)

	time.Sleep(time.Millisecond) // just in case

	newData := "new data"
	updatedPost, err := q.Upsert(ctx, sqlc.UpsertParams{
		UserID: userID,
		PostID: post.PostID,
		Data:   newData,
	})
	require.NoError(t, err)

	require.Equal(t, updatedPost.Data, newData, "got '%s', want '%s'")
	require.True(t, post.CreatedAt.Time.Before(updatedPost.UpdatedAt.Time))
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	post, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   data,
	})
	require.NoError(t, err)

	err = q.Delete(ctx, sqlc.DeleteParams{
		UserID: userID,
		PostID: post.PostID,
	})
	require.NoError(t, err)

	// Now try to read the deleted post; it should not exist.
	_, err = q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		PostID: post.PostID,
	})
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()

	err := q.Delete(ctx, sqlc.DeleteParams{
		UserID: "foo",
		PostID: uuid.NewString(),
	})
	require.NoError(t, err)
}
