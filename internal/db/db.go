package db

import (
	"fmt"
	"github.com/weeb-vip/list-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type DB struct {
	DB *gorm.DB
}

func NewDatabase(cfg config.DBConfig) *DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DataBase, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database connection")
	}

	// The pool is sized to be reused rather than refilled.
	//
	// MaxIdleConns matches MaxOpenConns deliberately: Go only retains up to
	// MaxIdleConns, so anything opened above it is closed again the moment the
	// query finishes. The previous settings kept far fewer idle than they were
	// willing to open, which meant most connections were built per-query --
	// TCP, TLS and auth each time, against RDS over the internet.
	//
	// 5 is small because the database is small. db.t4g.micro allows 79
	// connections in total, and roughly 36 pods share them; a warm pool of 5
	// serves more traffic than a churning pool of 25 while claiming a fraction
	// of that budget.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)

	// Long enough that connections survive quiet periods and get reused, short
	// enough that a failover or DNS change is picked up without a restart.
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Add tracing plugin
	err = db.Use(&TracingPlugin{})
	if err != nil {
		panic("failed to register tracing plugin")
	}

	return &DB{DB: db}
}
