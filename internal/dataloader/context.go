package dataloader

import (
	"context"
	"net/http"
	"github.com/weeb-vip/list-service/internal/services/user_anime"
)

type contextKey string

const (
	userAnimeLoaderKey contextKey = "userAnimeLoader"
)

// Middleware adds dataloaders to the request context
func Middleware(userAnimeService user_anime.UserAnimeServiceImpl) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// Create fresh dataloaders for each request
			userAnimeLoader := NewUserAnimeLoader(userAnimeService)
			ctx = context.WithValue(ctx, userAnimeLoaderKey, userAnimeLoader)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserAnimeLoader retrieves the user anime loader from context
func GetUserAnimeLoader(ctx context.Context) (*UserAnimeLoader, bool) {
	loader, ok := ctx.Value(userAnimeLoaderKey).(*UserAnimeLoader)
	return loader, ok
}