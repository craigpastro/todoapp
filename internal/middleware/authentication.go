package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/bufbuild/connect-go"
	ctxpkg "github.com/craigpastro/crudapp/internal/context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAuthenticationInterceptor(pool *pgxpool.Pool, secret string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			authHeader := strings.Split(req.Header().Get("Authentication"), "Bearer ")
			if len(authHeader) != 2 {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("malformed token"),
				)
			}

			jwtToken := authHeader[1]
			t, err := jwt.Parse(jwtToken, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("unauthenticated"),
				)
			}

			sub, err := t.Claims.GetSubject()
			if err != nil || sub == "" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("unauthenticated"),
				)
			}

			ctx = ctxpkg.SetUserIDInCtx(ctx, sub)

			return next(ctx, req)
		})
	})
}
