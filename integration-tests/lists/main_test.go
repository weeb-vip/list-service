//go:build integration

// Package lists holds the end-to-end tests for list changes and the activity
// events they leave in the outbox. They boot the real GraphQL handler
// in-process over the real database and drive it over HTTP with the headers
// the gateway sets.
package lists

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/http/handlers"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/internal/logger"
)

var (
	server   *httptest.Server
	database *gorm.DB
)

func TestMain(m *testing.M) {
	// Defaults match the CI service container; the Makefile and a developer's
	// shell override them.
	defaults := map[string]string{
		"DBHOST": "localhost", "DBPORT": "5432", "DBUSERNAME": "postgres", "DBPASSWORD": "postgres",
		"DBNAME": "weeb", "DBSSL": "disable", "DBMIGRATIONTABLE": "__migrations_list-service",
	}
	for k, v := range defaults {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
	cfg := config.LoadConfigOrPanic()
	logger.Logger(logger.WithServerName("list-service-test"), logger.WithVersion("test"), logger.WithEnvironment("test"))

	database = db.NewDatabase(cfg.DBConfig).DB
	if err := database.Exec("SELECT 1 FROM outbox_events LIMIT 1").Error; err != nil {
		fmt.Println("outbox_events table missing; run the migrations first:", err)
		os.Exit(1)
	}

	server = httptest.NewServer(handlers.BuildRootHandler(cfg))
	code := m.Run()
	server.Close()
	os.Exit(code)
}
