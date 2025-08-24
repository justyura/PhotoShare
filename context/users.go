package context

import (
    "context"

    "github.com/z-wentao/PhotoShare/models"
)

type key string

const (
    userKey key = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context{
    return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User{
    val := ctx.Value(userKey)
    // user type assertion here, make sure the value is converted to correct type(*models.User)
    user, ok := val.(*models.User)
    if !ok {
	return nil
    }
    return user
}
