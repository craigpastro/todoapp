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

type createCommand struct {
	storage storage.Storage
	tracer  trace.Tracer
}

func NewCreateCommand(storage storage.Storage, tracer trace.Tracer) *createCommand {
	return &createCommand{
		storage: storage,
		tracer:  tracer,
	}
}

func (c *createCommand) Execute(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	userID := req.UserId
	ctx, span := c.tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	postID, createdAt, err := c.storage.Create(ctx, userID, req.Data)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return &pb.CreateResponse{
		PostId:    postID,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}
