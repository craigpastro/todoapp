package server

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/cache"
	pb "github.com/craigpastro/crudapp/gen/proto/api/v1"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const (
	bufSize = 1024 * 1024
	data    = "some data"
)

var lis *bufconn.Listener

func TestMain(m *testing.M) {
	s := grpc.NewServer()
	tracer := instrumentation.NewNoopTracer()
	storage := memory.New(tracer)
	cache := cache.NewNoopCache()
	pb.RegisterServiceServer(s, NewServer(cache, storage, tracer))
	lis = bufconn.Listen(bufSize)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	resp, err := client.Create(ctx, &pb.CreateRequest{UserId: myid.New(), Data: data})
	require.NoError(t, err)
	require.NotEmpty(t, resp.PostId)
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, err := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	require.NoError(t, err)
	resp, err := client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})
	require.NoError(t, err)

	require.Equal(t, resp.Data, data, "got '%s', want '%s'", resp.Data, data)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	_, err = client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: "1"})

	require.ErrorContains(t, err, "Post does not exist")
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, err := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	require.NoError(t, err)

	newData := "new Data"
	_, err = client.Update(ctx, &pb.UpdateRequest{UserId: userID, PostId: create.PostId, Data: newData})
	require.NoError(t, err)
	resp, err := client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})
	require.NoError(t, err)

	require.Equal(t, resp.Data, newData, "got '%s', want '%s'", resp.Data, newData)
	require.True(t, resp.CreatedAt.AsTime().Before(resp.UpdatedAt.AsTime()))
}

func TestUpdateNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	_, err = client.Update(ctx, &pb.UpdateRequest{UserId: userID, PostId: "1", Data: data})
	require.ErrorContains(t, err, "Post does not exist")
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, err := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	require.NoError(t, err)
	_, err = client.Delete(ctx, &pb.DeleteRequest{UserId: userID, PostId: create.PostId})
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})
	require.ErrorContains(t, err, "Post does not exist")
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	postID := myid.New()
	_, err = client.Delete(ctx, &pb.DeleteRequest{UserId: userID, PostId: postID})
	require.NoError(t, err)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
