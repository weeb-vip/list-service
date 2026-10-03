package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig     AppConfig `env:"APP_CONFIG"`
	DBConfig      DBConfig
	DataDogConfig DataDogConfig
	NatsConfig    NatsConfig
	OutboxConfig  OutboxConfig
}

type AppConfig struct {
	APPName string `default:"list-service"`
	Port    int    `env:"PORT" default:"3000"`
	Version string `default:"x.x.x" env:"VERSION"`
	Env     string `default:"development" env:"ENV"`
}

type DBConfig struct {
	Host               string `default:"localhost" env:"DBHOST"`
	DataBase           string `default:"weeb" env:"DBNAME"`
	User               string `default:"weeb" env:"DBUSERNAME"`
	Password           string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port               uint   `default:"5432" env:"DBPORT"`
	SSLMode            string `default:"require" env:"DBSSL"`
	MigrationTableName string `env:"DBMIGRATIONTABLE" default:"__migrations_list-service"`
}

type DataDogConfig struct {
	DD_AGENT_HOST string `env:"DD_AGENT_HOST" default:"localhost"`
	DD_AGENT_PORT int    `env:"DD_AGENT_PORT" default:"8125"`
}

// NatsConfig is where `relay outbox` publishes. Only the URL: the subject is
// on each outbox row, and the driver makes one stream per subject.
type NatsConfig struct {
	URL string `default:"nats://localhost:4222" env:"NATSURL"`
}

// OutboxConfig tunes the relay (go-outbox-lib). Zero values take the library
// defaults.
type OutboxConfig struct {
	PollIntervalMs int `env:"OUTBOX_POLL_INTERVAL_MS" default:"250"`
	BatchSize      int `env:"OUTBOX_BATCH_SIZE" default:"100"`
	RetentionHours int `env:"OUTBOX_RETENTION_HOURS" default:"168"`
	CleanupMinutes int `env:"OUTBOX_CLEANUP_INTERVAL_MINUTES" default:"60"`
	BacklogSeconds int `env:"OUTBOX_BACKLOG_INTERVAL_SECONDS" default:"10"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	// Try to load config file, but don't fail if it doesn't exist
	// All important config should come from environment variables anyway
	configor.Load(&config)

	return config
}
