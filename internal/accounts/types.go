package accounts

import "time"

type Account struct {
	Address       string    `json:"address"`
	Type          string    `json:"type"`
	Balance       int64     `json:"balance"`
	TxCount       int64     `json:"tx_count"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

type AccountType string

const (
	AccountTypeTransparent AccountType = "transparent"
	AccountTypeShielded    AccountType = "shielded"
)
