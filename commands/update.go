package commands

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/errors"
	pb "github.com/craigpastro/crudapp/internal/gen/crudapp/v1"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateCommand struct {
	cache   cache.Cache
	storage storage.Storage
	tracer  trace.Tracer
}

func NewUpdateCommand(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *updateCommand {
	return &updateCommand{
		cache:   cache,
		storage: storage,
		tracer:  tracer,
	}
}

func (c *updateCommand) Execute(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	userID := req.UserId
	postID := req.PostId
	ctx, span := c.tracer.Start(ctx, "Update", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	updatedAt, err := c.storage.Update(ctx, userID, postID, req.Data)
	if err != nil {
		telemetry.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}
	c.cache.Remove(ctx, userID, postID)

	return &pb.UpdateResponse{
		PostId:    req.PostId,
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}
