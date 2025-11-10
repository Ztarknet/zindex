package blocks

import "time"

// Block represents a block in the blockchain
type Block struct {
	Height     int64     `db:"height" json:"height"`
	Hash       string    `db:"hash" json:"hash"`
	PrevHash   string    `db:"prev_hash" json:"prev_hash"`
	MerkleRoot string    `db:"merkle_root" json:"merkle_root"`
	Timestamp  int64     `db:"timestamp" json:"timestamp"`
	Difficulty string    `db:"difficulty" json:"difficulty"`
	Nonce      string    `db:"nonce" json:"nonce"`
	Version    int       `db:"version" json:"version"`
	TxCount    int       `db:"tx_count" json:"tx_count"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
