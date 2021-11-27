package models

import "github.com/jackc/pgtype"

type Block struct {
	Hash []byte `json:"hash" gorm:"primaryKey"`
	Size uint64 `json:"size"`
	// Header fields
	ParentHash  []byte         `json:"parent_hash"`
	UncleHash   []byte         `json:"uncle_hash"`
	Coinbase    []byte         `json:"coinbase"`
	Root        []byte         `json:"root"`
	TxHash      []byte         `json:"tx_hash"`
	ReceiptHash []byte         `json:"receipt_hash"`
	Bloom       []byte         `json:"bloom"`
	Difficulty  pgtype.Numeric `json:"difficulty"`
	Number      pgtype.Numeric `json:"number" gorm:"index:,sort:desc"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Time        uint64         `json:"time"`
	Extra       []byte         `json:"extra"`
	MixDigest   []byte         `json:"mix_digest"`
	Nonce       uint64         `json:"nonce"`
	BaseFee     pgtype.Numeric `json:"base_fee"`
}

func (db *DB) GetHead() (*Block, error) {
	var head Block
	err := db.Order("number desc").Limit(1).Find(&head).Error
	if err != nil {
		return nil, err
	}

	return &head, nil
}

func (db *DB) GetBlockByHash(blockHash []byte) (Block, error) {
	var block Block
	return block, db.Where("hash = ?", blockHash).First(&block).Error
}

func (db *DB) GetBlockByNumber(blockNumber pgtype.Numeric) (Block, error) {
	var block Block
	return block, db.Where("number = ?", blockNumber).First(&block).Error
}
