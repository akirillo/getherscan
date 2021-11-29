package test_utils

import (
	"context"
	"errors"
	"fmt"
	"getherscan/pkg/poller"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"gorm.io/gorm"
)

// Fetches the blocks whose hashes are in hexBlockHashes via RPC for
// later RLP-encoding. This is necessary for testing since the RPC
// endpoint we're using doesn't have access to archival state, so we
// can't query deeper than 128 blocks back and need to save blocks
// used in tests locally.
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

// Expects that blocks has the chain's blocks from NEWEST to OLDEST
func AssertCanonicalBlocks(testPoller *poller.Poller, blocks []types.Block) error {
	currentBlockModel, err := testPoller.DB.GetHead()
	if err != nil {
		return err
	}

	// Assert that blocks have been indexed, and in the proper
	// order
	for i, block := range blocks {
		if currentBlockModel.Hash != block.Hash().Hex() {
			return errors.New(fmt.Sprintf("Block at depth %d does not match", i))
		}

		// Assert that the block's transactions have been
		// indexed
		for _, transaction := range block.Transactions() {
			_, err := testPoller.DB.GetTransactionByHash(transaction.Hash().Hex(), false)
			if err != nil {
				return err
			}
		}

		currentBlockModel, err = testPoller.DB.GetBlockByHash(currentBlockModel.ParentHash)
		if err != nil {
			break
		}
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Last block checked (first block indexed) should
		// have no indexed parent
		return nil
	} else {
		return err
	}
}

// Expects that orphanedBlocks has the orphaned chain's blocks from
// NEWEST to OLDEST
func AssertOrphanedBlocks(testPoller *poller.Poller, orphanedBlocks []types.Block) error {
	if len(orphanedBlocks) == 0 {
		// Assert that there are no orphaned blocks
		orphanedBlockModels, err := testPoller.DB.GetAllOrphanedBlocks()
		if err != nil {
			return err
		}

		if len(orphanedBlockModels) > 0 {
			return errors.New(fmt.Sprintf("There are %d orphaned blocks, there should be none", len(orphanedBlockModels)))
		}

		return nil
	}

	currentOrphanedBlockModel, err := testPoller.DB.GetOrphanedBlockByHash(orphanedBlocks[0].Hash().Hex())
	if err != nil {
		return err
	}

	// Assert that blocks have been indexed, and in the proper
	// order
	for i, orphanedBlock := range orphanedBlocks {
		if currentOrphanedBlockModel.Hash != orphanedBlock.Hash().Hex() {
			return errors.New(fmt.Sprintf("Orphaned block at depth %d does not match", i))
		}

		// Assert that the block's transactions have been
		// indexed
		for _, orphanedTransaction := range orphanedBlock.Transactions() {
			_, err := testPoller.DB.GetOrphanedTransactionByHashAndBlockHash(orphanedTransaction.Hash().Hex(), orphanedBlock.Hash().Hex())
			if err != nil {
				return err
			}
		}

		currentOrphanedBlockModel, err = testPoller.DB.GetOrphanedBlockByHash(currentOrphanedBlockModel.ParentHash)
		if err != nil {
			break
		}
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Last block checked (first block indexed) should
		// have no indexed parent
		return nil
	} else {
		return err
	}
}
