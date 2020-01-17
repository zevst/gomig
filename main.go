package main

import (
	_ "github.com/go-sql-driver/mysql"
	"gomig/util"
	"log"
)

func main() {
	ctx := util.RegisterCloser()
	cmd := rootCmd()
	cmd.AddCommand(
		upCmd(ctx),
		downCmd(ctx),
		applyCmd(ctx),
	)
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
