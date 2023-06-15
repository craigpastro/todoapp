package context

import (
	"context"

	"github.com/craigpastro/crudapp/internal/gen/sqlc"
)

type ctxKey string

var userIDCtxKey = ctxKey("user-id-ctx-key")
var txCtxKey = ctxKey("tx-ctx-key")

func SetUserIDInCtx(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

func GetUserIDFromCtx(ctx context.Context) string {
	userID := ctx.Value(userIDCtxKey).(string)
	if userID == "" {
		// should never happen so panic
		panic("user id is empty")
	}

	return userID
}

func SetQueriesInCtx(ctx context.Context, q *sqlc.Queries) context.Context {
	return context.WithValue(ctx, txCtxKey, q)
}

func GetQueriesFromCtx(ctx context.Context) *sqlc.Queries {
	q, ok := ctx.Value(txCtxKey).(*sqlc.Queries)
	if !ok {
		// should never happen so panic
		panic("queries is empty")
	}

	return q
}
