package poller

import (
	"getherscan/pkg/models"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgtype"
)

func MakeBlockModel(block *types.Block) (*models.Block, error) {
	blockDifficulty := new(pgtype.Numeric)
	err := blockDifficulty.Set(block.Difficulty().String())
	if err != nil {
		return nil, err
	}

	blockNumber := new(pgtype.Numeric)
	err = blockNumber.Set(block.Number().String())
	if err != nil {
		return nil, err
	}

	blockNonce := new(pgtype.Numeric)
	err = blockNonce.Set(block.Nonce())
	if err != nil {
		return nil, err
	}

	blockBaseFee := new(pgtype.Numeric)
	err = blockBaseFee.Set(block.BaseFee().String())
	if err != nil {
		return nil, err
	}

	return &models.Block{
		Hash:        block.Hash().Bytes(),
		Size:        uint64(block.Size()),
		ParentHash:  block.ParentHash().Bytes(),
		UncleHash:   block.UncleHash().Bytes(),
		Coinbase:    block.Coinbase().Bytes(),
		Root:        block.Root().Bytes(),
		TxHash:      block.TxHash().Bytes(),
		ReceiptHash: block.ReceiptHash().Bytes(),
		Bloom:       block.Bloom().Bytes(),
		Difficulty:  *blockDifficulty,
		Number:      *blockNumber,
		GasLimit:    block.GasLimit(),
		GasUsed:     block.GasUsed(),
		Time:        block.Time(),
		Extra:       block.Extra(),
		MixDigest:   block.MixDigest().Bytes(),
		Nonce:       *blockNonce,
		BaseFee:     *blockBaseFee,
	}, nil
}

func MakeTransactionModel(transaction *types.Transaction, blockHash []byte) (*models.Transaction, error) {
	message, err := transaction.AsMessage(
		types.LatestSignerForChainID(transaction.ChainId()),
		nil,
	)
	if err != nil {
		return nil, err
	}

	transactionChainID := new(pgtype.Numeric)
	err = transactionChainID.Set(transaction.ChainId().String())
	if err != nil {
		return nil, err
	}

	transactionGasPrice := new(pgtype.Numeric)
	err = transactionGasPrice.Set(transaction.GasPrice().String())
	if err != nil {
		return nil, err
	}

	transactionGasTipCap := new(pgtype.Numeric)
	err = transactionGasTipCap.Set(transaction.GasTipCap().String())
	if err != nil {
		return nil, err
	}

	transactionGasFeeCap := new(pgtype.Numeric)
	err = transactionGasFeeCap.Set(transaction.GasFeeCap().String())
	if err != nil {
		return nil, err
	}

	transactionValue := new(pgtype.Numeric)
	err = transactionValue.Set(transaction.Value().String())
	if err != nil {
		return nil, err
	}

	transactionNonce := new(pgtype.Numeric)
	err = transactionNonce.Set(transaction.Nonce())
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		Hash:      transaction.Hash().Bytes(),
		Size:      uint64(transaction.Size()),
		From:      message.From().Bytes(),
		Type:      byte(transaction.Type()),
		ChainID:   *transactionChainID,
		Data:      transaction.Data(),
		Gas:       transaction.Gas(),
		GasPrice:  *transactionGasPrice,
		GasTipCap: *transactionGasTipCap,
		GasFeeCap: *transactionGasFeeCap,
		Value:     *transactionValue,
		Nonce:     *transactionNonce,
		To:        transaction.To().Bytes(),
		BlockHash: blockHash,
	}, nil
}

func MakeBalanceModel(balanceBigInt *big.Int, address, blockHash []byte) (*models.Balance, error) {
	balance := new(pgtype.Numeric)
	err := balance.Set(balanceBigInt.String())
	if err != nil {
		return nil, err
	}

	return &models.Balance{
		Address:   address,
		BlockHash: blockHash,
		Balance:   *balance,
	}, nil
}

func MakeOrphanedBlockModel(block *types.Block) (*models.OrphanedBlock, error) {
	orphanedBlockDifficulty := new(pgtype.Numeric)
	err := orphanedBlockDifficulty.Set(block.Difficulty().String())
	if err != nil {
		return nil, err
	}

	orphanedBlockNumber := new(pgtype.Numeric)
	err = orphanedBlockNumber.Set(block.Number().String())
	if err != nil {
		return nil, err
	}

	orphanedBlockNonce := new(pgtype.Numeric)
	err = orphanedBlockNonce.Set(block.Nonce())
	if err != nil {
		return nil, err
	}

	orphanedBlockBaseFee := new(pgtype.Numeric)
	err = orphanedBlockBaseFee.Set(block.BaseFee().String())
	if err != nil {
		return nil, err
	}

	return &models.OrphanedBlock{
		Hash:        block.Hash().Bytes(),
		Size:        uint64(block.Size()),
		ParentHash:  block.ParentHash().Bytes(),
		UncleHash:   block.UncleHash().Bytes(),
		Coinbase:    block.Coinbase().Bytes(),
		Root:        block.Root().Bytes(),
		TxHash:      block.TxHash().Bytes(),
		ReceiptHash: block.ReceiptHash().Bytes(),
		Bloom:       block.Bloom().Bytes(),
		Difficulty:  *orphanedBlockDifficulty,
		Number:      *orphanedBlockNumber,
		GasLimit:    block.GasLimit(),
		GasUsed:     block.GasUsed(),
		Time:        block.Time(),
		Extra:       block.Extra(),
		MixDigest:   block.MixDigest().Bytes(),
		Nonce:       *orphanedBlockNonce,
		BaseFee:     *orphanedBlockBaseFee,
	}, nil
}

func MakeOrphanedTransactionModel(transaction *types.Transaction, blockHash []byte) (*models.OrphanedTransaction, error) {
	message, err := transaction.AsMessage(
		types.LatestSignerForChainID(transaction.ChainId()),
		nil,
	)
	if err != nil {
		return nil, err
	}

	orphanedTransactionChainID := new(pgtype.Numeric)
	err = orphanedTransactionChainID.Set(transaction.ChainId().String())
	if err != nil {
		return nil, err
	}

	orphanedTransactionGasPrice := new(pgtype.Numeric)
	err = orphanedTransactionGasPrice.Set(transaction.GasPrice().String())
	if err != nil {
		return nil, err
	}

	orphanedTransactionGasTipCap := new(pgtype.Numeric)
	err = orphanedTransactionGasTipCap.Set(transaction.GasTipCap().String())
	if err != nil {
		return nil, err
	}

	orphanedTransactionGasFeeCap := new(pgtype.Numeric)
	err = orphanedTransactionGasFeeCap.Set(transaction.GasFeeCap().String())
	if err != nil {
		return nil, err
	}

	orphanedTransactionValue := new(pgtype.Numeric)
	err = orphanedTransactionValue.Set(transaction.Value().String())
	if err != nil {
		return nil, err
	}

	orphanedTransactionNonce := new(pgtype.Numeric)
	err = orphanedTransactionNonce.Set(transaction.Nonce())
	if err != nil {
		return nil, err
	}

	return &models.OrphanedTransaction{
		Hash:              transaction.Hash().Bytes(),
		Size:              uint64(transaction.Size()),
		From:              message.From().Bytes(),
		Type:              byte(transaction.Type()),
		ChainID:           *orphanedTransactionChainID,
		Data:              transaction.Data(),
		Gas:               transaction.Gas(),
		GasPrice:          *orphanedTransactionGasPrice,
		GasTipCap:         *orphanedTransactionGasTipCap,
		GasFeeCap:         *orphanedTransactionGasFeeCap,
		Value:             *orphanedTransactionValue,
		Nonce:             *orphanedTransactionNonce,
		To:                transaction.To().Bytes(),
		OrphanedBlockHash: blockHash,
	}, nil
}
