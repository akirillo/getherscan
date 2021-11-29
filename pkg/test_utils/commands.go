package test_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
)

func SaveBlocksAction(cliCtx *cli.Context) error {
	rpcEndpoint := cliCtx.Args().Get(0)
	ethClient, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return err
	}

	hexBlockHashesFilePath, err := filepath.Abs(cliCtx.Args().Get(1))
	if err != nil {
		return err
	}

	hexBlockHashesFile, err := os.Open(hexBlockHashesFilePath)
	if err != nil {
		return err
	}

	var hexBlockHashes []string
	err = json.NewDecoder(hexBlockHashesFile).Decode(&hexBlockHashes)
	if err != nil {
		return err
	}

	blocks, err := SaveBlocks(ethClient, hexBlockHashes)
	if err != nil {
		return err
	}

	blocksDirPath, err := filepath.Abs(cliCtx.Args().Get(2))
	if err != nil {
		return err
	}

	for i, block := range blocks {
		blockFilePath := fmt.Sprintf("%s/%d.block", blocksDirPath, i)
		if err != nil {
			return err
		}

		blockFile, err := os.Create(blockFilePath)
		if err != nil {
			return err
		}
		defer blockFile.Close()

		err = block.EncodeRLP(blockFile)
		if err != nil {
			return err
		}
	}

	return nil
}

var SaveBlocksCommand = cli.Command{
	Name:      "save_blocks",
	Usage:     "Fetches the blocks whose hashes are in the provided JSON file using the provided RPC endpoint, marshals them, and saves them to the provided path.",
	ArgsUsage: "Provide an RPC endpoint, a path to a JSON file containing an array of hex block hashes to fetch, and a path at which to save the fetched blocks.",
	Action:    SaveBlocksAction,
}
