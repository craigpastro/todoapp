package middleware

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/telemetry"
)

func NewLoggingInterceptor(logger telemetry.Logger) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err != nil {
				return nil, err
			}

			logger.Info("res", telemetry.String("procedure", req.Spec().Procedure), telemetry.Any("req", req.Any()), telemetry.Any("res", res.Any()))

			return res, nil
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
