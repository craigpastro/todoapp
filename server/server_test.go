package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/internal/gen/api/v1"
	"github.com/craigpastro/crudapp/internal/gen/api/v1/v1connect"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/stretchr/testify/require"
)

const (
	addr = "localhost:12345"
	data = "some data"
)

func TestMain(m *testing.M) {
	cache := cache.NewNoopCache()
	tracer := instrumentation.NewNoopTracer()
	storage := memory.New(tracer)
	mux := http.NewServeMux()
	mux.Handle(v1connect.NewCrudAppServiceHandler(NewServer(cache, storage, tracer)))

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func TestAPI(t *testing.T) {
	client := v1connect.NewCrudAppServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr))

	t.Run("create", func(t *testing.T) {
		req := connect.NewRequest(&pb.CreateRequest{UserId: myid.New(), Data: data})

		res, err := client.Create(context.Background(), req)
		require.NoError(t, err)

		require.NotEmpty(t, res.Msg.PostId)
		require.NotEmpty(t, res.Msg.CreatedAt)
	})

	t.Run("read", func(t *testing.T) {
		ctx := context.Background()
		userID := myid.New()
		createReq := connect.NewRequest(&pb.CreateRequest{UserId: userID, Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		readReq := connect.NewRequest(&pb.ReadRequest{UserId: userID, PostId: createRes.Msg.PostId})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		require.Equal(t, readRes.Msg.Data, data, "got '%s', want '%s'", readRes.Msg.Data, data)
	})

	t.Run("read not exist", func(t *testing.T) {
		req := connect.NewRequest(&pb.ReadRequest{UserId: myid.New(), PostId: "foo"})
		_, err := client.Read(context.Background(), req)
		require.ErrorContains(t, err, "Post does not exist")
	})

	t.Run("update", func(t *testing.T) {
		ctx := context.Background()
		userID := myid.New()

		createReq := connect.NewRequest(&pb.CreateRequest{UserId: userID, Data: data})
		createRes, err := client.Create(ctx, createReq)
		require.NoError(t, err)

		newData := "new Data"
		updateReq := connect.NewRequest(&pb.UpdateRequest{UserId: userID, PostId: createRes.Msg.PostId, Data: newData})
		_, err = client.Update(ctx, updateReq)
		require.NoError(t, err)

		readReq := connect.NewRequest(&pb.ReadRequest{UserId: userID, PostId: createRes.Msg.PostId})
		readRes, err := client.Read(ctx, readReq)
		require.NoError(t, err)

		require.Equal(t, readRes.Msg.Data, newData, "got '%s', want '%s'", readRes.Msg.Data, newData)
	})

	t.Run("update not exist", func(t *testing.T) {
		req := connect.NewRequest(&pb.UpdateRequest{UserId: myid.New(), PostId: "foo", Data: "new data"})
		_, err := client.Update(context.Background(), req)
		require.ErrorContains(t, err, "Post does not exist")
	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()
		userID := myid.New()
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
		req := connect.NewRequest(&pb.DeleteRequest{UserId: myid.New(), PostId: "foo"})
		_, err := client.Delete(context.Background(), req)
		require.NoError(t, err)
	})

}
