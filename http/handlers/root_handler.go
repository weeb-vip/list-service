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
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/internal/db/repositories/dummy"
	"github.com/weeb-vip/list-service/internal/directives"
	dummy2 "github.com/weeb-vip/list-service/internal/services/dummy"
	"net/http"
)

func BuildRootHandler(conf config.Config) http.Handler {
	database := db.NewDatabase(conf.DBConfig)
	dummyRepository := dummy.NewDummyRepository(database)
	dummyService := dummy2.NewDummyService(dummyRepository)

	resolvers := &graph.Resolver{
		Config:       conf,
		DummyService: dummyService,
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

	return requestinfo.Handler()(logger.Handler()(srv))
}
