package main

import (
	"getherscan/pkg/api_server"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "API Server"
	app.Commands = []cli.Command{
		api_server.ServeCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
