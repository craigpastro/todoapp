package server

import (
	"context"
	"errors"

	pb "github.com/craigpastro/crudapp/api/proto/v1"
	"github.com/craigpastro/crudapp/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedServiceServer

	Storage storage.Storage
}

func NewServer(storage storage.Storage) *server {
	return &server{
		Storage: storage,
	}
}

func (s *server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	postID, createdAt, err := s.Storage.Create(ctx, in.UserId, in.Data)
	if err != nil {
		return nil, handleStorageError(err)
	}

	return &pb.CreateResponse{
		PostId:    postID,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

func (s *server) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	record, err := s.Storage.Read(ctx, in.UserId, in.PostId)
	if err != nil {
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
	records, err := s.Storage.ReadAll(ctx, in.UserId)
	if err != nil {
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
	updatedAt, err := s.Storage.Update(ctx, in.UserId, in.PostId, in.Data)
	if err != nil {
		return nil, handleStorageError(err)
	}

	return &pb.UpdateResponse{
		PostId:    in.PostId,
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}

func (s *server) Delete(ctx context.Context, in *pb.DeleteRequest) (*emptypb.Empty, error) {
	if err := s.Storage.Delete(ctx, in.UserId, in.PostId); err != nil {
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
