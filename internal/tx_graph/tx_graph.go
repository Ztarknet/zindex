package tx_graph

import (
	"context"
	"fmt"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS tx_edges (
			id SERIAL PRIMARY KEY,
			from_txid VARCHAR(64) NOT NULL,
			to_txid VARCHAR(64) NOT NULL,
			vout INT NOT NULL,
			vin INT NOT NULL,
			value BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_tx_edges_from ON tx_edges(from_txid);
		CREATE INDEX IF NOT EXISTS idx_tx_edges_to ON tx_edges(to_txid);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create tx_graph schema: %w", err)
	}

	return nil
}

func GetTransaction(txid string) (*Transaction, error) {
	tx, err := postgres.PostgresQueryOne[Transaction](
		`SELECT txid, block_height, version, locktime, expiry_height, size, created_at
		 FROM transactions WHERE txid = $1`,
		txid,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func GetTransactionGraph(txid string, depth int) ([]TransactionEdge, error) {
	query := `
		WITH RECURSIVE tx_graph AS (
			SELECT from_txid, to_txid, vout, vin, value, 0 as depth
			FROM tx_edges
			WHERE from_txid = $1

			UNION ALL

			SELECT e.from_txid, e.to_txid, e.vout, e.vin, e.value, g.depth + 1
			FROM tx_edges e
			INNER JOIN tx_graph g ON e.from_txid = g.to_txid
			WHERE g.depth < $2
		)
		SELECT from_txid, to_txid, vout, vin, value FROM tx_graph
	`

	edges, err := postgres.PostgresQuery[TransactionEdge](query, txid, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction graph: %w", err)
	}

	return edges, nil
}
