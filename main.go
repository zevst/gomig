package main

import (
	_ "github.com/go-sql-driver/mysql"
	"gomig/util"
	"log"
)

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
