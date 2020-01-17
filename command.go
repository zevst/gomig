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

var migrationDir string

func (m Action) is(value string) bool {
	return string(m) == strings.ToLower(value)
}

func rootCmd() *cobra.Command {
	var dbFileName string
	cmd := &cobra.Command{
		Use:  "Gomig",
		Long: "Gomig is a migration tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if len(dbFileName) > 0 {
				viper.SetConfigFile(dbFileName)
				if err := viper.MergeInConfig(); err != nil {
					return err
				}
			}
			return viper.Unmarshal(&config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().StringVarP(&dbFileName, "config", "c", "", "database config name or absolute path")
	cmd.PersistentFlags().StringVarP(&migrationDir, "dir", "d", getEnv("GOMIG_DIR", "migrations"), "Sets the migration location directory")
	return cmd
}

func upCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:  "up",
		Long: "Up all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(config.Databases) == 0 {
				return ErrDatabaseNotConfigured
			}
			if len(base) > 0 {
				if dbConfig, ok := config.Databases[base]; ok {
					return up(ctx, dbConfig)
				}
				return ErrDatabaseNotFound
			}
			var err error
			for _, dbConfig := range config.Databases {
				err = multierr.Append(err, up(ctx, dbConfig))
			}
			return err
		},
		SilenceErrors: true,
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
			if len(config.Databases) == 0 {
				return ErrDatabaseNotConfigured
			}
			if len(base) > 0 {
				if dbConfig, ok := config.Databases[base]; ok {
					return down(ctx, dbConfig)
				}
				return ErrDatabaseNotFound
			}
			var err error
			for _, dbConfig := range config.Databases {
				err = multierr.Append(err, down(ctx, dbConfig))
			}
			return err
		},
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "database name in the config")
	return cmd
}

func applyCmd(ctx context.Context) *cobra.Command {
	var base, file string
	cmd := &cobra.Command{
		Use:  "apply",
		Long: "Apply migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(config.Databases) == 0 {
				return ErrDatabaseNotConfigured
			}
			if len(base) > 0 {
				if dbConfig, ok := config.Databases[base]; ok {
					return apply(ctx, dbConfig, file)
				}
			}
			return ErrDatabaseNotFound
		},
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "Sets database config name or absolute path")
	cmd.Flags().StringVarP(&file, "file", "f", "", "migration file path")
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
