package models

import "github.com/jackc/pgtype"

type OrphanedBlock struct {
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
	Difficulty  pgtype.Numeric `json:"difficulty" gorm:"type:numeric"`
	Number      pgtype.Numeric `json:"number" gorm:"type:numeric"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Time        uint64         `json:"time"`
	Extra       []byte         `json:"extra"`
	MixDigest   []byte         `json:"mix_digest"`
	Nonce       pgtype.Numeric `json:"nonce" gorm:"type:numeric"`
	BaseFee     pgtype.Numeric `json:"base_fee" gorm:"type:numeric"`
}
