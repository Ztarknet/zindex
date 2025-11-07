package starks

import (
	"database/sql"
	"fmt"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS stark_proofs (
			id SERIAL PRIMARY KEY,
			txid VARCHAR(64) NOT NULL,
			block_height BIGINT NOT NULL,
			proof_data BYTEA NOT NULL,
			verification_key BYTEA,
			public_inputs BYTEA,
			verified BOOLEAN DEFAULT FALSE,
			verification_time TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_stark_proofs_txid ON stark_proofs(txid);
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_block_height ON stark_proofs(block_height);
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_verified ON stark_proofs(verified);
	`

	_, err := postgres.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create starks schema: %w", err)
	}

	return nil
}

func GetProof(id int64) (*StarkProof, error) {
	var proof StarkProof
	err := postgres.DB.QueryRow(
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE id = $1`,
		id,
	).Scan(&proof.ID, &proof.TxID, &proof.BlockHeight, &proof.ProofData, &proof.VerificationKey,
		&proof.PublicInputs, &proof.Verified, &proof.VerificationTime, &proof.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	return &proof, nil
}

func GetProofsByTransaction(txid string) ([]StarkProof, error) {
	rows, err := postgres.DB.Query(
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE txid = $1`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query proofs: %w", err)
	}
	defer rows.Close()

	var proofs []StarkProof
	for rows.Next() {
		var proof StarkProof
		if err := rows.Scan(&proof.ID, &proof.TxID, &proof.BlockHeight, &proof.ProofData,
			&proof.VerificationKey, &proof.PublicInputs, &proof.Verified,
			&proof.VerificationTime, &proof.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan proof: %w", err)
		}
		proofs = append(proofs, proof)
	}

	return proofs, nil
}

func GetProofStats() (*ProofStats, error) {
	var stats ProofStats
	err := postgres.DB.QueryRow(
		`SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE verified = TRUE) as verified,
			COUNT(*) FILTER (WHERE verified = FALSE) as unverified
		 FROM stark_proofs`,
	).Scan(&stats.TotalProofs, &stats.VerifiedProofs, &stats.UnverifiedProofs)

	if err != nil {
		return nil, fmt.Errorf("failed to get proof stats: %w", err)
	}

	if stats.TotalProofs > 0 {
		stats.VerificationRate = float64(stats.VerifiedProofs) / float64(stats.TotalProofs)
	}

	return &stats, nil
}

func GetUnverifiedProofs(limit int) ([]StarkProof, error) {
	rows, err := postgres.DB.Query(
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE verified = FALSE
		 ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query unverified proofs: %w", err)
	}
	defer rows.Close()

	var proofs []StarkProof
	for rows.Next() {
		var proof StarkProof
		if err := rows.Scan(&proof.ID, &proof.TxID, &proof.BlockHeight, &proof.ProofData,
			&proof.VerificationKey, &proof.PublicInputs, &proof.Verified,
			&proof.VerificationTime, &proof.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan proof: %w", err)
		}
		proofs = append(proofs, proof)
	}

	return proofs, nil
}
