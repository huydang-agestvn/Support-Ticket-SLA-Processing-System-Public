package auth

import "context"

type contextKey string

const userContextKey contextKey = "current_user"

func WithUser(ctx context.Context, user UserPrincipal) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (UserPrincipal, bool) {
	user, ok := ctx.Value(userContextKey).(UserPrincipal)
	return user, ok
}
