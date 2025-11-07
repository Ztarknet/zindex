package tze_graph

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS tze_transactions (
			txid VARCHAR(64) PRIMARY KEY,
			block_height BIGINT NOT NULL,
			tze_type VARCHAR(64) NOT NULL,
			payload BYTEA,
			witness_data BYTEA,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_tze_transactions_block_height ON tze_transactions(block_height);
		CREATE INDEX IF NOT EXISTS idx_tze_transactions_type ON tze_transactions(tze_type);

		CREATE TABLE IF NOT EXISTS tze_witnesses (
			id SERIAL PRIMARY KEY,
			txid VARCHAR(64) NOT NULL REFERENCES tze_transactions(txid),
			witness_type VARCHAR(64) NOT NULL,
			data BYTEA,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_tze_witnesses_txid ON tze_witnesses(txid);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create tze_graph schema: %w", err)
	}

	return nil
}

func GetTZETransaction(txid string) (*TZETransaction, error) {
	tx, err := postgres.PostgresQueryOne[TZETransaction](
		`SELECT txid, block_height, tze_type, payload, witness_data, created_at
		 FROM tze_transactions WHERE txid = $1`,
		txid,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get TZE transaction: %w", err)
	}

	return tx, nil
}

func GetTZETransactionsByType(tzeType string, limit, offset int) ([]TZETransaction, error) {
	txs, err := postgres.PostgresQuery[TZETransaction](
		`SELECT txid, block_height, tze_type, payload, witness_data, created_at
		 FROM tze_transactions WHERE tze_type = $1
		 ORDER BY block_height DESC LIMIT $2 OFFSET $3`,
		tzeType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query TZE transactions: %w", err)
	}

	return txs, nil
}

func GetTZEWitnesses(txid string) ([]TZEWitness, error) {
	witnesses, err := postgres.PostgresQuery[TZEWitness](
		`SELECT id, txid, witness_type, data, created_at
		 FROM tze_witnesses WHERE txid = $1`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query TZE witnesses: %w", err)
	}

	return witnesses, nil
}
