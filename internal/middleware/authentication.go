package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/bufbuild/connect-go"
	ctxpkg "github.com/craigpastro/crudapp/internal/context"
	"github.com/craigpastro/crudapp/internal/gen/sqlc"
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

			// TODO: this should go away after RLS
			ctx = ctxpkg.SetUserIDInCtx(ctx, sub)

			conn, err := pool.Acquire(ctx)
			if err != nil {
				// TODO: refactor errors to their own package and handle them in middleware
				return nil, connect.NewError(
					connect.CodeInternal,
					errors.New("internal server error"),
				)
			}
			defer conn.Release()

			// tx, err := conn.Begin(ctx)
			// if err != nil {
			// 	return nil, connect.NewError(
			// 		connect.CodeInternal,
			// 		errors.New("internal server error"),
			// 	)
			// }
			// defer tx.Rollback(context.Background())
			// q := sqlc.New(conn).WithTx(tx)

			q := sqlc.New(conn)
			q.Foo(ctx, sub)
			ctx = ctxpkg.SetQueriesInCtx(ctx, q)

			resp, err := next(ctx, req)
			if err != nil {
				return nil, err
			}

			// tx.Commit(ctx)

			return resp, nil
		})
	})
}
