package tx_graph

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

// InitSchema creates the transaction graph tables and indexes
func InitSchema() error {
	schema := `
		-- Transactions table
		CREATE TABLE IF NOT EXISTS transactions (
			txid VARCHAR(64) PRIMARY KEY,
			block_height BIGINT NOT NULL,
			block_hash VARCHAR(64) NOT NULL,
			version INT NOT NULL,
			locktime BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL,
			total_output BIGINT NOT NULL,
			total_fee BIGINT NOT NULL,
			size INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Transaction outputs table
		CREATE TABLE IF NOT EXISTS transaction_outputs (
			txid VARCHAR(64) NOT NULL,
			vout INT NOT NULL,
			value BIGINT NOT NULL,
			spent_by_txid VARCHAR(64),
			spent_by_vin INT,
			spent_at_height BIGINT,
			PRIMARY KEY (txid, vout),
			FOREIGN KEY (txid) REFERENCES transactions(txid) ON DELETE CASCADE
		);

		-- Transaction inputs table
		CREATE TABLE IF NOT EXISTS transaction_inputs (
			txid VARCHAR(64) NOT NULL,
			vin INT NOT NULL,
			value BIGINT NOT NULL,
			prev_txid VARCHAR(64) NOT NULL,
			prev_vout INT NOT NULL,
			sequence BIGINT NOT NULL,
			PRIMARY KEY (txid, vin),
			FOREIGN KEY (txid) REFERENCES transactions(txid) ON DELETE CASCADE
		);

		-- Indexes for transactions
		CREATE INDEX IF NOT EXISTS idx_transactions_block_height ON transactions(block_height);
		CREATE INDEX IF NOT EXISTS idx_transactions_block_hash ON transactions(block_hash);
		CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
		CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

		-- Indexes for transaction outputs
		CREATE INDEX IF NOT EXISTS idx_tx_outputs_txid ON transaction_outputs(txid);
		CREATE INDEX IF NOT EXISTS idx_tx_outputs_spent_by ON transaction_outputs(spent_by_txid) WHERE spent_by_txid IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_tx_outputs_unspent ON transaction_outputs(txid, vout) WHERE spent_by_txid IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tx_outputs_value ON transaction_outputs(value);

		-- Indexes for transaction inputs
		CREATE INDEX IF NOT EXISTS idx_tx_inputs_txid ON transaction_inputs(txid);
		CREATE INDEX IF NOT EXISTS idx_tx_inputs_prev ON transaction_inputs(prev_txid, prev_vout);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create tx_graph schema: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by its txid
func GetTransaction(txid string) (*Transaction, error) {
	tx, err := postgres.PostgresQueryOne[Transaction](
		`SELECT txid, block_height, block_hash, version, locktime, type,
		        total_output, total_fee, size, created_at
		 FROM transactions WHERE txid = $1`,
		txid,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

// GetTransactionsByBlock retrieves all transactions in a block
func GetTransactionsByBlock(blockHeight int64) ([]Transaction, error) {
	txs, err := postgres.PostgresQuery[Transaction](
		`SELECT txid, block_height, block_hash, version, locktime, type,
		        total_output, total_fee, size, created_at
		 FROM transactions WHERE block_height = $1
		 ORDER BY txid`,
		blockHeight,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by block: %w", err)
	}

	return txs, nil
}

// GetTransactionsByType retrieves transactions by type with pagination
func GetTransactionsByType(txType string, limit, offset int) ([]Transaction, error) {
	txs, err := postgres.PostgresQuery[Transaction](
		`SELECT txid, block_height, block_hash, version, locktime, type,
		        total_output, total_fee, size, created_at
		 FROM transactions WHERE type = $1
		 ORDER BY block_height DESC, txid
		 LIMIT $2 OFFSET $3`,
		txType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by type: %w", err)
	}

	return txs, nil
}

// GetRecentTransactions retrieves the most recent transactions
func GetRecentTransactions(limit, offset int) ([]Transaction, error) {
	txs, err := postgres.PostgresQuery[Transaction](
		`SELECT txid, block_height, block_hash, version, locktime, type,
		        total_output, total_fee, size, created_at
		 FROM transactions
		 ORDER BY block_height DESC, created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	return txs, nil
}

// GetTransactionOutputs retrieves all outputs for a transaction
func GetTransactionOutputs(txid string) ([]TransactionOutput, error) {
	outputs, err := postgres.PostgresQuery[TransactionOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height
		 FROM transaction_outputs
		 WHERE txid = $1
		 ORDER BY vout`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction outputs: %w", err)
	}

	return outputs, nil
}

// GetTransactionOutput retrieves a specific output
func GetTransactionOutput(txid string, vout int) (*TransactionOutput, error) {
	output, err := postgres.PostgresQueryOne[TransactionOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height
		 FROM transaction_outputs
		 WHERE txid = $1 AND vout = $2`,
		txid, vout,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction output: %w", err)
	}

	return output, nil
}

// GetUnspentOutputs retrieves all unspent outputs for a transaction
func GetUnspentOutputs(txid string) ([]TransactionOutput, error) {
	outputs, err := postgres.PostgresQuery[TransactionOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height
		 FROM transaction_outputs
		 WHERE txid = $1 AND spent_by_txid IS NULL
		 ORDER BY vout`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get unspent outputs: %w", err)
	}

	return outputs, nil
}

// GetTransactionInputs retrieves all inputs for a transaction
func GetTransactionInputs(txid string) ([]TransactionInput, error) {
	inputs, err := postgres.PostgresQuery[TransactionInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, sequence
		 FROM transaction_inputs
		 WHERE txid = $1
		 ORDER BY vin`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction inputs: %w", err)
	}

	return inputs, nil
}

// GetTransactionInput retrieves a specific input
func GetTransactionInput(txid string, vin int) (*TransactionInput, error) {
	input, err := postgres.PostgresQueryOne[TransactionInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, sequence
		 FROM transaction_inputs
		 WHERE txid = $1 AND vin = $2`,
		txid, vin,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction input: %w", err)
	}

	return input, nil
}

// GetOutputSpenders retrieves all transactions that spent outputs from a given transaction
func GetOutputSpenders(txid string) ([]TransactionOutput, error) {
	outputs, err := postgres.PostgresQuery[TransactionOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height
		 FROM transaction_outputs
		 WHERE txid = $1 AND spent_by_txid IS NOT NULL
		 ORDER BY vout`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get output spenders: %w", err)
	}

	return outputs, nil
}

// GetInputSources retrieves all transactions that provided inputs to a given transaction
func GetInputSources(txid string) ([]TransactionInput, error) {
	inputs, err := postgres.PostgresQuery[TransactionInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, sequence
		 FROM transaction_inputs
		 WHERE txid = $1
		 ORDER BY vin`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get input sources: %w", err)
	}

	return inputs, nil
}

// GetTransactionGraph builds a graph of connected transactions
// Returns transactions that are connected through inputs/outputs
func GetTransactionGraph(txid string, depth int) ([]string, error) {
	query := `
		WITH RECURSIVE tx_graph AS (
			-- Start with the given transaction
			SELECT $1::VARCHAR AS txid, 0 AS depth

			UNION

			-- Find transactions that spent outputs from current level
			SELECT DISTINCT o.spent_by_txid AS txid, g.depth + 1 AS depth
			FROM tx_graph g
			JOIN transaction_outputs o ON o.txid = g.txid
			WHERE o.spent_by_txid IS NOT NULL AND g.depth < $2

			UNION

			-- Find transactions that provided inputs to current level
			SELECT DISTINCT i.prev_txid AS txid, g.depth + 1 AS depth
			FROM tx_graph g
			JOIN transaction_inputs i ON i.txid = g.txid
			WHERE g.depth < $2
		)
		SELECT DISTINCT txid FROM tx_graph WHERE txid IS NOT NULL ORDER BY txid
	`

	type result struct {
		TxID string `db:"txid"`
	}

	results, err := postgres.PostgresQuery[result](query, txid, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction graph: %w", err)
	}

	txids := make([]string, len(results))
	for i, r := range results {
		txids[i] = r.TxID
	}

	return txids, nil
}
