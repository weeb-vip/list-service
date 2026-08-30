package dataloader

import (
	"context"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
	"github.com/weeb-vip/list-service/internal/services/user_work"
	"net/http"
)

type contextKey string

const (
	userAnimeLoaderKey contextKey = "userAnimeLoader"
	userWorkLoaderKey  contextKey = "userWorkLoader"
)

// Middleware adds dataloaders to the request context
func Middleware(userAnimeService user_anime.UserAnimeServiceImpl, userWorkService user_work.UserWorkServiceImpl) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Create fresh dataloaders for each request
			userAnimeLoader := NewUserAnimeLoader(userAnimeService)
			ctx = context.WithValue(ctx, userAnimeLoaderKey, userAnimeLoader)

			userWorkLoader := NewUserWorkLoader(userWorkService)
			ctx = context.WithValue(ctx, userWorkLoaderKey, userWorkLoader)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserAnimeLoader retrieves the user anime loader from context
func GetUserAnimeLoader(ctx context.Context) (*UserAnimeLoader, bool) {
	loader, ok := ctx.Value(userAnimeLoaderKey).(*UserAnimeLoader)
	return loader, ok
}

// GetUserWorkLoader retrieves the user work loader from context
func GetUserWorkLoader(ctx context.Context) (*UserWorkLoader, bool) {
	loader, ok := ctx.Value(userWorkLoaderKey).(*UserWorkLoader)
	return loader, ok
}
