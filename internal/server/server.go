package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	ctxpkg "github.com/craigpastro/crudapp/internal/context"
	"github.com/craigpastro/crudapp/internal/gen/sqlc"
	pb "github.com/craigpastro/crudapp/internal/gen/todoapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/todoapp/v1/todoappv1connect"
	"github.com/craigpastro/crudapp/internal/instrumentation"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	tracer = otel.Tracer("internal/server")

	ErrTodoIDDoesNotExist = errors.New("todo id does not exist")
)

type server struct {
	todoappv1connect.UnimplementedTodoAppServiceHandler

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
		Todo:   req.Msg.GetTodo(),
	})
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		Todo: &pb.Todo{
			UserId:    post.UserID,
			TodoId:    post.TodoID,
			Todo:      post.Todo,
			CreatedAt: timestamppb.New(post.CreatedAt.Time),
			UpdatedAt: timestamppb.New(post.UpdatedAt.Time),
		},
	}), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	ctx, span := tracer.Start(ctx, "Read")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)
	todoID := req.Msg.GetTodoId()

	row, err := s.queries.Read(ctx, sqlc.ReadParams{
		UserID: userID,
		TodoID: todoID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, ErrTodoIDDoesNotExist))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.ReadResponse{
		Todo: &pb.Todo{
			UserId:    row.UserID,
			TodoId:    row.TodoID,
			Todo:      row.Todo,
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
	todos := make([]*pb.Todo, 0, len(rows))
	for _, row := range rows {
		lastIndex = row.ID

		todos = append(todos, &pb.Todo{
			UserId:    row.UserID,
			TodoId:    row.TodoID,
			Todo:      row.Todo,
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		})
	}

	return connect.NewResponse(&pb.ReadAllResponse{
		Todos:     todos,
		LastIndex: lastIndex,
	}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[pb.UpdateRequest]) (*connect.Response[pb.UpdateResponse], error) {
	ctx, span := tracer.Start(ctx, "Update")
	defer span.End()

	userID := ctxpkg.GetUserIDFromCtx(ctx)
	msg := req.Msg
	todoID := msg.GetTodoId()

	row, err := s.queries.Update(ctx, sqlc.UpdateParams{
		UserID: userID,
		TodoID: todoID,
		Todo:   msg.GetTodo(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newPublicError(connect.NewError(connect.CodeInvalidArgument, ErrTodoIDDoesNotExist))
		}

		instrumentation.TraceError(span, err)
		return nil, newInternalError(err)
	}

	return connect.NewResponse(&pb.UpdateResponse{
		Todo: &pb.Todo{
			UserId:    row.UserID,
			TodoId:    row.TodoID,
			Todo:      row.Todo,
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		},
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	userID := ctxpkg.GetUserIDFromCtx(ctx)
	todoID := req.Msg.GetTodoId()

	ctx, span := tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", todoID)))
	defer span.End()

	if err := s.queries.Delete(ctx, sqlc.DeleteParams{
		UserID: userID,
		TodoID: todoID,
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
