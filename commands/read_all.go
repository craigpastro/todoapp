package commands

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/errors"
	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type readAllCommand struct {
	cache   cache.Cache
	storage storage.Storage
	tracer  trace.Tracer
}

func NewReadAllCommand(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *readAllCommand {
	return &readAllCommand{
		cache:   cache,
		storage: storage,
		tracer:  tracer,
	}
}

func (c *readAllCommand) Execute(ctx context.Context, req *pb.ReadAllRequest) (*pb.ReadAllResponse, error) {
	userID := req.UserId
	ctx, span := c.tracer.Start(ctx, "ReadAll", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	records, err := c.storage.ReadAll(ctx, userID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
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
