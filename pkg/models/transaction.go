package models

import (
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type Transaction struct {
	Hash    string         `json:"hash" gorm:"primaryKey"`
	Size    uint64         `json:"size"`
	From    string         `json:"from"`
	Type    byte           `json:"type"`
	ChainID pgtype.Numeric `json:"chain_id" gorm:"type:numeric"`
	// TODO: Model for access list tuples?
	Data      []byte         `json:"data"`
	Gas       uint64         `json:"gas"`
	GasPrice  pgtype.Numeric `json:"gas_price" gorm:"type:numeric"`
	GasTipCap pgtype.Numeric `json:"gas_tip_cap" gorm:"type:numeric"`
	GasFeeCap pgtype.Numeric `json:"gas_fee_cap" gorm:"type:numeric"`
	Value     pgtype.Numeric `json:"value" gorm:"type:numeric"`
	Nonce     pgtype.Numeric `json:"nonce" gorm:"type:numeric"`
	To        string         `json:"to"`
	// TODO: Figure out how to handle signatures
	BlockHash string `json:"block_hash"`
	Block     Block  `json:"block" gorm:"foreignKey:BlockHash"`
}

func (db *DB) GetTransactionsForBlockHash(blockHash string) ([]Transaction, error) {
	var transactions []Transaction
	return transactions, db.Where("block_hash = ?", blockHash).Find(&transactions).Error
}

func (db *DB) GetTransactionByHash(transactionHash string, includeBlock bool) (*Transaction, error) {
	var transaction Transaction

	if includeBlock {
		return &transaction, db.Preload("Block").Where("hash = ?", transactionHash).First(&transaction).Error
	}

	return &transaction, db.Where("hash = ?", transactionHash).First(&transaction).Error
}

func (db *DB) GetMostExpensiveTransactionForBlockHash(blockHash string) (*Transaction, error) {
	var transaction Transaction
	result := db.Where("block_hash = ?", blockHash).Order("gas*gas_price desc").Limit(1).Find(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &transaction, nil
}
