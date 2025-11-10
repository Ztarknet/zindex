package accounts

import "time"

// Account represents a blockchain account with its basic properties
type Account struct {
	Address     string    `json:"address" db:"address"`
	Balance     int64     `json:"balance" db:"balance"`
	FirstSeenAt time.Time `json:"first_seen_at" db:"first_seen_at"`
}

// AccountTransaction represents a transaction associated with an account
type AccountTransaction struct {
	Address     string `json:"address" db:"address"`
	TxID        string `json:"txid" db:"txid"`
	BlockHeight int64  `json:"block_height" db:"block_height"`
	Type        string `json:"type" db:"type"` // receive, send
}

// AccountTransactionType represents the direction of a transaction relative to an account
type AccountTransactionType string

const (
	TxTypeReceive AccountTransactionType = "receive" // receiving transaction (funds coming in)
	TxTypeSend    AccountTransactionType = "send"    // sending transaction (funds going out)
)
