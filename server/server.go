package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/commands"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

type server struct {
	crudappv1connect.UnimplementedCrudAppServiceHandler

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

type validator interface {
	Validate() error
}

func validate[T validator](msg T) error {
	if err := msg.Validate(); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

func (s *server) Create(ctx context.Context, req *connect.Request[pb.CreateRequest]) (*connect.Response[pb.CreateResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	c := commands.NewCreateCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	c := commands.NewReadCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest]) (*connect.Response[pb.ReadAllResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	c := commands.NewReadAllCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[pb.UpdateRequest]) (*connect.Response[pb.UpdateResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	c := commands.NewUpdateCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	c := commands.NewDeleteCommand(s.Cache, s.Storage, s.Tracer)
	res, err := c.Execute(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
