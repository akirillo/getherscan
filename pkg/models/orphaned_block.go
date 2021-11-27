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
	Difficulty  pgtype.Numeric `json:"difficulty"`
	Number      pgtype.Numeric `json:"number"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Time        uint64         `json:"time"`
	Extra       []byte         `json:"extra"`
	MixDigest   []byte         `json:"mix_digest"`
	Nonce       uint64         `json:"nonce"`
	BaseFee     pgtype.Numeric `json:"base_fee"`
}
