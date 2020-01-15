package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"strings"
)

type Action string

const (
	UP   Action = "up"
	DOWN Action = "down"
)

func (m Action) is(value string) bool {
	return string(m) == strings.ToLower(value)
}

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
			return viper.Unmarshal(&config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.PersistentFlags().StringVarP(&dbFileName, "config", "c", "", "Sets database config name or absolute path")
	cmd.PersistentFlags().StringVarP(&migrationDir, "dir", "d", getEnv("GOMIG_DIR", "migrations"), "Sets the migration location directory")
	cmd.PersistentFlags().BoolVarP(&transaction, "tx", "t", false, "If necessary, you can execute a request in a transaction")
	return cmd
}

func upCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:  "up",
		Long: "Up all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(config.Databases) > 0 {
				if dbConfig, ok := config.Databases[base]; ok {
					return NewGomig(ctx, dbConfig, base).exec(UP)
				}
			}
			for base, dbConfig := range config.Databases {
				return NewGomig(ctx, dbConfig, base).exec(UP)
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "Sets database config name or absolute path")
	return cmd
}

func downCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:  "down",
		Long: "Down all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(base) > 0 {
				if dbConfig, ok := config.Databases[base]; ok {
					return NewGomig(ctx, dbConfig, base).exec(DOWN)
				}
			}
			var err error
			for base, dbConfig := range config.Databases {
				err = multierr.Append(err, NewGomig(ctx, dbConfig, base).exec(DOWN))
			}
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "Sets database config name or absolute path")
	return cmd
}
