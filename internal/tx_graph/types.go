package tx_graph

import "time"

// Transaction represents a Zcash transaction with its basic properties
type Transaction struct {
	TxID        string    `json:"txid" db:"txid"`
	BlockHeight int64     `json:"block_height" db:"block_height"`
	BlockHash   string    `json:"block_hash" db:"block_hash"`
	Version     int       `json:"version" db:"version"`
	Locktime    int64     `json:"locktime" db:"locktime"`
	Type        string    `json:"type" db:"type"` // coinbase, tze, t2t, t2z, z2t, z2z
	TotalOutput int64     `json:"total_output" db:"total_output"`
	TotalFee    int64     `json:"total_fee" db:"total_fee"`
	Size        int       `json:"size" db:"size"`
	InputCount  int       `json:"input_count" db:"input_count"`
	OutputCount int       `json:"output_count" db:"output_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TransactionOutput represents an output of a transaction
type TransactionOutput struct {
	TxID          string  `json:"txid" db:"txid"`
	Vout          int     `json:"vout" db:"vout"`
	Value         int64   `json:"value" db:"value"`
	SpentByTxID   *string `json:"spent_by_txid,omitempty" db:"spent_by_txid"`     // nullable
	SpentByVin    *int    `json:"spent_by_vin,omitempty" db:"spent_by_vin"`       // nullable
	SpentAtHeight *int64  `json:"spent_at_height,omitempty" db:"spent_at_height"` // nullable
}

// TransactionInput represents an input of a transaction
type TransactionInput struct {
	TxID     string `json:"txid" db:"txid"`
	Vin      int    `json:"vin" db:"vin"`
	Value    int64  `json:"value" db:"value"`
	PrevTxID string `json:"prev_txid" db:"prev_txid"`
	PrevVout int    `json:"prev_vout" db:"prev_vout"`
	Sequence int64  `json:"sequence" db:"sequence"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TxTypeCoinbase TransactionType = "coinbase"
	TxTypeTZE      TransactionType = "tze" // transparent zcash extension
	TxTypeT2T      TransactionType = "t2t" // transparent to transparent
	TxTypeT2Z      TransactionType = "t2z" // transparent to shielded
	TxTypeZ2T      TransactionType = "z2t" // shielded to transparent
	TxTypeZ2Z      TransactionType = "z2z" // shielded to shielded
)
