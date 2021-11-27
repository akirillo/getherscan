package models

import "github.com/jackc/pgtype"

type Transaction struct {
	Hash    []byte         `json:"hash" gorm:"primaryKey"`
	Size    uint64         `json:"size"`
	From    []byte         `json:"from"`
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
	To        []byte         `json:"to"`
	// TODO: Figure out how to handle signatures
	BlockHash []byte `json:"block_hash"`
	Block     Block  `json:"block" gorm:"foreignKey:BlockHash"`
}

func (db *DB) GetTransactionsForBlockHash(blockHash []byte) ([]Transaction, error) {
	var transactions []Transaction
	return transactions, db.Where("block_hash = ?", blockHash).Find(&transactions).Error
}

func (db *DB) GetTransactionByHash(transactionHash []byte) (Transaction, error) {
	var transaction Transaction
	return transaction, db.Preload("Block").Where("hash = ?", transactionHash).First(&transaction).Error
}
