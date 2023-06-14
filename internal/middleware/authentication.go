package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

var userIDCtxKey = ctxKey("user-id-ctx-key")

func GetUserIDFromCtx(ctx context.Context) string {
	userID := ctx.Value(userIDCtxKey).(string)
	if userID == "" {
		// should never happen so panic
		panic("user id is empty")
	}

	return userID
}

func NewAuthenticationInterceptor(secret string) connect.UnaryInterceptorFunc {
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
				fmt.Println(">>>", err)
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("unauthenticated"),
				)
			}

			sub, err := t.Claims.GetSubject()
			if err != nil || sub == "" {
				fmt.Println(">>>", sub, err)
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("unauthenticated"),
				)
			}

			ctx = context.WithValue(ctx, userIDCtxKey, sub)

			return next(ctx, req)
		})
	})
}
