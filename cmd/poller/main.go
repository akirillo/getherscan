package main

import (
	"getherscan/pkg/poller"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Poller"
	app.Commands = []cli.Command{
		poller.PollCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
