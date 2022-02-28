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

type readCommand struct {
	storage storage.Storage
	tracer  trace.Tracer
}

func NewReadCommand(storage storage.Storage, tracer trace.Tracer) *readCommand {
	return &readCommand{
		storage: storage,
		tracer:  tracer,
	}
}

func (c *readCommand) Execute(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	userID := req.UserId
	postID := req.PostId
	ctx, span := c.tracer.Start(ctx, "Read", trace.WithAttributes(attribute.String("userID", userID), attribute.String("postID", postID)))
	defer span.End()

	record, err := c.storage.Read(ctx, userID, postID)
	if err != nil {
		instrumentation.TraceError(span, err)
		return nil, errors.HandleStorageError(err)
	}

	return &pb.ReadResponse{
		UserId:    record.UserID,
		PostId:    record.PostID,
		Data:      record.Data,
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}, nil
}
