package middleware

import (
	"context"

	"github.com/bufbuild/connect-go"
)

type validator interface {
	Validate() error
}

func NewValidatorInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if r, ok := req.Any().(validator); ok {
				if err := r.Validate(); err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
			}

			return next(ctx, req)
		})
	})
}
