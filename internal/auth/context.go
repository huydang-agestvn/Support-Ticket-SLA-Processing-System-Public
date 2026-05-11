package auth

import (
	"context"
	"errors"
)

type contextKey string

const userContextKey contextKey = "current_user"

var ErrUserNotFoundInContext = errors.New("user not found in context")

func WithUser(ctx context.Context, user UserPrincipal) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) (UserPrincipal, error) {
	user, ok := ctx.Value(userContextKey).(UserPrincipal)
	if !ok {
		return UserPrincipal{}, ErrUserNotFoundInContext
	}

	return user, nil
}