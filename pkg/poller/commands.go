package poller

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func PollAction(cliCtx *cli.Context) error {
	var err error

	wsRPCEndpoint := cliCtx.Args().Get(0)
	dbHost := cliCtx.Args().Get(1)
	dbPort := cliCtx.Args().Get(2)
	dbUser := cliCtx.Args().Get(3)
	dbPassword := cliCtx.Args().Get(4)
	dbName := cliCtx.Args().Get(5)
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
	)

	trackedAddresses := []string{}
	if cliCtx.Args().Get(6) != "" {
		trackedAddresses, err = GetTrackedAddressesFromFile(cliCtx.Args().Get(6))
		if err != nil {
			return err
		}
	}

	poller := new(Poller)
	err = poller.Initialize(wsRPCEndpoint, dbConnectionString, trackedAddresses)
	if err != nil {
		return err
	}

	err = poller.DB.ClearDB()
	if err != nil {
		return err
	}

	go poller.Poll()

	log.Println("Listening for new blocks...")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)
	signal.Notify(signalChannel, syscall.SIGINT)

	<-signalChannel

	return nil
}

var PollCommand = cli.Command{
	Name:      "poll",
	Usage:     "Listens for new blocks on the provided websocket RPC endpoint and indexes them to the provided PostgreSQL connection. Optionally accepts a JSON array of hex addresses for which to index balances.",
	ArgsUsage: "Provide a websocket RPC endpoint, a PostgreSQL connection string, and, optionally, a path to a JSON file containing an array of hex addresses to track.",
	Action:    PollAction,
}
