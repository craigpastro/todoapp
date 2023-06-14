package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/cenkalti/backoff"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtcl9yb2JvdG8ifQ.oUD_0r5Q1H_akjeJFWYAxbcr2fckBEb7M25wVJw432Y"
	port  = 12345
	data  = "some data"
)

var (
	client crudappv1connect.CrudAppServiceClient
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
	defer func() {
		_ = container.Terminate(context.Background())
	}()

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	containerPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	connString := fmt.Sprintf("postgres://postgres:password@%s:%s/postgres", host, containerPort.Port())

	go func() {
		run(ctx, &config{
			Port:                port,
			JWTSecret:           "PMBrjiOH5RMo6nQHidA62XctWGxDG0rw",
			PostgresConnString:  connString,
			PostgresAutoMigrate: true,
		})
	}()

	client = crudappv1connect.NewCrudAppServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://localhost:%d", port),
	)

	// Until we have a health endpoint
	cfg := backoff.NewExponentialBackOff()
	cfg.MaxElapsedTime = 3 * time.Second
	err = backoff.Retry(func() error {
		_, err := client.ReadAll(ctx, createRequest(&pb.ReadAllRequest{}))
		if err != nil {
			return err
		}
		return nil
	}, cfg)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestAPI(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		req := createRequest(&pb.CreateRequest{Data: data})
		res, err := client.Create(context.Background(), req)
		require.NoError(t, err)

		post := res.Msg.GetPost()

		require.NotEmpty(t, post.GetPostId())
		require.NotEmpty(t, post.GetCreatedAt())
	})

	t.Run("read", func(t *testing.T) {
		createReq := createRequest(&pb.CreateRequest{Data: data})
		createRes, err := client.Create(context.Background(), createReq)
		require.NoError(t, err)

		readReq := createRequest(&pb.ReadRequest{PostId: createRes.Msg.Post.PostId})
		readRes, err := client.Read(context.Background(), readReq)
		require.NoError(t, err)

		post := readRes.Msg.GetPost()

		require.Equal(t, post.GetData(), data)
	})

	t.Run("read not exist", func(t *testing.T) {
		req := createRequest(&pb.ReadRequest{PostId: "foo"})
		_, err := client.Read(context.Background(), req)
		require.ErrorContains(t, err, "post does not exist")
	})

	t.Run("upsert", func(t *testing.T) {
		ctx := context.Background()

		createReq := createRequest(&pb.CreateRequest{Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		createdPost := createRes.Msg.GetPost()
		newData := "new Data"

		upsertReq := createRequest(&pb.UpsertRequest{
			PostId: createdPost.GetPostId(),
			Data:   newData,
		})
		_, err = client.Upsert(ctx, upsertReq)
		require.NoError(t, err)

		readReq := createRequest(&pb.ReadRequest{
			PostId: createdPost.GetPostId(),
		})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		post := readRes.Msg.GetPost()

		require.Equal(t, post.GetData(), newData)
	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()

		createReq := createRequest(&pb.CreateRequest{Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		createdPost := createRes.Msg.GetPost()

		deleteReq := createRequest(&pb.DeleteRequest{PostId: createdPost.GetPostId()})
		_, err = client.Delete(ctx, deleteReq)
		require.NoError(t, err)

		// Now try to read the deleted record; it should not exist.
		readReq := createRequest(&pb.ReadRequest{PostId: createdPost.GetPostId()})
		_, err = client.Read(ctx, readReq)
		require.ErrorContains(t, err, "post does not exist")
	})

	t.Run("delete not exist", func(t *testing.T) {
		req := createRequest(&pb.DeleteRequest{PostId: "foo"})
		_, err := client.Delete(context.Background(), req)
		require.NoError(t, err)
	})
}

func createRequest[T any](t *T) *connect.Request[T] {
	req := connect.NewRequest(t)
	req.Header().Add("Authentication", fmt.Sprintf("Bearer %s", token))
	return req
}
