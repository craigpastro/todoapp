package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/craigpastro/todoapp/internal/gen/sqlc"
	pb "github.com/craigpastro/todoapp/internal/gen/todoapp/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const aTodo = "buy veggies"

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

	pool := MustNew(&Config{
		ConnString:        fmt.Sprintf("postgres://authenticator:password@%s:%s/postgres", host, port.Port()),
		Migrate:           true,
		MigrateConnString: fmt.Sprintf("postgres://postgres:password@%s:%s/postgres", host, port.Port()),
	})
	defer pool.Close()

	q = sqlc.New(pool)

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	created, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Todo:   aTodo,
	})
	require.NoError(t, err)

	read, err := q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		TodoID: created.TodoID,
	})
	require.NoError(t, err)

	require.Equal(t, created.UserID, read.UserID)
	require.Equal(t, created.TodoID, read.TodoID)
	require.Equal(t, created.Todo, read.Todo)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	_, err := q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		TodoID: uuid.NewString(),
	})
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestReadAll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	post1, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Todo:   "data1",
	})
	require.NoError(t, err)

	post2, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Todo:   "data2",
	})
	require.NoError(t, err)

	posts, err := q.ReadPage(ctx, sqlc.ReadPageParams{
		UserID: userID,
	})
	require.NoError(t, err)

	require.Len(t, posts, 2)

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(pb.Todo{}),
		cmpopts.IgnoreFields(pb.Todo{}, "CreatedAt", "UpdatedAt"),
	}

	require.True(t, cmp.Equal(post1, posts[0], opts...))
	require.True(t, cmp.Equal(post2, posts[1], opts...))
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	t.Run("updateFailsWhenIDDoesNotExist", func(t *testing.T) {
		_, err := q.Update(ctx, sqlc.UpdateParams{
			UserID: userID,
			TodoID: uuid.NewString(),
			Todo:   aTodo,
		})
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("updateSucceedsWhenIDExists", func(t *testing.T) {
		post, err := q.Create(ctx, sqlc.CreateParams{
			UserID: userID,
			Todo:   aTodo,
		})
		require.NoError(t, err)

		time.Sleep(time.Millisecond) // just in case

		newTodo := "get some sleep"
		updatedTodo, err := q.Update(ctx, sqlc.UpdateParams{
			UserID: userID,
			TodoID: post.TodoID,
			Todo:   newTodo,
		})
		require.NoError(t, err)

		require.Equal(t, updatedTodo.Todo, newTodo, "got '%s', want '%s'")
		require.True(t, post.CreatedAt.Time.Before(updatedTodo.UpdatedAt.Time))
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	todo, err := q.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Todo:   aTodo,
	})
	require.NoError(t, err)

	err = q.Delete(ctx, sqlc.DeleteParams{
		UserID: userID,
		TodoID: todo.TodoID,
	})
	require.NoError(t, err)

	// Now try to read the deleted post; it should not exist.
	_, err = q.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		TodoID: todo.TodoID,
	})
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()

	err := q.Delete(ctx, sqlc.DeleteParams{
		UserID: "foo",
		TodoID: uuid.NewString(),
	})
	require.NoError(t, err)
}
