package middleware

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"
)

func NewLoggingInterceptor(logger *zap.Logger) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			res, err := next(ctx, req)

			fields := []zap.Field{
				zap.String("procedure", req.Spec().Procedure),
				zap.Duration("took", time.Since(start)),
				zap.Any("req", req.Any()),
				zap.Any("res", res.Any()),
			}
			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("rpc_error", fields...)
				return nil, err
			}

			logger.Info("rpc_complete", fields...)

			return res, nil
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
