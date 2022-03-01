package server

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/commands"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedServiceServer

	Cache   cache.Cache
	Storage storage.Storage
	Tracer  trace.Tracer
}

func NewServer(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *server {
	return &server{
		Cache:   cache,
		Storage: storage,
		Tracer:  tracer,
	}
}

func (s *server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	c := commands.NewCreateCommand(s.Cache, s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	c := commands.NewReadCommand(s.Cache, s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) ReadAll(ctx context.Context, req *pb.ReadAllRequest) (*pb.ReadAllResponse, error) {
	c := commands.NewReadAllCommand(s.Cache, s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	c := commands.NewUpdateCommand(s.Cache, s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	c := commands.NewDeleteCommand(s.Cache, s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}
