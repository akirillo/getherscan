package poller

import (
	"bytes"
	"context"
	"errors"
	"getherscan/pkg/models"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type Poller struct {
	DB               *models.DB
	EthClient        *ethclient.Client
	Context          context.Context
	TrackedAddresses [][]byte
}

func (poller *Poller) Initialize(wsRPCEndpoint, dbConnectionString string, trackedAddresses [][]byte) error {
	poller.DB = new(models.DB)
	err := poller.DB.Initialize(dbConnectionString)
	if err != nil {
		return err
	}

	poller.EthClient, err = ethclient.Dial(wsRPCEndpoint)
	if err != nil {
		return err
	}

	poller.Context = context.Background()

	poller.TrackedAddresses = trackedAddresses

	return nil
}

func (poller *Poller) Poll() error {
	headerChannel := make(chan *types.Header)
	subscription, err := poller.EthClient.SubscribeNewHead(poller.Context, headerChannel)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-subscription.Err():
			return err
		case header := <-headerChannel:
			err = poller.Index(header.Hash())
			if err != nil {
				return err
			}
		}
	}
}

func (poller *Poller) Index(blockHash common.Hash) error {
	// Fetch full new block
	block, err := poller.EthClient.BlockByHash(poller.Context, blockHash)
	if err != nil {
		return err
	}

	// Fetch latest indexed block
	head, err := poller.DB.GetHead()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No blocks have been indexed yet
		err = poller.IndexNewBlock(block)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	blockModel, err := MakeBlockModel(block)
	if err != nil {
		return err
	}

	// If new block's parent hash is not that of the current local
	// head, check if reorg is needed
	if bytes.Compare(head.Hash, blockModel.ParentHash) != 0 {
		// Find common ancestor between current local head and
		// new block (assumes we have indexed this far into
		// the past)
		commonAncestorHash, err := poller.FindCommonAncestorHash(head, blockModel)
		if err != nil {
			return err
		}

		currentTotalDifficulty, err := poller.GetTotalDifficultySince(commonAncestorHash, head)
		if err != nil {
			return err
		}

		newTotalDifficulty, err := poller.GetTotalDifficultySince(commonAncestorHash, blockModel)
		if err != nil {
			return err
		}

		if newTotalDifficulty.Cmp(currentTotalDifficulty) > 0 {
			// (Effective) total difficulty of new block
			// is higher than that of local head, reorg

			err = poller.Reorg(block, head, commonAncestorHash)
			if err != nil {
				return err
			}
		} else if newTotalDifficulty.Cmp(currentTotalDifficulty) == 0 && blockModel.Number.Int.Cmp(head.Number.Int) < 0 {
			// (Effective) total difficulties of both new
			// block and local head are equal, but new
			// block has a lower block number (will
			// necessarily have a higher total difficulty
			// once it reaches the same block number),
			// reorg

			err = poller.Reorg(block, head, commonAncestorHash)
			if err != nil {
				return err
			}
		} else {
			// Otherwise, we don't consider this new block canonical

			err = poller.IndexNewOrphanedBlock(block)
			if err != nil {
				return err
			}
		}
	} else {
		err = poller.IndexNewBlock(block)
	}

	return nil
}

func (poller *Poller) IndexNewBlock(block *types.Block) error {
	// Create model for block and write to DB

	blockModel, err := MakeBlockModel(block)
	if err != nil {
		return err
	}

	err = poller.DB.Create(blockModel).Error
	if err != nil {
		return err
	}

	// For each transaction in the block, create a model for it
	// and write to DB

	for _, transaction := range block.Transactions() {
		transactionModel, err := MakeTransactionModel(transaction, blockModel.Hash)
		if err != nil {
			return err
		}

		err = poller.DB.Create(transactionModel).Error
		if err != nil {
			return err
		}
	}

	// For each tracked address, create a model for it and write
	// it to the DB

	err = poller.IndexAddressBalancesForBlock(blockModel.Number.Int, blockModel.Hash)
	if err != nil {
		return err
	}

	log.Printf("Indexed block %s", common.BytesToHash(blockModel.Hash).Hex())

	return nil
}

func (poller *Poller) IndexAddressBalancesForBlock(blockNumber *big.Int, blockHash []byte) error {
	for _, address := range poller.TrackedAddresses {
		balance, err := poller.EthClient.BalanceAt(
			poller.Context,
			common.BytesToAddress(address),
			blockNumber,
		)
		if err != nil {
			return err
		}

		balanceModel, err := MakeBalanceModel(balance, address, blockHash)
		if err != nil {
			return err
		}

		err = poller.DB.Create(balanceModel).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (poller *Poller) IndexNewOrphanedBlock(block *types.Block) error {
	// Create model for block and write to DB

	orphanedBlockModel, err := MakeOrphanedBlockModel(block)
	if err != nil {
		return err
	}

	err = poller.DB.Create(orphanedBlockModel).Error
	if err != nil {
		return err
	}

	// For each transaction in the block, create a model
	// for it and write to DB

	for _, transaction := range block.Transactions() {
		orphanedTransactionModel, err := MakeOrphanedTransactionModel(transaction, orphanedBlockModel.Hash)
		if err != nil {
			return err
		}

		err = poller.DB.Create(orphanedTransactionModel).Error
		if err != nil {
			return err
		}
	}

	log.Printf("Indexed orphaned block %s", common.BytesToHash(orphanedBlockModel.Hash).Hex())

	return nil
}

// Assumes that the common ancestor of blockA and blockB has been previously indexed
func (poller *Poller) FindCommonAncestorHash(blockA, blockB *models.Block) ([]byte, error) {
	var err error

	// Step backwards through each fork until they point to the
	// same ancestor

	for bytes.Compare(blockA.ParentHash, blockB.ParentHash) != 0 {
		err = poller.DB.Where("hash = ?", blockA.ParentHash).First(blockA).Error
		if err != nil {
			return nil, err
		}

		err = poller.DB.Where("hash = ?", blockB.ParentHash).First(blockB).Error
		if err != nil {
			return nil, err
		}
	}

	return blockA.ParentHash, nil
}

func (poller *Poller) GetTotalDifficultySince(ancestorHash []byte, block *models.Block) (*big.Int, error) {
	var err error
	totalDifficulty := big.NewInt(0)

	// Starting with given block, add difficulties up to (but
	// excluding) the block with ancestorHash
	for currentBlock := block; bytes.Compare(currentBlock.Hash, ancestorHash) != 0; {
		totalDifficulty = new(big.Int).Add(totalDifficulty, currentBlock.Difficulty.Int)

		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(currentBlock).Error
		if err != nil {
			return nil, err
		}
	}

	return totalDifficulty, nil
}

func (poller *Poller) Reorg(newHead *types.Block, oldHead *models.Block, commonAncestorHash []byte) error {
	var err error

	// For each block from (and including) oldHead up to (but
	// excluding) the block with commonAncestorHash, orphan the
	// block

	for currentBlock := oldHead; bytes.Compare(currentBlock.Hash, commonAncestorHash) != 0; {
		err = poller.OrphanBlock(currentBlock)
		if err != nil {
			return err
		}

		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(currentBlock).Error
		if err != nil {
			return err
		}
	}

	// Index newHead, then for each block from (but excluding)
	// newHead up to (but excluding) the block with
	// commonAncestorHash, canonicalize the block.

	err = poller.IndexNewBlock(newHead)
	if err != nil {
		return err
	}

	var firstToCanonicalize models.OrphanedBlock
	err = poller.DB.Where("hash = ?", newHead.ParentHash().Bytes()).First(&firstToCanonicalize).Error
	if err != nil {
		return err
	}

	for currentBlock := &firstToCanonicalize; bytes.Compare(currentBlock.Hash, commonAncestorHash) != 0; {
		err = poller.CanonicalizeBlock(currentBlock)
		if err != nil {
			return err
		}

		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(currentBlock).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (poller *Poller) OrphanBlock(block *models.Block) error {
	// Delete transactions associated with block, save temporarily

	transactions, err := poller.DB.GetTransactionsForBlockHash(block.Hash)
	if err != nil {
		return err
	}

	err = poller.DB.Delete(&models.Transaction{}, "block_hash = ?", block.Hash).Error
	if err != nil {
		return err
	}

	// Delete balances associated with block

	err = poller.DB.Delete(&models.Balance{}, "block_hash = ?", block.Hash).Error
	if err != nil {
		return err
	}

	// Delete block

	err = poller.DB.Delete(block).Error
	if err != nil {
		return err
	}

	// Create model for orphaned block

	err = poller.DB.Create(&models.OrphanedBlock{
		Hash:        block.Hash,
		Size:        block.Size,
		ParentHash:  block.ParentHash,
		UncleHash:   block.UncleHash,
		Coinbase:    block.Coinbase,
		Root:        block.Root,
		TxHash:      block.TxHash,
		ReceiptHash: block.ReceiptHash,
		Bloom:       block.Bloom,
		Difficulty:  block.Difficulty,
		Number:      block.Number,
		GasLimit:    block.GasLimit,
		GasUsed:     block.GasUsed,
		Time:        block.Time,
		Extra:       block.Extra,
		MixDigest:   block.MixDigest,
		Nonce:       block.Nonce,
		BaseFee:     block.BaseFee,
	}).Error
	if err != nil {
		return err
	}

	// Create models for orphaned transactions

	for _, transaction := range transactions {
		err = poller.DB.Create(&models.OrphanedTransaction{
			Hash:              transaction.Hash,
			Size:              transaction.Size,
			From:              transaction.From,
			Type:              transaction.Type,
			ChainID:           transaction.ChainID,
			Data:              transaction.Data,
			Gas:               transaction.Gas,
			GasPrice:          transaction.GasPrice,
			GasTipCap:         transaction.GasTipCap,
			GasFeeCap:         transaction.GasFeeCap,
			Value:             transaction.Value,
			Nonce:             transaction.Nonce,
			To:                transaction.To,
			OrphanedBlockHash: transaction.BlockHash,
		}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (poller *Poller) CanonicalizeBlock(orphanedBlock *models.OrphanedBlock) error {
	// Delete orphaned transactions associated with orphaned
	// block, save temporarily

	orphanedTransactions, err := poller.DB.GetOrphanedTransactionsForBlockHash(orphanedBlock.Hash)
	if err != nil {
		return err
	}

	err = poller.DB.Delete(&models.OrphanedTransaction{}, "orphaned_block_hash = ?", orphanedBlock.Hash).Error
	if err != nil {
		return err
	}

	// Delete orphaned block

	err = poller.DB.Delete(orphanedBlock).Error
	if err != nil {
		return err
	}

	// Create model for block

	err = poller.DB.Create(&models.Block{
		Hash:        orphanedBlock.Hash,
		Size:        orphanedBlock.Size,
		ParentHash:  orphanedBlock.ParentHash,
		UncleHash:   orphanedBlock.UncleHash,
		Coinbase:    orphanedBlock.Coinbase,
		Root:        orphanedBlock.Root,
		TxHash:      orphanedBlock.TxHash,
		ReceiptHash: orphanedBlock.ReceiptHash,
		Bloom:       orphanedBlock.Bloom,
		Difficulty:  orphanedBlock.Difficulty,
		Number:      orphanedBlock.Number,
		GasLimit:    orphanedBlock.GasLimit,
		GasUsed:     orphanedBlock.GasUsed,
		Time:        orphanedBlock.Time,
		Extra:       orphanedBlock.Extra,
		MixDigest:   orphanedBlock.MixDigest,
		Nonce:       orphanedBlock.Nonce,
		BaseFee:     orphanedBlock.BaseFee,
	}).Error
	if err != nil {
		return err
	}

	// Create models for transactions

	for _, orphanedTransaction := range orphanedTransactions {
		err = poller.DB.Create(&models.Transaction{
			Hash:      orphanedTransaction.Hash,
			Size:      orphanedTransaction.Size,
			From:      orphanedTransaction.From,
			Type:      orphanedTransaction.Type,
			ChainID:   orphanedTransaction.ChainID,
			Data:      orphanedTransaction.Data,
			Gas:       orphanedTransaction.Gas,
			GasPrice:  orphanedTransaction.GasPrice,
			GasTipCap: orphanedTransaction.GasTipCap,
			GasFeeCap: orphanedTransaction.GasFeeCap,
			Value:     orphanedTransaction.Value,
			Nonce:     orphanedTransaction.Nonce,
			To:        orphanedTransaction.To,
			BlockHash: orphanedTransaction.OrphanedBlockHash,
		}).Error
		if err != nil {
			return err
		}
	}

	// Create models for balances

	err = poller.IndexAddressBalancesForBlock(orphanedBlock.Number.Int, orphanedBlock.Hash)
	if err != nil {
		return err
	}

	return nil
}
