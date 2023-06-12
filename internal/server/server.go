package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/internal/errors"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/instrumentation"
	"github.com/craigpastro/crudapp/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("internal/server")

type server struct {
	crudappv1connect.UnimplementedCrudAppServiceHandler

	Storage storage.Storage
}

func NewServer(storage storage.Storage) *server {
	return &server{
		Storage: storage,
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
	msg := req.Msg
	if err := validate(msg); err != nil {
		return nil, err
	}

	userID := msg.GetUserId()
	ctx, span := tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	post, err := s.Storage.Create(ctx, userID, msg.GetData())
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		Post: post,
	}), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	msg := req.Msg
	if err := validate(msg); err != nil {
		return nil, err
	}

	userID := msg.GetUserId()
	postID := msg.GetPostId()
	ctx, span := tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	post, err := s.Storage.Read(ctx, userID, postID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.ReadResponse{
		Post: post,
	}), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest]) (*connect.Response[pb.ReadAllResponse], error) {
	msg := req.Msg
	if err := validate(msg); err != nil {
		return nil, err
	}

	userID := msg.GetUserId()
	ctx, span := tracer.Start(ctx, "ReadAll", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	posts, lastIndex, err := s.Storage.ReadAll(ctx, userID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.ReadAllResponse{
		Posts:     posts,
		LastIndex: lastIndex,
	}), nil
}

func (s *server) Upsert(ctx context.Context, req *connect.Request[pb.UpsertRequest]) (*connect.Response[pb.UpsertResponse], error) {
	msg := req.Msg
	if err := validate(msg); err != nil {
		return nil, err
	}

	userID := msg.GetUserId()
	postID := msg.GetPostId()

	ctx, span := tracer.Start(ctx, "Update", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	post, err := s.Storage.Upsert(ctx, userID, postID, msg.GetData())
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.UpsertResponse{
		Post: post,
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	msg := req.Msg
	if err := validate(msg); err != nil {
		return nil, err
	}

	userID := msg.GetUserId()
	postID := msg.GetPostId()
	ctx, span := tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := s.Storage.Delete(ctx, userID, postID); err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.DeleteResponse{}), nil
}
