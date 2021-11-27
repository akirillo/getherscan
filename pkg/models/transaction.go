package models

import "github.com/jackc/pgtype"

type Transaction struct {
	Hash    []byte         `json:"hash" gorm:"primaryKey"`
	Size    uint64         `json:"size"`
	From    []byte         `json:"from"`
	Type    byte           `json:"type"`
	ChainID pgtype.Numeric `json:"chain_id"`
	// TODO: Model for access list tuples?
	Data      []byte         `json:"data"`
	Gas       uint64         `json:"gas"`
	GasPrice  pgtype.Numeric `json:"gas_price"`
	GasTipCap pgtype.Numeric `json:"gas_tip_cap"`
	GasFeeCap pgtype.Numeric `json:"gas_fee_cap"`
	Value     pgtype.Numeric `json:"value"`
	Nonce     uint64         `json:"nonce"`
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
