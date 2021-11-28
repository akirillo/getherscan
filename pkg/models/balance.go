package models

import "github.com/jackc/pgtype"

type Balance struct {
	Address   string `json:"address" gorm:"primaryKey"`
	BlockHash string `json:"block_hash" gorm:"primaryKey"`
	// Not sure if we need this belongs_to relationship
	Block   Block          `json:"block" gorm:"foreignKey:BlockHash"`
	Balance pgtype.Numeric `json:"balance" gorm:"type:numeric"`
}

func (db *DB) GetAddressBalanceByBlockHash(address, blockHash string) (Balance, error) {
	var balance Balance
	return balance, db.Where("address = ? AND block_hash = ?", address, blockHash).First(&balance).Error
}
