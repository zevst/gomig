package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/zevst/gomig/util"
	"github.com/zevst/zlog"
	"go.uber.org/zap"
)

func main() {
	ctx := util.RegisterCloser()
	cmd := rootCmd()
	cmd.AddCommand(
		upCmd(ctx),
		downCmd(ctx),
		applyCmd(ctx),
		createCmd(),
	)
	if err := cmd.Execute(); err != nil {
		zlog.Fatal("Main", zap.Error(err))
	}
}
