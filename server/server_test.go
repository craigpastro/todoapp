package server

import (
	"context"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/craigpastro/crudapp/cache"
	pb "github.com/craigpastro/crudapp/gen/proto/api/v1"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage/memory"
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
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	resp, err := client.Create(ctx, &pb.CreateRequest{UserId: myid.New(), Data: data})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if resp.PostId == "" {
		t.Error("PostId is somehow empty")
	}
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, _ := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	resp, _ := client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})

	if resp.Data != data {
		t.Errorf("unexpected data: got '%s', but wanted '%s'", resp.Data, data)
	}
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	_, err = client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: "1"})
	if !strings.Contains(err.Error(), "Post does not exist") {
		t.Errorf("unexpected error: got '%v', but wanted 'Post does not exist'", err)
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, _ := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	newData := "new Data"
	client.Update(ctx, &pb.UpdateRequest{UserId: userID, PostId: create.PostId, Data: newData})
	resp, _ := client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})

	if resp.Data != newData {
		t.Errorf("wrong data: got '%s', but wanted '%s'", resp.Data, newData)
	}

	if resp.CreatedAt.AsTime().After(resp.UpdatedAt.AsTime()) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestUpdateNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	_, err = client.Update(ctx, &pb.UpdateRequest{UserId: userID, PostId: "1", Data: data})
	if !strings.Contains(err.Error(), "Post does not exist") {
		t.Errorf("unexpected error: got '%v', but wanted 'Post does not exist'", err)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	create, _ := client.Create(ctx, &pb.CreateRequest{UserId: userID, Data: data})
	client.Delete(ctx, &pb.DeleteRequest{UserId: userID, PostId: create.PostId})

	// Now try to read the deleted record; it should not exist.
	_, err = client.Read(ctx, &pb.ReadRequest{UserId: userID, PostId: create.PostId})
	if !strings.Contains(err.Error(), "Post does not exist") {
		t.Errorf("unexpected error: got '%v', but wanted 'Post does not exist'", err)
	}
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), opt)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceClient(conn)

	userID := myid.New()
	postID := myid.New()
	if _, err := client.Delete(ctx, &pb.DeleteRequest{UserId: userID, PostId: postID}); err != nil {
		t.Errorf("error not nil: %v", err)
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
