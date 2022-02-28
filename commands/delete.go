package commands

import (
	"context"

	"github.com/craigpastro/crudapp/errors"
	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

type deleteCommand struct {
	storage storage.Storage
	tracer  trace.Tracer
}

func NewDeleteCommand(storage storage.Storage, tracer trace.Tracer) *deleteCommand {
	return &deleteCommand{
		storage: storage,
		tracer:  tracer,
	}
}

func (c *deleteCommand) Execute(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	userID := req.UserId
	postID := req.PostId
	ctx, span := c.tracer.Start(ctx, "Delete", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	if err := c.storage.Delete(ctx, userID, postID); err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return &emptypb.Empty{}, nil
}
