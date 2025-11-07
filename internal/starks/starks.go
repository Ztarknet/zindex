package starks

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create starks schema: %w", err)
	}

	return nil
}

func GetProof(id int64) (*StarkProof, error) {
	proof, err := postgres.PostgresQueryOne[StarkProof](
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE id = $1`,
		id,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	return proof, nil
}

func GetProofsByTransaction(txid string) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE txid = $1`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query proofs: %w", err)
	}

	return proofs, nil
}

func GetProofStats() (*ProofStats, error) {
	stats, err := postgres.PostgresQueryOne[ProofStats](
		`SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE verified = TRUE) as verified,
			COUNT(*) FILTER (WHERE verified = FALSE) as unverified
		 FROM stark_proofs`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get proof stats: %w", err)
	}

	if stats.TotalProofs > 0 {
		stats.VerificationRate = float64(stats.VerifiedProofs) / float64(stats.TotalProofs)
	}

	return stats, nil
}

func GetUnverifiedProofs(limit int) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT id, txid, block_height, proof_data, verification_key, public_inputs, verified, verification_time, created_at
		 FROM stark_proofs WHERE verified = FALSE
		 ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query unverified proofs: %w", err)
	}

	return proofs, nil
}
