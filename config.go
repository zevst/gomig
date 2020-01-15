package main

import (
	"database/sql"
	"errors"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var ErrNothing = errors.New("gomig: nothing to do")

func init() {
	_ = godotenv.Load()
	viper.SetEnvPrefix("GOMIG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigType(getEnv("GOMIG_CONFIG_TYPE", "yaml"))
	viper.AddConfigPath(getEnv("GOMIG_CONFIG_PATH", "config"))
	viper.SetConfigName(getEnv("GOMIG_CONFIG_NAME", "db"))
}

type database struct {
	TableName string `mapstructure:"table_name"`
	Dialect   string `mapstructure:"dialect"`
	Dsn       string `mapstructure:"dsn"`
}

func (d *database) Connect() (*sql.DB, error) {
	return sql.Open(d.Dialect, d.Dsn)
}

type Config struct {
	Databases map[string]*database `mapstructure:"db"`
}

var config *Config

var transaction bool
var migrationDir string
var dbMigrationFiles []string

func getEnv(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}
