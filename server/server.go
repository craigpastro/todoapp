package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/commands"
	pb "github.com/craigpastro/crudapp/internal/gen/api/v1"
	"github.com/craigpastro/crudapp/internal/gen/api/v1/v1connect"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

type server struct {
	v1connect.UnimplementedCrudAppServiceHandler

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

func (s *server) Create(ctx context.Context, req *connect.Request[pb.CreateRequest]) (*connect.Response[pb.CreateResponse], error) {
	c := commands.NewCreateCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	c := commands.NewReadCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest]) (*connect.Response[pb.ReadAllResponse], error) {
	c := commands.NewReadAllCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[pb.UpdateRequest]) (*connect.Response[pb.UpdateResponse], error) {
	c := commands.NewUpdateCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	c := commands.NewDeleteCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
