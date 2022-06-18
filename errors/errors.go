package errors

import (
	"errors"

	"github.com/craigpastro/crudapp/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleStorageError(err error) error {
	if errors.Is(err, storage.ErrPostDoesNotExist) {
		return status.Error(codes.InvalidArgument, "Post does not exist")
	}

	return status.Error(codes.Internal, "Internal server error")
}
