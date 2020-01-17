package main

import (
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gomig/driver"
	"gomig/errors"
	"os"
	"strings"
)

var (
	ErrNothing                = errors.New("Nothing to do")
	ErrDatabaseNotFound       = errors.New("Database not found")
	ErrDatabaseNotConfigured  = errors.New("Database not configured")
	ErrDialectNotFound        = errors.New("Database dialect not found. You can add your own dialect using gorm.RegisterDialect")
	ErrFilesNotFound          = errors.New("Files not found")
	ErrUndefinedMigrationType = errors.New("Undefined type of migration file")
)

func init() {
	_ = godotenv.Load()
	viper.SetEnvPrefix("GOMIG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigType(getEnv("GOMIG_CONFIG_TYPE", "yaml"))
	viper.AddConfigPath(getEnv("GOMIG_CONFIG_PATH", "config"))
	viper.SetConfigName(getEnv("GOMIG_CONFIG_NAME", "default"))
}

type database struct {
	Dialect      string   `mapstructure:"dialect"`
	Dsn          string   `mapstructure:"dsn"`
	TableOptions []string `mapstructure:"table_options"`
}

func (d *database) Connect() (conn *gorm.DB, err error) {
	switch d.Dialect {
	case driver.MySQL:
		conn, err = driver.MySQLConn(d.Dsn)
	case driver.PgSQL:
		conn, err = driver.PgSQLConn(d.Dsn)
	default:
		return nil, ErrDialectNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(d.TableOptions) > 0 {
		conn.Set("gorm:table_options", strings.Join(d.TableOptions, " "))
	}
	return conn, conn.AutoMigrate(&Entity{}).Error

}

type Config struct {
	Databases map[string]*database `mapstructure:"db"`
}

var config *Config

func getEnv(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}
