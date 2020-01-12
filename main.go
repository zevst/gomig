package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gomig/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	_ = godotenv.Load()
	viper.SetEnvPrefix("GOMIG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigType(getEnv("GOMIG_DB_CONFIG_TYPE", "yaml"))
	viper.AddConfigPath(getEnv("GOMIG_DB_CONFIG_PATH", "config"))
	viper.SetConfigName(getEnv("GOMIG_DB_CONFIG_NAME", "db"))
}

type database struct {
	Logging bool   `mapstructure:"logging"`
	Dialect string `mapstructure:"dialect"`
	Dsn     string `mapstructure:"dsn"`
}

type dbMap map[string]*database

var dbConfig dbMap

func rootCmd(ctx context.Context) *cobra.Command {
	var dbFileName string
	cmd := &cobra.Command{
		Use:  "Gomig",
		Long: "Gomig is a migration tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if len(dbFileName) > 0 {
				viper.SetConfigName(dbFileName)
				if err := viper.MergeInConfig(); err != nil {
					return err
				}
			}
			return viper.Unmarshal(&dbConfig)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println(dbConfig)
			return cmd.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.PersistentFlags().StringVarP(&dbFileName, "db_config", "d", "", "Sets database config name or absolute path")
	return cmd
}

func downCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{}
	return cmd
}

type Gomig struct {
	ctx context.Context
	db  *sql.DB
	err error
}

func NewGomig(ctx context.Context, db *sql.DB) *Gomig {
	return &Gomig{ctx: ctx, db: db}
}

func (g *Gomig) Error() error {
	return g.err
}

func (g *Gomig) exec(filePath string) error {
	tx, err := g.db.BeginTx(g.ctx, nil)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	_, g.err = tx.ExecContext(g.ctx, string(b))
	if g.err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

func upCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:  "up",
		Long: "Up all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println(dbConfig)
			if len(base) > 0 {

			}
			for baseName, config := range dbConfig {
				log.Println(baseName)
				db, err := sql.Open(config.Dialect, config.Dsn)
				if err != nil {
					return err
				}
				err = filepath.Walk("migrations", func(filePath string, f os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if f.IsDir() {
						return nil
					}
					gomig := NewGomig(ctx, db)
					if err := gomig.exec(filePath); err != nil {
						return err
					}
					return gomig.Error()
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "Sets database config name or absolute path")
	return cmd
}

func main() {
	ctx := util.RegisterCloser()
	cmd := rootCmd(ctx)

	cmd.AddCommand(
		upCmd(ctx),
		downCmd(ctx),
	)
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func getEnv(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}
