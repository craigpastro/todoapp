package commands

import (
	"context"

	"github.com/craigpastro/crudapp/errors"
	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateCommand struct {
	storage storage.Storage
	tracer  trace.Tracer
}

func NewUpdateCommand(storage storage.Storage, tracer trace.Tracer) *updateCommand {
	return &updateCommand{
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
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return &pb.UpdateResponse{
		PostId:    req.PostId,
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}
