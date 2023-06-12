package storage

import (
	"context"
	"errors"

	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
)

var ErrPostDoesNotExist = errors.New("post does not exist")

type Storage interface {
	Create(ctx context.Context, userID, data string) (*pb.Post, error)
	Read(ctx context.Context, userID, postID string) (*pb.Post, error)
	ReadAll(ctx context.Context, userID string) ([]*pb.Post, int64, error)
	Upsert(ctx context.Context, userID, postID, data string) (*pb.Post, error)
	Delete(ctx context.Context, userID, postID string) error
}
