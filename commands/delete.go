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
)

type deleteCommand struct {
	cache   cache.Cache
	storage storage.Storage
	tracer  trace.Tracer
}

func NewDeleteCommand(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *deleteCommand {
	return &deleteCommand{
		cache:   cache,
		storage: storage,
		tracer:  tracer,
	}
}

func (c *deleteCommand) Execute(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	userID := req.GetUserId()
	postID := req.GetPostId()
	ctx, span := c.tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := c.storage.Delete(ctx, userID, postID); err != nil {
		telemetry.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}
	c.cache.Remove(ctx, userID, postID)

	return &pb.DeleteResponse{}, nil
}
