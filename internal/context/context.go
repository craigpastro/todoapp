package context

import "context"

type ctxKey string

var userIDCtxKey = ctxKey("user-id-ctx-key")

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
