package middleware

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/todoapp/internal/server"
	"golang.org/x/exp/slog"
)

func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			res, err := next(ctx, req)

			fields := []any{
				"procedure", req.Spec().Procedure,
				"took", time.Since(start),
				"req", req.Any(),
			}

			if err != nil {
				fields = append(fields, "error", err.Error())

				if e, ok := err.(*server.ServerError); ok && e.Internal != nil {
					fields = append(fields, "internal_error", e.Internal.Error())
				}

				slog.ErrorCtx(ctx, "req_error", fields...)
				return nil, err
			}

			fields = append(fields, "res", res.Any())

			slog.InfoCtx(ctx, "req_complete", fields...)

			return res, nil
		})
	})
}
