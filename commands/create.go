package commands

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/errors"
	pb "github.com/craigpastro/crudapp/gen/proto/api/v1"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type createCommand struct {
	cache   cache.Cache
	storage storage.Storage
	tracer  trace.Tracer
}

func NewCreateCommand(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *createCommand {
	return &createCommand{
		cache:   cache,
		storage: storage,
		tracer:  tracer,
	}
}

func (c *createCommand) Execute(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	userID := req.UserId
	ctx, span := c.tracer.Start(ctx, "Create", trace.WithAttributes(attribute.String("userID", userID)))
	defer span.End()

	record, err := c.storage.Create(ctx, userID, req.Data)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return &pb.CreateResponse{
		PostId:    record.PostID,
		CreatedAt: timestamppb.New(record.CreatedAt),
	}, nil
}
