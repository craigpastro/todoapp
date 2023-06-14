package middleware

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/internal/server"
	"go.uber.org/zap"
)

func NewLoggingInterceptor(logger *zap.Logger) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			res, err := next(ctx, req)

			fields := []zap.Field{
				zap.String("procedure", req.Spec().Procedure),
				zap.Duration("took", time.Since(start)),
				zap.Any("req", req.Any()),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				if e, ok := err.(*server.ServerError); ok && e.Internal != nil {
					fields = append(fields, zap.String("internal_error", e.Internal.Error()))
				}

				logger.Error("rpc_error", fields...)
				return nil, err
			}

			fields = append(fields, zap.Any("res", res.Any()))

			logger.Info("rpc_complete", fields...)

			return res, nil
		})
	})
}
