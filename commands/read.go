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

type readCommand struct {
	cache   cache.Cache
	storage storage.Storage
	tracer  trace.Tracer
}

func NewReadCommand(cache cache.Cache, storage storage.Storage, tracer trace.Tracer) *readCommand {
	return &readCommand{
		cache:   cache,
		storage: storage,
		tracer:  tracer,
	}
}

func (c *readCommand) Execute(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	userID := req.UserId
	postID := req.PostId
	ctx, span := c.tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	var err error
	record, ok := c.cache.Get(ctx, userID, postID)
	if !ok {
		record, err = c.storage.Read(ctx, userID, postID)
		if err != nil {
			instrumentation.TraceError(span, err)
			return nil, errors.HandleStorageError(err)
		}
		c.cache.Add(ctx, userID, postID, record)
	}

	return &pb.ReadResponse{
		UserId:    record.UserID,
		PostId:    record.PostID,
		Data:      record.Data,
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}, nil
}
