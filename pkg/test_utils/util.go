package test_utils

import (
	"context"
	"fmt"
	"getherscan/pkg/poller"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

// Fetches the blocks whose hashes are in hexBlockHashes via RPC for
// later RLP-encoding. This is necessary for testing since the RPC
// endpoint we're using doesn't have access to archival state, so we
// can't query more than 128 blocks in the past and need to save
// blocks used in tests locally.
func SaveBlocks(ethClient *ethclient.Client, hexBlockHashes []string) ([]types.Block, error) {
	blocks := make([]types.Block, len(hexBlockHashes))
	for i, hexBlockHash := range hexBlockHashes {
		blockHash := common.HexToHash(hexBlockHash)
		block, err := ethClient.BlockByHash(context.Background(), blockHash)
		if err != nil {
			return nil, err
		}

		blocks[i] = *block
	}

	return blocks, nil
}

// Reads RLP-encoded blocks from the given directory path
func GetBlocksFromDir(blocksDirPath string) ([]types.Block, error) {
	absBlocksDirPath, err := filepath.Abs(blocksDirPath)
	if err != nil {
		return nil, err
	}

	blocksDirEntries, err := os.ReadDir(absBlocksDirPath)
	if err != nil {
		return nil, err
	}

	blocks := make([]types.Block, len(blocksDirEntries))
	for i, blocksDirEntry := range blocksDirEntries {
		// Assumes that every entry in the provided directory
		// is a file containing an RLP-encoded block
		blockFilePath := fmt.Sprintf("%s/%s", blocksDirPath, blocksDirEntry.Name())
		blockFile, err := os.Open(blockFilePath)
		if err != nil {
			return nil, err
		}
		defer blockFile.Close()

		var block types.Block
		rlpStream := rlp.NewStream(blockFile, 0)
		err = block.DecodeRLP(rlpStream)
		if err != nil {
			return nil, err
		}

		blocks[i] = block
	}

	return blocks, nil
}

// Reads in a file containing a JSON array of block hashes, and
// indexes the blocks in the order in which their hashes appear in the
// array
func TestPoll(testPoller *poller.Poller, blocks []types.Block) error {
	var err error

	for _, block := range blocks {
		err = testPoller.Index(&block)
		if err != nil {
			return err
		}
	}

	return nil
}
