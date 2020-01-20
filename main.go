package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/zevst/gomig/util"
	"log"
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
		log.Fatalln(err)
	}
}
