package http

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"github.com/weeb-vip/golang-template/config"
	"github.com/weeb-vip/golang-template/http/handlers"
	"github.com/weeb-vip/golang-template/metrics"
	"log"
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

func StartServer() error {
	cfg := config.LoadConfigOrPanic()
	router := SetupServer(cfg)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", cfg.AppConfig.Port)

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.Port), router)
}
