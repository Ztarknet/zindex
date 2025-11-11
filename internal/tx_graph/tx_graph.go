package tx_graph

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

// DBTX is an interface that both pgxpool.Pool and pgx.Tx implement
// This allows functions to work with either a connection pool or a transaction
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func init() {
	// Register this module's schema initialization with the postgres package
	postgres.RegisterModuleSchema("TX_GRAPH", InitSchema)
}

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
			-- Non-recursive term: Start with the given transaction
			SELECT $1::VARCHAR AS txid, 0 AS depth

			UNION

			-- Recursive term: Find both spenders and sources in one unified query
			SELECT DISTINCT connected_tx AS txid, g.depth + 1 AS depth
			FROM tx_graph g
			CROSS JOIN LATERAL (
				-- Transactions that spent outputs from current level (forward traversal)
				SELECT o.spent_by_txid AS connected_tx
				FROM transaction_outputs o
				WHERE o.txid = g.txid AND o.spent_by_txid IS NOT NULL

				UNION

				-- Transactions that provided inputs to current level (backward traversal)
				SELECT i.prev_txid AS connected_tx
				FROM transaction_inputs i
				WHERE i.txid = g.txid
			) AS connections
			WHERE g.depth < $2 AND connected_tx IS NOT NULL
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

// StoreTransaction inserts or updates a transaction in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreTransaction(postgresTx DBTX, txid string, blockHeight int64, blockHash string, version int, locktime int64, txType string, totalOutput int64, totalFee int64, size int) error {
	ctx := context.Background()

	query := `
		INSERT INTO transactions (txid, block_height, block_hash, version, locktime, type, total_output, total_fee, size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (txid) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			block_hash = EXCLUDED.block_hash,
			version = EXCLUDED.version,
			locktime = EXCLUDED.locktime,
			type = EXCLUDED.type,
			total_output = EXCLUDED.total_output,
			total_fee = EXCLUDED.total_fee,
			size = EXCLUDED.size
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, txid, blockHeight, blockHash, version, locktime, txType, totalOutput, totalFee, size)
	if err != nil {
		return fmt.Errorf("failed to store transaction %s: %w", txid, err)
	}

	return nil
}

// StoreTransactionOutput inserts or updates a transaction output in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreTransactionOutput(postgresTx DBTX, txid string, vout int, value int64) error {
	ctx := context.Background()

	query := `
		INSERT INTO transaction_outputs (txid, vout, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (txid, vout) DO UPDATE SET
			value = EXCLUDED.value
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, txid, vout, value)
	if err != nil {
		return fmt.Errorf("failed to store transaction output %s:%d: %w", txid, vout, err)
	}

	return nil
}

// StoreTransactionInput inserts or updates a transaction input in the database
// and marks the corresponding output as spent
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreTransactionInput(postgresTx DBTX, txid string, vin int, value int64, prevTxid string, prevVout int, sequence int64, blockHeight int64) error {
	ctx := context.Background()

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	// Insert the input
	inputQuery := `
		INSERT INTO transaction_inputs (txid, vin, value, prev_txid, prev_vout, sequence)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (txid, vin) DO UPDATE SET
			value = EXCLUDED.value,
			prev_txid = EXCLUDED.prev_txid,
			prev_vout = EXCLUDED.prev_vout,
			sequence = EXCLUDED.sequence
	`

	_, err := postgresTx.Exec(ctx, inputQuery, txid, vin, value, prevTxid, prevVout, sequence)
	if err != nil {
		return fmt.Errorf("failed to store transaction input %s:%d: %w", txid, vin, err)
	}

	// Mark the previous output as spent
	outputQuery := `
		UPDATE transaction_outputs
		SET spent_by_txid = $1,
		    spent_by_vin = $2,
		    spent_at_height = $3
		WHERE txid = $4 AND vout = $5
	`

	_, err = postgresTx.Exec(ctx, outputQuery, txid, vin, blockHeight, prevTxid, prevVout)
	if err != nil {
		return fmt.Errorf("failed to mark output %s:%d as spent: %w", prevTxid, prevVout, err)
	}

	return nil
}
