package starks

import "time"

// Verifier represents a STARK proof verifier
type Verifier struct {
	VerifierID       string    `json:"verifier_id" db:"verifier_id"`
	VerifierName     string    `json:"verifier_name" db:"verifier_name"`
	VerifierMetadata string    `json:"verifier_metadata" db:"verifier_metadata"`
	Balance          int64     `json:"balance" db:"balance"`
	FirstSeenAt      time.Time `json:"first_seen_at" db:"first_seen_at"`
}

// StarkProof represents a STARK proof associated with a transaction
type StarkProof struct {
	VerifierID  string `json:"verifier_id" db:"verifier_id"`
	TxID        string `json:"txid" db:"txid"`
	BlockHeight int64  `json:"block_height" db:"block_height"`
	ProofSize   int64  `json:"proof_size" db:"proof_size"`
}

// ZtarknetFacts represents Ztarknet-specific facts from STARK proofs
type ZtarknetFacts struct {
	VerifierID       string `json:"verifier_id" db:"verifier_id"`
	TxID             string `json:"txid" db:"txid"`
	BlockHeight      int64  `json:"block_height" db:"block_height"`
	ProofSize        int64  `json:"proof_size" db:"proof_size"`
	OldState         string `json:"old_state" db:"old_state"`
	NewState         string `json:"new_state" db:"new_state"`
	ProgramHash      string `json:"program_hash" db:"program_hash"`
	InnerProgramHash string `json:"inner_program_hash" db:"inner_program_hash"`
}
