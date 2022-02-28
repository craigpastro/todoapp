package server

import (
	"context"

	"github.com/craigpastro/crudapp/commands"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedServiceServer

	Storage storage.Storage
	Tracer  trace.Tracer
}

func NewServer(tracer trace.Tracer, storage storage.Storage) *server {
	return &server{
		Storage: storage,
		Tracer:  tracer,
	}
}

func (s *server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	c := commands.NewCreateCommand(s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	c := commands.NewReadCommand(s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) ReadAll(ctx context.Context, req *pb.ReadAllRequest) (*pb.ReadAllResponse, error) {
	c := commands.NewReadAllCommand(s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	c := commands.NewUpdateCommand(s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}

func (s *server) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	c := commands.NewDeleteCommand(s.Storage, s.Tracer)
	return c.Execute(ctx, req)
}
