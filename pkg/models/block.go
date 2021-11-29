package models

import (
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type Block struct {
	Hash string `json:"hash" gorm:"primaryKey"`
	Size uint64 `json:"size"`
	// Header fields
	ParentHash  string         `json:"parent_hash"`
	UncleHash   string         `json:"uncle_hash"`
	Coinbase    string         `json:"coinbase"`
	Root        string         `json:"root"`
	TxHash      string         `json:"tx_hash"`
	ReceiptHash string         `json:"receipt_hash"`
	Bloom       []byte         `json:"bloom"`
	Difficulty  pgtype.Numeric `json:"difficulty" gorm:"type:numeric"`
	Number      pgtype.Numeric `json:"number" gorm:"index:,sort:desc;type:numeric"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Time        uint64         `json:"time"`
	Extra       []byte         `json:"extra"`
	MixDigest   string         `json:"mix_digest"`
	Nonce       pgtype.Numeric `json:"nonce" gorm:"type:numeric"`
	BaseFee     pgtype.Numeric `json:"base_fee" gorm:"type:numeric"`
}

func (db *DB) GetHead() (*Block, error) {
	var head Block
	result := db.Order("number desc").Limit(1).Find(&head)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &head, nil
}

func (db *DB) GetBlockByHash(blockHash string) (*Block, error) {
	var block Block
	return &block, db.Where("hash = ?", blockHash).First(&block).Error
}

func (db *DB) GetBlockByNumber(blockNumber pgtype.Numeric) (*Block, error) {
	var block Block
	return &block, db.Where("number = ?", blockNumber).First(&block).Error
}
