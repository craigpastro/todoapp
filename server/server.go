package server

import (
	"context"
	"errors"

	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedServiceServer

	Tracer  trace.Tracer
	Storage storage.Storage
}

func NewServer(tracer trace.Tracer, storage storage.Storage) *server {
	return &server{
		Tracer:  tracer,
		Storage: storage,
	}
}

func (s *server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	userID := in.UserId
	ctx, span := s.Tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	postID, createdAt, err := s.Storage.Create(ctx, userID, in.Data)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, handleStorageError(err)
	}

	return &pb.CreateResponse{
		PostId:    postID,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

func (s *server) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	userID := in.UserId
	postID := in.PostId
	ctx, span := s.Tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	record, err := s.Storage.Read(ctx, userID, postID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, handleStorageError(err)
	}

	return &pb.ReadResponse{
		UserId:    record.UserID,
		PostId:    record.PostID,
		Data:      record.Data,
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}, nil
}

func (s *server) ReadAll(ctx context.Context, in *pb.ReadAllRequest) (*pb.ReadAllResponse, error) {
	userID := in.UserId
	ctx, span := s.Tracer.Start(ctx, "ReadAll", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	records, err := s.Storage.ReadAll(ctx, userID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, handleStorageError(err)
	}

	posts := []*pb.ReadResponse{}
	for _, record := range records {
		posts = append(posts, &pb.ReadResponse{
			UserId:    record.UserID,
			PostId:    record.PostID,
			Data:      record.Data,
			CreatedAt: timestamppb.New(record.CreatedAt),
			UpdatedAt: timestamppb.New(record.UpdatedAt),
		})
	}

	return &pb.ReadAllResponse{Posts: posts}, nil
}

func (s *server) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	userID := in.UserId
	postID := in.PostId
	ctx, span := s.Tracer.Start(ctx, "Update", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	updatedAt, err := s.Storage.Update(ctx, userID, postID, in.Data)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, handleStorageError(err)
	}

	return &pb.UpdateResponse{
		PostId:    in.PostId,
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}

func (s *server) Delete(ctx context.Context, in *pb.DeleteRequest) (*emptypb.Empty, error) {
	userID := in.UserId
	postID := in.PostId
	ctx, span := s.Tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := s.Storage.Delete(ctx, userID, postID); err != nil {
		instrumentation.TraceError(span, err)
		return &emptypb.Empty{}, handleStorageError(err)
	}

	return &emptypb.Empty{}, nil
}

func handleStorageError(err error) error {
	if errors.Is(err, storage.ErrPostDoesNotExist) {
		return status.Error(codes.InvalidArgument, "Post does not exist")
	}

	return status.Error(codes.Internal, "Internal server error")
}
