package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	ctxpkg "github.com/craigpastro/crudapp/internal/context"
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

	db storage.Storage
}

func NewServer(db storage.Storage) *server {
	return &server{
		db: db,
	}
}

func (s *server) Create(ctx context.Context, req *connect.Request[pb.CreateRequest]) (*connect.Response[pb.CreateResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)

	ctx, span := tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	post, err := s.db.Create(ctx, userID, req.Msg.GetData())
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		Post: post,
	}), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)
	postID := req.Msg.GetPostId()

	ctx, span := tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	post, err := s.db.Read(ctx, userID, postID)
	if err != nil {
		if errors.Is(err, storage.ErrPostDoesNotExist) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, err))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.ReadResponse{
		Post: post,
	}), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest]) (*connect.Response[pb.ReadAllResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)

	ctx, span := tracer.Start(ctx, "ReadAll", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	posts, lastIndex, err := s.db.ReadAll(ctx, userID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.ReadAllResponse{
		Posts:     posts,
		LastIndex: lastIndex,
	}), nil
}

func (s *server) Upsert(ctx context.Context, req *connect.Request[pb.UpsertRequest]) (*connect.Response[pb.UpsertResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)
	msg := req.Msg
	postID := msg.GetPostId()

	ctx, span := tracer.Start(ctx, "Update", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	post, err := s.db.Upsert(ctx, userID, postID, msg.GetData())
	if err != nil {
		if errors.Is(err, storage.ErrPostDoesNotExist) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, err))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.UpsertResponse{
		Post: post,
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)
	postID := req.Msg.GetPostId()

	ctx, span := tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := s.db.Delete(ctx, userID, postID); err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.DeleteResponse{}), nil
}

type ServerError struct {
	Internal error
	Public   error
}

func (e *ServerError) Error() string {
	return e.Public.Error()
}

func newInternalError(err error) *ServerError {
	return &ServerError{
		Public:   connect.NewError(connect.CodeInternal, errors.New("internal server error")),
		Internal: err,
	}
}

func newPublicError(err error) *ServerError {
	return &ServerError{
		Public: err,
	}
}
