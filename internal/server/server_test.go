package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/storage/postgres"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	addr = "localhost:12345"
	data = "some data"
)

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

	db := postgres.MustNew(connString, true)
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle(crudappv1connect.NewCrudAppServiceHandler(NewServer(db)))

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func TestAPI(t *testing.T) {
	client := crudappv1connect.NewCrudAppServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr))

	t.Run("create", func(t *testing.T) {
		req := connect.NewRequest(&pb.CreateRequest{UserId: ulid.Make().String(), Data: data})

		res, err := client.Create(context.Background(), req)
		require.NoError(t, err)

		require.NotEmpty(t, res.Msg.PostId)
		require.NotEmpty(t, res.Msg.CreatedAt)
	})

	t.Run("read", func(t *testing.T) {
		ctx := context.Background()
		userID := ulid.Make().String()
		createReq := connect.NewRequest(&pb.CreateRequest{UserId: userID, Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		readReq := connect.NewRequest(&pb.ReadRequest{UserId: userID, PostId: createRes.Msg.PostId})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		require.Equal(t, readRes.Msg.Data, data, "got '%s', want '%s'", readRes.Msg.Data, data)
	})

	t.Run("read not exist", func(t *testing.T) {
		req := connect.NewRequest(&pb.ReadRequest{UserId: ulid.Make().String(), PostId: "foo"})
		_, err := client.Read(context.Background(), req)
		require.ErrorContains(t, err, "Post does not exist")
	})

	t.Run("upsert", func(t *testing.T) {
		ctx := context.Background()
		userID := ulid.Make().String()

		createReq := connect.NewRequest(&pb.CreateRequest{UserId: userID, Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		newData := "new Data"
		upsertReq := connect.NewRequest(&pb.UpsertRequest{UserId: userID, PostId: createRes.Msg.PostId, Data: newData})
		_, err = client.Upsert(ctx, upsertReq)
		require.NoError(t, err)

		readReq := connect.NewRequest(&pb.ReadRequest{UserId: userID, PostId: createRes.Msg.PostId})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		require.Equal(t, readRes.Msg.Data, newData, "got '%s', want '%s'", readRes.Msg.Data, newData)
	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()
		userID := ulid.Make().String()
		createReq := connect.NewRequest(&pb.CreateRequest{UserId: userID, Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		deleteReq := connect.NewRequest(&pb.DeleteRequest{UserId: userID, PostId: createRes.Msg.PostId})
		_, err = client.Delete(ctx, deleteReq)
		require.NoError(t, err)

		// Now try to read the deleted record; it should not exist.
		readReq := connect.NewRequest(&pb.ReadRequest{UserId: userID, PostId: createRes.Msg.PostId})
		_, err = client.Read(ctx, readReq)
		require.ErrorContains(t, err, "Post does not exist")
	})

	t.Run("delete not exist", func(t *testing.T) {
		req := connect.NewRequest(&pb.DeleteRequest{UserId: ulid.Make().String(), PostId: "foo"})
		_, err := client.Delete(context.Background(), req)
		require.NoError(t, err)
	})

}
