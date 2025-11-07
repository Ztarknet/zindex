package starks

import "time"

type StarkProof struct {
	ID               int64     `json:"id"`
	TxID             string    `json:"txid"`
	BlockHeight      int64     `json:"block_height"`
	ProofData        []byte    `json:"proof_data"`
	VerificationKey  []byte    `json:"verification_key"`
	PublicInputs     []byte    `json:"public_inputs"`
	Verified         bool      `json:"verified"`
	VerificationTime *time.Time `json:"verification_time,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type ProofStats struct {
	TotalProofs      int64   `json:"total_proofs"`
	VerifiedProofs   int64   `json:"verified_proofs"`
	UnverifiedProofs int64   `json:"unverified_proofs"`
	VerificationRate float64 `json:"verification_rate"`
}
