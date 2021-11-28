package models

import "github.com/jackc/pgtype"

type OrphanedBlock struct {
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
	Number      pgtype.Numeric `json:"number" gorm:"type:numeric"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Time        uint64         `json:"time"`
	Extra       []byte         `json:"extra"`
	MixDigest   string         `json:"mix_digest"`
	Nonce       pgtype.Numeric `json:"nonce" gorm:"type:numeric"`
	BaseFee     pgtype.Numeric `json:"base_fee" gorm:"type:numeric"`
}
