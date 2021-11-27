package poller

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
)

func PollAction(cliCtx *cli.Context) error {
	wsRPCEndpoint := cliCtx.Args().Get(0)
	dbConnectionString := cliCtx.Args().Get(1)
	timeout, err := strconv.ParseInt(cliCtx.Args().Get(2), 10, 0)
	if err != nil {
		return err
	}

	trackedAddresses := [][]byte{}
	if cliCtx.Args.Get(3) != "" {
		trackedAddressesPath, err := filepath.Abs(cliCtx.Args().Get(3))
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
	cancelCtx, err := poller.Initialize(wsRPCEndpoint, dbConnectionString, timeout, trackedAddresses)
	if err != nil {
		return err
	}
	defer cancelCtx()

	err = poller.Poll()
	if err != nil {
		return err
	}

	return nil
}

var PollCommand = cli.Command{
	Name:      "poll",
	Usage:     "Listens for new blocks on the provided websocket RPC endpoint and indexes them to the provided PostgreSQL connection, using the provided timeout. Optionally accepts a JSON array of hex addresses for which to index balances.",
	ArgsUsage: "Provide a websocket RPC endpoint, a PostgreSQL connection string, a timeout denoted in seconds, and, optionally, a path to a JSON file containing an array of hex addresses to track.",
	Action:    PollAction,
}
