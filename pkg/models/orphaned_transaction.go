package models

import "github.com/jackc/pgtype"

type OrphanedTransaction struct {
	Hash    [HASH_LENGTH]byte    `json:"hash" gorm:"primaryKey"`
	Size    uint64               `json:"size"`
	From    [ADDRESS_LENGTH]byte `json:"from"`
	Type    byte                 `json:"type"`
	ChainID pgtype.Numeric       `json:"chain_id"`
	// TODO: Model for access list tuples?
	Data      []byte               `json:"data"`
	Gas       uint64               `json:"gas"`
	GasPrice  pgtype.Numeric       `json:"gas_price"`
	GasTipCap pgtype.Numeric       `json:"gas_tip_cap"`
	GasFeeCap pgtype.Numeric       `json:"gas_fee_cap"`
	Value     pgtype.Numeric       `json:"value"`
	Nonce     uint64               `json:"nonce"`
	To        [ADDRESS_LENGTH]byte `json:"to"`
	// TODO: Figure out how to handle signatures
	OrphanedBlockHash [HASH_LENGTH]byte `json:"orphaned_block_hash" gorm:"primaryKey"`
	OrphanedBlock     UncleBlock        `json:"orphaned_block" gorm:"foreignKey:OrphanedBlockHash"`
}

func (db *DB) GetOrphanedTransactionsForBlockHash(orphanedBlockHash []byte) ([]OrphanedTransaction, error) {
	var orphanedTransactions []OrphanedTransaction
	return orphanedTransactions, db.Where("orphaned_block_hash = ?", orphanedBlockHash).Find(&orphanedTransactions).Error
}

func (db *DB) GetOrphanedTransactionsByHash(orphanedTransactionHash []byte) ([]OrphanedTransaction, error) {
	var orphanedTransactions []OrphanedTransaction
	return orphanedTransactions, db.Preload("OrphanedBlock").Where("hash = ?", orphanedTransactionHash).Find(&orphanedTransactions).Error
}