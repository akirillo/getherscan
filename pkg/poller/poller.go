package poller

import (
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
	TrackedAddresses []string
}

func (poller *Poller) Initialize(wsRPCEndpoint, dbConnectionString string, trackedAddresses []string) error {
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

	// If new block's parent hash is not that of the current local
	// head, check if reorg is needed
	if head.Hash != block.ParentHash().Hex() {
		orphanedBlockModel, err := MakeOrphanedBlockModel(block)
		if err != nil {
			return err
		}
		// Find canonical ancestor of the new block (assumes
		// we have indexed this far into the past)
		canonicalAncestorHash, err := poller.FindCanonicalAncestorHash(orphanedBlockModel.ParentHash)
		if err != nil {
			return err
		}

		currentTotalDifficulty, err := poller.GetTotalCanonicalDifficultySince(*canonicalAncestorHash, head)
		if err != nil {
			return err
		}

		newTotalDifficulty, err := poller.GetTotalOrphanedDifficultySince(*canonicalAncestorHash, orphanedBlockModel)
		if err != nil {
			return err
		}

		if newTotalDifficulty.Cmp(currentTotalDifficulty) > 0 {
			// (Effective) total difficulty of new block
			// is higher than that of local head, reorg

			err = poller.Reorg(block, head, *canonicalAncestorHash)
			if err != nil {
				return err
			}
		} else if newTotalDifficulty.Cmp(currentTotalDifficulty) == 0 && orphanedBlockModel.Number.Int.Cmp(head.Number.Int) < 0 {
			// (Effective) total difficulties of both new
			// block and local head are equal, but new
			// block has a lower block number (will
			// necessarily have a higher total difficulty
			// once it reaches the same block number),
			// reorg

			err = poller.Reorg(block, head, *canonicalAncestorHash)
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

	log.Printf("Indexed block %s\n", blockModel.Hash)

	return nil
}

func (poller *Poller) IndexAddressBalancesForBlock(blockNumber *big.Int, blockHash string) error {
	for _, address := range poller.TrackedAddresses {
		balance, err := poller.EthClient.BalanceAt(
			poller.Context,
			common.HexToAddress(address),
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

	log.Printf("Indexed orphaned block %s\n", orphanedBlockModel.Hash)

	return nil
}

// Assumes that the canonical ancestor of the associated orphaned
// block has been previously indexed
func (poller *Poller) FindCanonicalAncestorHash(orphanedBlockParentHash string) (*string, error) {
	var err error

	// Step backwards through orphaned fork until an orphaned
	// parent can't be found

	for err == nil {
		var tempOrphanedBlock models.OrphanedBlock
		err = poller.DB.Where("hash = ?", orphanedBlockParentHash).First(&tempOrphanedBlock).Error
		orphanedBlockParentHash = tempOrphanedBlock.ParentHash
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Could not find orphaned block matching
		// orphanedBlockParentHash, parent block must
		// be canonical
		return &orphanedBlockParentHash, nil
	}

	return nil, err
}

func (poller *Poller) GetTotalCanonicalDifficultySince(ancestorHash string, block *models.Block) (*big.Int, error) {
	var err error
	totalDifficulty := big.NewInt(0)

	// Starting with given block, add difficulties up to (but
	// excluding) the block with ancestorHash
	for currentBlock := block; currentBlock.Hash != ancestorHash; {
		totalDifficulty = new(big.Int).Add(totalDifficulty, currentBlock.Difficulty.Int)

		var tempCurrentBlock models.Block
		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(&tempCurrentBlock).Error
		if err != nil {
			return nil, err
		}
		currentBlock = &tempCurrentBlock
	}

	return totalDifficulty, nil
}

func (poller *Poller) GetTotalOrphanedDifficultySince(ancestorHash string, orphanedBlock *models.OrphanedBlock) (*big.Int, error) {
	var err error
	totalDifficulty := orphanedBlock.Difficulty.Int

	// Starting with the given orphaned block, add difficulties up
	// to (but excluding) the block with ancestorHash
	for currentOrphanedBlock := orphanedBlock; currentOrphanedBlock.ParentHash != ancestorHash; {
		var tempCurrentOrphanedBlock models.OrphanedBlock
		err = poller.DB.Where("hash = ?", currentOrphanedBlock.ParentHash).First(&tempCurrentOrphanedBlock).Error
		if err != nil {
			return nil, err
		}
		currentOrphanedBlock = &tempCurrentOrphanedBlock

		totalDifficulty = new(big.Int).Add(totalDifficulty, currentOrphanedBlock.Difficulty.Int)
	}

	return totalDifficulty, nil
}

func (poller *Poller) Reorg(newHead *types.Block, oldHead *models.Block, commonAncestorHash string) error {
	var err error

	// For each block from (and including) oldHead up to (but
	// excluding) the block with commonAncestorHash, orphan the
	// block

	for currentBlock := oldHead; currentBlock.Hash != commonAncestorHash; {
		err = poller.OrphanBlock(currentBlock)
		if err != nil {
			return err
		}

		var tempCurrentBlock models.Block
		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(&tempCurrentBlock).Error
		if err != nil {
			return err
		}
		currentBlock = &tempCurrentBlock
	}

	// Index newHead, then for each block from (but excluding)
	// newHead up to (but excluding) the block with
	// commonAncestorHash, canonicalize the block.

	err = poller.IndexNewBlock(newHead)
	if err != nil {
		return err
	}

	var firstToCanonicalize models.OrphanedBlock
	err = poller.DB.Where("hash = ?", newHead.ParentHash().Hex()).First(&firstToCanonicalize).Error
	if err != nil {
		return err
	}

	for currentBlock := &firstToCanonicalize; currentBlock.Hash != commonAncestorHash; {
		err = poller.CanonicalizeBlock(currentBlock)
		if err != nil {
			return err
		}

		var tempCurrentBlock models.OrphanedBlock
		err = poller.DB.Where("hash = ?", currentBlock.ParentHash).First(&tempCurrentBlock).Error
		if err != nil {
			return err
		}
		currentBlock = &tempCurrentBlock
	}

	log.Printf(
		"Reorged block %s for block %s/n",
		oldHead.Hash,
		newHead.Hash().Hex(),
	)

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
