package main

import (
	"getherscan/pkg/test_utils"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "TestUtils"
	app.Commands = []cli.Command{
		test_utils.SaveBlocksCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
