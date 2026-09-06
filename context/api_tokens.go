package context

import (
	"context"

	"github.com/z-wentao/PhotoShare/models"
)

const apiTokenKey key = "apiToken"

func WithAPIToken(ctx context.Context, token *models.APIToken) context.Context {
	return context.WithValue(ctx, apiTokenKey, token)
}

func APIToken(ctx context.Context) *models.APIToken {
	val := ctx.Value(apiTokenKey)
	token, ok := val.(*models.APIToken)
	if !ok {
		return nil
	}
	return token
}
