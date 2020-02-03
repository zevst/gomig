package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zevst/zlog"
	"go.uber.org/multierr"
	"go.uber.org/zap"
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
	var configFilePath string
	cmd := &cobra.Command{
		Use:  "Gomig",
		Long: "Gomig is a migration tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if len(configFilePath) > 0 {
				viper.SetConfigFile(configFilePath)
				if err := viper.MergeInConfig(); err != nil {
					return err
				}
			}
			if err := viper.Unmarshal(&config); err != nil {
				return err
			}
			if config.Loggers != nil {
				zlog.Start(config.Loggers.Core(), zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPostRun: func(*cobra.Command, []string) {
			zlog.End()
		},
	}
	cmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", getEnv("GOMIG_CONFIG_FILE_PATH", ""), "config file path")
	cmd.PersistentFlags().StringVarP(&migrationDir, "dir", "d", getEnv("GOMIG_DIR", "migrations"), "directory with migrations")
	return cmd
}

func upCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply all up migrations",
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
	cmd.Flags().StringVarP(&base, "base", "b", "", "database name in the config")
	return cmd
}

func downCmd(ctx context.Context) *cobra.Command {
	var base string
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Apply all down migrations",
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
		Use:   "apply",
		Short: "Apply migration from file",
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
	cmd.Flags().StringVarP(&base, "base", "b", "", "database name in the config")
	cmd.Flags().StringVarP(&file, "file", "f", "", "migration file path")
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func createCmd() *cobra.Command {
	var base, name, out string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create migration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return create(base, name, out)
		},
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "database name in the config")
	cmd.Flags().StringVarP(&name, "name", "n", "", "migration name")
	cmd.Flags().StringVarP(&out, "out", "o", getEnv("GOMIG_DIR", "migrations"), "out directory")
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
