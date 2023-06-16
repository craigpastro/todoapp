package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	ctxpkg "github.com/craigpastro/crudapp/internal/context"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/gen/sqlc"
	"github.com/craigpastro/crudapp/internal/instrumentation"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	tracer = otel.Tracer("internal/server")

	ErrPostDoesNotExist = errors.New("post does not exist")
)

type server struct {
	crudappv1connect.UnimplementedCrudAppServiceHandler

	queries *sqlc.Queries
}

func NewServer(queries *sqlc.Queries) *server {
	return &server{
		queries: queries,
	}
}

func (s *server) Create(ctx context.Context, req *connect.Request[pb.CreateRequest]) (*connect.Response[pb.CreateResponse], error) {
	ctx, span := tracer.Start(ctx, "Create")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)

	post, err := s.queries.Create(ctx, sqlc.CreateParams{
		UserID: userID,
		Data:   req.Msg.GetData(),
	})
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		Post: &pb.Post{
			UserId:    post.UserID,
			PostId:    post.PostID,
			Data:      post.Data,
			CreatedAt: timestamppb.New(post.CreatedAt.Time),
			UpdatedAt: timestamppb.New(post.UpdatedAt.Time),
		},
	}), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	ctx, span := tracer.Start(ctx, "Read")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)
	postID := req.Msg.GetPostId()

	row, err := s.queries.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		PostID: postID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, ErrPostDoesNotExist))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.ReadResponse{
		Post: &pb.Post{
			UserId:    row.UserID,
			PostId:    row.PostID,
			Data:      row.Data,
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		},
	}), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest]) (*connect.Response[pb.ReadAllResponse], error) {
	ctx, span := tracer.Start(ctx, "ReadAll")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)

	rows, err := s.queries.ReadPage(ctx, sqlc.ReadPageParams{
		UserID: userID,
	})
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	var lastIndex int64
	posts := make([]*pb.Post, 0, len(rows))
	for _, row := range rows {
		lastIndex = row.ID

		posts = append(posts, &pb.Post{
			UserId:    row.UserID,
			PostId:    row.PostID,
			Data:      row.Data,
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		})
	}

	return connect.NewResponse(&pb.ReadAllResponse{
		Posts:     posts,
		LastIndex: lastIndex,
	}), nil
}

func (s *server) Upsert(ctx context.Context, req *connect.Request[pb.UpsertRequest]) (*connect.Response[pb.UpsertResponse], error) {
	ctx, span := tracer.Start(ctx, "Update")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)
	msg := req.Msg
	postID := msg.GetPostId()

	row, err := s.queries.Upsert(ctx, sqlc.UpsertParams{
		UserID: userID,
		PostID: postID,
		Data:   msg.GetData(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, ErrPostDoesNotExist))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.UpsertResponse{
		Post: &pb.Post{
			UserId:    row.UserID,
			PostId:    row.PostID,
			Data:      row.Data,
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		},
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)
	postID := req.Msg.GetPostId()

	ctx, span := tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := s.queries.Delete(ctx, sqlc.DeleteParams{
		UserID: userID,
		PostID: postID,
	}); err != nil {
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
