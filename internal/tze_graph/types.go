package tze_graph

import "time"

type TZETransaction struct {
	TxID        string    `json:"txid"`
	BlockHeight int64     `json:"block_height"`
	TZEType     string    `json:"tze_type"`
	Payload     []byte    `json:"payload"`
	WitnessData []byte    `json:"witness_data"`
	CreatedAt   time.Time `json:"created_at"`
}

type TZEWitness struct {
	ID          int64     `json:"id"`
	TxID        string    `json:"txid"`
	WitnessType string    `json:"witness_type"`
	Data        []byte    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
}
