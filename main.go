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
	zlog.Info("End migration", zap.Error(cmd.Execute()))
	zlog.End()
}
