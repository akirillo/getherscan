package api_server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func ServeAction(cliCtx *cli.Context) error {
	dbConnectionString := cliCtx.Args().Get(0)
	port := cliCtx.Args().Get(1)

	apiServer := new(APIServer)
	err := apiServer.Initialize(dbConnectionString)
	if err != nil {
		return err
	}

	go apiServer.Serve(port)

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
