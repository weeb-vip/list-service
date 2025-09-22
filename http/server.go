package http

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/http/handlers"
	"github.com/weeb-vip/list-service/http/middleware"
	"github.com/weeb-vip/list-service/metrics"
	"net/http"
)

func SetupServer(cfg config.Config) *chi.Mux {

	router := chi.NewRouter()

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://localhost:3000"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	router.Handle("/ui/playground", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", handlers.BuildRootHandler(cfg))
	router.Handle("/healthcheck", handlers.HealthCheckHandler())
	router.Handle("/metrics", metrics.NewPrometheusInstance().Handler())

	return router
}

func SetupServerWithContext(ctx context.Context, cfg config.Config) *chi.Mux {

	router := chi.NewRouter()

	// Add tracing middleware to all routes
	router.Use(middleware.TracingMiddleware())

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://localhost:3000"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	router.Handle("/ui/playground", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", handlers.BuildRootHandlerWithContext(ctx, cfg))
	router.Handle("/healthcheck", handlers.HealthCheckHandler())
	router.Handle("/metrics", metrics.NewPrometheusInstance().Handler())

	return router
}

func StartServer() error {
	cfg := config.LoadConfigOrPanic()
	router := SetupServer(cfg)

	log.Info().
		Int("port", cfg.AppConfig.Port).
		Str("playground_url", fmt.Sprintf("http://localhost:%d/", cfg.AppConfig.Port)).
		Msg("Starting GraphQL server")

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.Port), router)
}

func StartServerWithContext(ctx context.Context) error {
	cfg := config.LoadConfigOrPanic()
	router := SetupServerWithContext(ctx, cfg)

	log.Ctx(ctx).Info().
		Int("port", cfg.AppConfig.Port).
		Str("playground_url", fmt.Sprintf("http://localhost:%d/", cfg.AppConfig.Port)).
		Msg("Starting GraphQL server")

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.Port), router)
}
