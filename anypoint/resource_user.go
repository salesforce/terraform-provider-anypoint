package anypoint

import (
	"context"

	"github.com/mulesoft-consulting/cloudhub-client-go/user"
)

/*
 * Returns authentication context (includes authorization header)
 */
func getUserAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, user.ContextAccessToken, pco.access_token)
}
