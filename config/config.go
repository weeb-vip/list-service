package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig     AppConfig `env:"APP_CONFIG"`
	DBConfig      DBConfig
	DataDogConfig DataDogConfig
	PulsarConfig  PulsarConfig
}

type AppConfig struct {
	APPName string `default:"scraper-api"`
	Port    int    `env:"PORT" default:"3000"`
	Version string `default:"x.x.x" env:"VERSION"`
}

type DBConfig struct {
	Host               string `default:"localhost" env:"DBHOST"`
	DataBase           string `default:"weeb" env:"DBNAME"`
	User               string `default:"weeb" env:"DBUSERNAME"`
	Password           string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port               uint   `default:"3306" env:"DBPORT"`
	SSLMode            string `default:"disable" env:"DBSSL"`
	MigrationTableName string `env:"DBMIGRATIONTABLE" default:"__migrations_golang-template"`
}

type DataDogConfig struct {
	DD_AGENT_HOST string `env:"DD_AGENT_HOST" default:"localhost"`
	DD_AGENT_PORT int    `env:"DD_AGENT_PORT" default:"8125"`
}

type PulsarConfig struct {
	URL           string `default:"pulsar://localhost:6650" env:"PULSARURL"`
	ProducerTopic string `default:"public/default/myanimelist.public.user-list" env:"PULSARPRODUCERTOPIC"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
