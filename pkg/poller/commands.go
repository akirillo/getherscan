package poller

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
)

func PollAction(cliCtx *cli.Context) error {
	var err error

	wsRPCEndpoint := cliCtx.Args().Get(0)
	dbConnectionString := cliCtx.Args().Get(1)

	trackedAddresses := [][]byte{}
	if cliCtx.Args().Get(2) != "" {
		trackedAddressesPath, err := filepath.Abs(cliCtx.Args().Get(2))
		if err != nil {
			return err
		}

		trackedAddressesFile, err := os.Open(trackedAddressesPath)
		if err != nil {
			return err
		}
		defer trackedAddressesFile.Close()

		var trackedAddressHexStrings []string
		err = json.NewDecoder(trackedAddressesFile).Decode(&trackedAddressHexStrings)
		if err != nil {
			return err
		}

		trackedAddresses = make([][]byte, len(trackedAddressHexStrings))
		for i, trackedAddressHexString := range trackedAddressHexStrings {
			if !common.IsHexAddress(trackedAddressHexString) {
				return errors.New("Addresses to track are improperly formatted")
			}

			trackedAddresses[i] = common.FromHex(trackedAddressHexString)
		}
	}

	poller := new(Poller)
	err = poller.Initialize(wsRPCEndpoint, dbConnectionString, trackedAddresses)
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
