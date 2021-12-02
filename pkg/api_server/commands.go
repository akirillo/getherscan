package api_server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func ServeAction(cliCtx *cli.Context) error {
	dbHost := cliCtx.Args().Get(0)
	dbPort := cliCtx.Args().Get(1)
	dbUser := cliCtx.Args().Get(2)
	dbPassword := cliCtx.Args().Get(3)
	dbName := cliCtx.Args().Get(4)
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
	)
	port := cliCtx.Args().Get(5)

	apiServer := new(APIServer)
	err := apiServer.Initialize(dbConnectionString, port)
	if err != nil {
		return err
	}

	go apiServer.Serve()

	log.Printf("Listening on port %s\n", port)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)
	signal.Notify(signalChannel, syscall.SIGINT)

	<-signalChannel

	return nil
}

var ServeCommand = cli.Command{
	Name:      "serve",
	Usage:     "Listens for and serves query requests for the indexer on the provided port, using the provided PostgreSQL connection.",
	ArgsUsage: "Provide a PostgreSQL connection string, and a port number.",
	Action:    ServeAction,
}
