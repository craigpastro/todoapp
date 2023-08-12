package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/retrier"
	pb "github.com/craigpastro/todoapp/internal/gen/todoapp/v1"
	"github.com/craigpastro/todoapp/internal/gen/todoapp/v1/todoappv1connect"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtcl9yb2JvdG8ifQ.oUD_0r5Q1H_akjeJFWYAxbcr2fckBEb7M25wVJw432Y"
	port  = 12345
	aTodo = "buy some veggies"
)

var (
	client todoappv1connect.TodoAppServiceClient
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	containerPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	go func() {
		run(ctx, &config{
			Port:                      port,
			JWTSecret:                 "PMBrjiOH5RMo6nQHidA62XctWGxDG0rw",
			PostgresConnString:        fmt.Sprintf("postgres://authenticator:password@%s:%s/postgres", host, containerPort.Port()),
			PostgresAutoMigrate:       true,
			PostgresMigrateConnString: fmt.Sprintf("postgres://postgres:password@%s:%s/postgres", host, containerPort.Port()),
		})
	}()

	client = todoappv1connect.NewTodoAppServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://localhost:%d", port),
	)

	// Until we have a health endpoint
	cfg := retrier.NewExponentialBackoff()
	cfg.Timeout = 3 * time.Second
	err = retrier.Do(func() error {
		_, err := client.ReadAll(ctx, createRequest(&pb.ReadAllRequest{}))
		if err != nil {
			return err
		}
		return nil
	}, cfg)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = container.Terminate(context.Background())

	os.Exit(code)
}

func TestAPI(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		req := createRequest(&pb.CreateRequest{Todo: aTodo})
		res, err := client.Create(context.Background(), req)
		require.NoError(t, err)

		todo := res.Msg

		require.NotEmpty(t, todo.GetTodoId())
		require.NotEmpty(t, todo.GetCreatedAt())
	})

	t.Run("read", func(t *testing.T) {
		createReq := createRequest(&pb.CreateRequest{Todo: aTodo})
		createRes, err := client.Create(context.Background(), createReq)
		require.NoError(t, err)

		readReq := createRequest(&pb.ReadRequest{TodoId: createRes.Msg.GetTodoId()})
		readRes, err := client.Read(context.Background(), readReq)
		require.NoError(t, err)

		todo := readRes.Msg.GetTodo()

		require.Equal(t, todo, aTodo)
	})

	t.Run("read not exist", func(t *testing.T) {
		req := createRequest(&pb.ReadRequest{TodoId: "foo"})
		_, err := client.Read(context.Background(), req)
		require.ErrorContains(t, err, "todo id does not exist")
	})

	t.Run("upsert", func(t *testing.T) {
		ctx := context.Background()

		createReq := createRequest(&pb.CreateRequest{Todo: aTodo})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		createdTodo := createRes.Msg
		newTodo := "call parents"

		upsertReq := createRequest(&pb.UpdateRequest{
			TodoId: createdTodo.GetTodoId(),
			Todo:   newTodo,
		})
		_, err = client.Update(ctx, upsertReq)
		require.NoError(t, err)

		readReq := createRequest(&pb.ReadRequest{
			TodoId: createdTodo.GetTodoId(),
		})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		todo := readRes.Msg.GetTodo()

		require.Equal(t, todo, newTodo)
	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()

		createReq := createRequest(&pb.CreateRequest{Todo: aTodo})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		createdTodo := createRes.Msg

		deleteReq := createRequest(&pb.DeleteRequest{TodoId: createdTodo.GetTodoId()})
		_, err = client.Delete(ctx, deleteReq)
		require.NoError(t, err)

		// Now try to read the deleted record; it should not exist.
		readReq := createRequest(&pb.ReadRequest{TodoId: createdTodo.GetTodoId()})
		_, err = client.Read(ctx, readReq)
		require.ErrorContains(t, err, "todo id does not exist")
	})

	t.Run("delete not exist", func(t *testing.T) {
		req := createRequest(&pb.DeleteRequest{TodoId: "foo"})
		_, err := client.Delete(context.Background(), req)
		require.NoError(t, err)
	})
}

func createRequest[T any](t *T) *connect.Request[T] {
	req := connect.NewRequest(t)
	req.Header().Add("Authentication", fmt.Sprintf("Bearer %s", token))
	return req
}
