package tx_graph

import "time"

type Transaction struct {
	TxID         string    `json:"txid"`
	BlockHeight  int64     `json:"block_height"`
	Version      int       `json:"version"`
	Locktime     int64     `json:"locktime"`
	ExpiryHeight int64     `json:"expiry_height"`
	Size         int       `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}

type TransactionEdge struct {
	FromTxID string `json:"from_txid"`
	ToTxID   string `json:"to_txid"`
	Vout     int    `json:"vout"`
	Vin      int    `json:"vin"`
	Value    int64  `json:"value"`
}
