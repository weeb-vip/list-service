package handlers

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/graph"
	"github.com/weeb-vip/list-service/graph/generated"
	"github.com/weeb-vip/list-service/http/handlers/logger"
	"github.com/weeb-vip/list-service/http/handlers/requestinfo"
	"github.com/weeb-vip/list-service/http/middleware"
	"github.com/weeb-vip/list-service/internal/dataloader"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/internal/db/repositories/user_anime"
	"github.com/weeb-vip/list-service/internal/db/repositories/user_list"
	"github.com/weeb-vip/list-service/internal/directives"
	user_anime2 "github.com/weeb-vip/list-service/internal/services/user_anime"
	user_list2 "github.com/weeb-vip/list-service/internal/services/user_list"
	"net/http"
)

func BuildRootHandler(conf config.Config) http.Handler {
	database := db.NewDatabase(conf.DBConfig)
	userListRepository := user_list.NewUserListRepository(database)
	userListService := user_list2.NewUserListService(userListRepository)
	userAnimeRepository := user_anime.NewUserAnimeRepository(database)
	userAnimeService := user_anime2.NewUserAnimeService(userAnimeRepository)

	resolvers := &graph.Resolver{
		Config:           conf,
		UserListService:  userListService,
		UserAnimeService: userAnimeService,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}
	cfg.Directives.Authenticated = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		req := requestinfo.FromContext(ctx)

		if req.UserID == nil {
			// unauthorized
			return nil, fmt.Errorf("Access denied")
		}

		return next(ctx)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	return requestinfo.Handler()(logger.Handler()(dataloader.Middleware(userAnimeService)(srv)))
}

func BuildRootHandlerWithContext(ctx context.Context, conf config.Config) http.Handler {
	database := db.NewDatabase(conf.DBConfig)
	userListRepository := user_list.NewUserListRepository(database)
	userListService := user_list2.NewUserListService(userListRepository)
	userAnimeRepository := user_anime.NewUserAnimeRepository(database)
	userAnimeService := user_anime2.NewUserAnimeService(userAnimeRepository)

	resolvers := &graph.Resolver{
		Config:           conf,
		UserListService:  userListService,
		UserAnimeService: userAnimeService,
		Context:          ctx,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}
	cfg.Directives.Authenticated = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		req := requestinfo.FromContext(ctx)

		if req.UserID == nil {
			// unauthorized
			return nil, fmt.Errorf("Access denied")
		}

		return next(ctx)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	// Add GraphQL tracing extension
	srv.Use(&middleware.GraphQLTracingExtension{})

	return requestinfo.Handler()(logger.Handler()(dataloader.Middleware(userAnimeService)(srv)))
}
