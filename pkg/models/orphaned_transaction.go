package models

import "github.com/jackc/pgtype"

type OrphanedTransaction struct {
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
	OrphanedBlockHash string        `json:"orphaned_block_hash" gorm:"primaryKey"`
	OrphanedBlock     OrphanedBlock `json:"orphaned_block" gorm:"foreignKey:OrphanedBlockHash"`
}

func (db *DB) GetOrphanedTransactionsForBlockHash(orphanedBlockHash string) ([]OrphanedTransaction, error) {
	var orphanedTransactions []OrphanedTransaction
	return orphanedTransactions, db.Where("orphaned_block_hash = ?", orphanedBlockHash).Find(&orphanedTransactions).Error
}

func (db *DB) GetOrphanedTransactionsByHash(orphanedTransactionHash string) ([]OrphanedTransaction, error) {
	var orphanedTransactions []OrphanedTransaction
	return orphanedTransactions, db.Preload("OrphanedBlock").Where("hash = ?", orphanedTransactionHash).Find(&orphanedTransactions).Error
}

func (db *DB) GetOrphanedTransactionByHashAndBlockHash(orphanedTransactionHash, orphanedBlockHash string) (*OrphanedTransaction, error) {
	var orphanedTransaction OrphanedTransaction
	return &orphanedTransaction, db.Where("hash = ? AND orphaned_block_hash = ?", orphanedTransactionHash, orphanedBlockHash).First(&orphanedTransaction).Error
}
