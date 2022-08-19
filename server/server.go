package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/errors"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	msg := req.Msg
	userID := msg.GetUserId()
	ctx, span := s.Tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	record, err := s.Storage.Create(ctx, userID, msg.GetData())
	if err != nil {
		telemetry.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		PostId:    record.PostID,
		CreatedAt: timestamppb.New(record.CreatedAt),
	}), nil
}

func (s *server) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	msg := req.Msg
	userID := msg.GetUserId()
	postID := msg.GetPostId()
	ctx, span := s.Tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	var err error
	record, ok := s.Cache.Get(ctx, userID, postID)
	if !ok {
		record, err = s.Storage.Read(ctx, userID, postID)
		if err != nil {
			telemetry.TraceError(span, err)
			return nil, errors.HandleStorageError(err)
		}
		s.Cache.Add(ctx, userID, postID, record)
	}

	return connect.NewResponse(&pb.ReadResponse{
		UserId:    record.UserID,
		PostId:    record.PostID,
		Data:      record.Data,
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}), nil
}

func (s *server) ReadAll(ctx context.Context, req *connect.Request[pb.ReadAllRequest], stream *connect.ServerStream[pb.ReadAllResponse]) error {
	if err := validate(req.Msg); err != nil {
		return err
	}

	msg := req.Msg
	userID := msg.GetUserId()
	ctx, span := s.Tracer.Start(ctx, "ReadAll", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	iter, err := s.Storage.ReadAll(ctx, userID)
	if err != nil {
		telemetry.TraceError(span, err)
		return errors.HandleStorageError(err)
	}

	for iter.Next(ctx) {
		var record storage.Record
		if err := iter.Get(&record); err != nil {
			telemetry.TraceError(span, err)
			return errors.HandleStorageError(err)
		}

		stream.Send(&pb.ReadAllResponse{
			UserId:    record.UserID,
			PostId:    record.PostID,
			Data:      record.Data,
			CreatedAt: timestamppb.New(record.CreatedAt),
			UpdatedAt: timestamppb.New(record.UpdatedAt),
		})
	}

	iter.Close(ctx)

	return nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[pb.UpdateRequest]) (*connect.Response[pb.UpdateResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	msg := req.Msg
	userID := msg.GetUserId()
	postID := msg.GetPostId()
	ctx, span := s.Tracer.Start(ctx, "Update", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	updatedAt, err := s.Storage.Update(ctx, userID, postID, msg.GetData())
	if err != nil {
		telemetry.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}
	s.Cache.Remove(ctx, userID, postID)

	return connect.NewResponse(&pb.UpdateResponse{
		PostId:    msg.PostId,
		UpdatedAt: timestamppb.New(updatedAt),
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	msg := req.Msg
	userID := msg.GetUserId()
	postID := msg.GetPostId()
	ctx, span := s.Tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := s.Storage.Delete(ctx, userID, postID); err != nil {
		telemetry.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}
	s.Cache.Remove(ctx, userID, postID)

	return connect.NewResponse(&pb.DeleteResponse{}), nil
}
