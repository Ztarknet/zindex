package tze_graph

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func init() {
	// Register this module's schema initialization with the postgres package
	postgres.RegisterModuleSchema("TZE_GRAPH", InitSchema)
}

func InitSchema() error {
	schema := `
		-- TZE Inputs table
		CREATE TABLE IF NOT EXISTS tze_inputs (
			txid VARCHAR(64) NOT NULL,
			vin INT NOT NULL,
			value BIGINT NOT NULL,
			prev_txid VARCHAR(64) NOT NULL,
			prev_vout INT NOT NULL,
			tze_type INT NOT NULL,  -- 4-byte extension_id (0=demo, 1=stark_verify)
			tze_mode INT NOT NULL,  -- 4-byte mode (demo: 0=open, 1=close; stark_verify: 0=initialize, 1=verify)
			PRIMARY KEY (txid, vin)
		);

		-- Indexes for tze_inputs
		CREATE INDEX IF NOT EXISTS idx_tze_inputs_txid ON tze_inputs(txid);
		CREATE INDEX IF NOT EXISTS idx_tze_inputs_prev ON tze_inputs(prev_txid, prev_vout);
		CREATE INDEX IF NOT EXISTS idx_tze_inputs_type ON tze_inputs(tze_type);
		CREATE INDEX IF NOT EXISTS idx_tze_inputs_mode ON tze_inputs(tze_mode);
		CREATE INDEX IF NOT EXISTS idx_tze_inputs_type_mode ON tze_inputs(tze_type, tze_mode);

		-- TZE Outputs table
		CREATE TABLE IF NOT EXISTS tze_outputs (
			txid VARCHAR(64) NOT NULL,
			vout INT NOT NULL,
			value BIGINT NOT NULL,
			spent_by_txid VARCHAR(64),
			spent_by_vin INT,
			spent_at_height BIGINT,
			tze_type INT NOT NULL,  -- 4-byte extension_id (0=demo, 1=stark_verify)
			tze_mode INT NOT NULL,  -- 4-byte mode (demo: 0=open, 1=close; stark_verify: 0=initialize, 1=verify)
			precondition BYTEA,
			PRIMARY KEY (txid, vout)
		);

		-- Indexes for tze_outputs
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_txid ON tze_outputs(txid);
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_spent_by ON tze_outputs(spent_by_txid)
			WHERE spent_by_txid IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_unspent ON tze_outputs(txid, vout)
			WHERE spent_by_txid IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_type ON tze_outputs(tze_type);
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_mode ON tze_outputs(tze_mode);
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_type_mode ON tze_outputs(tze_type, tze_mode);
		CREATE INDEX IF NOT EXISTS idx_tze_outputs_value ON tze_outputs(value);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create tze_graph schema: %w", err)
	}

	return nil
}

// ValidatePreconditionSize validates that a precondition does not exceed the configured maximum size
func ValidatePreconditionSize(precondition []byte) error {
	maxSize := config.Conf.Modules.TzeGraph.MaxPreconditionSize
	if len(precondition) > maxSize {
		return fmt.Errorf("precondition size (%d bytes) exceeds maximum allowed size (%d bytes)",
			len(precondition), maxSize)
	}
	return nil
}

// ============================================================================
// TZE INPUT QUERIES
// ============================================================================

// GetTzeInputs retrieves all inputs for a transaction
func GetTzeInputs(txid string) ([]TzeInput, error) {
	inputs, err := postgres.PostgresQuery[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE txid = $1
		 ORDER BY vin`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze inputs: %w", err)
	}

	return inputs, nil
}

// GetTzeInput retrieves a specific input by txid and vin
func GetTzeInput(txid string, vin int) (*TzeInput, error) {
	input, err := postgres.PostgresQueryOne[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE txid = $1 AND vin = $2`,
		txid, vin,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tze input: %w", err)
	}

	return input, nil
}

// GetTzeInputsByType retrieves all inputs of a specific TZE type with pagination
func GetTzeInputsByType(tzeType TzeType, limit, offset int) ([]TzeInput, error) {
	inputs, err := postgres.PostgresQuery[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE tze_type = $1
		 ORDER BY txid, vin
		 LIMIT $2 OFFSET $3`,
		tzeType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze inputs by type: %w", err)
	}

	return inputs, nil
}

// GetTzeInputsByMode retrieves all inputs of a specific TZE mode with pagination
func GetTzeInputsByMode(tzeMode TzeMode, limit, offset int) ([]TzeInput, error) {
	inputs, err := postgres.PostgresQuery[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE tze_mode = $1
		 ORDER BY txid, vin
		 LIMIT $2 OFFSET $3`,
		tzeMode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze inputs by mode: %w", err)
	}

	return inputs, nil
}

// GetTzeInputsByTypeAndMode retrieves all inputs matching both type and mode with pagination
func GetTzeInputsByTypeAndMode(tzeType TzeType, tzeMode TzeMode, limit, offset int) ([]TzeInput, error) {
	inputs, err := postgres.PostgresQuery[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE tze_type = $1 AND tze_mode = $2
		 ORDER BY txid, vin
		 LIMIT $3 OFFSET $4`,
		tzeType, tzeMode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze inputs by type and mode: %w", err)
	}

	return inputs, nil
}

// GetTzeInputsByPrevOutput retrieves all inputs spending a specific previous output
func GetTzeInputsByPrevOutput(prevTxid string, prevVout int) ([]TzeInput, error) {
	inputs, err := postgres.PostgresQuery[TzeInput](
		`SELECT txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode
		 FROM tze_inputs
		 WHERE prev_txid = $1 AND prev_vout = $2
		 ORDER BY txid, vin`,
		prevTxid, prevVout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze inputs by prev output: %w", err)
	}

	return inputs, nil
}

// ============================================================================
// TZE OUTPUT QUERIES
// ============================================================================

// GetTzeOutputs retrieves all outputs for a transaction
func GetTzeOutputs(txid string) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE txid = $1
		 ORDER BY vout`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze outputs: %w", err)
	}

	return outputs, nil
}

// GetTzeOutput retrieves a specific output by txid and vout
func GetTzeOutput(txid string, vout int) (*TzeOutput, error) {
	output, err := postgres.PostgresQueryOne[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE txid = $1 AND vout = $2`,
		txid, vout,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tze output: %w", err)
	}

	return output, nil
}

// GetUnspentTzeOutputs retrieves all unspent outputs for a transaction
func GetUnspentTzeOutputs(txid string) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE txid = $1 AND spent_by_txid IS NULL
		 ORDER BY vout`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get unspent tze outputs: %w", err)
	}

	return outputs, nil
}

// GetAllUnspentTzeOutputs retrieves all unspent TZE outputs with pagination
func GetAllUnspentTzeOutputs(limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE spent_by_txid IS NULL
		 ORDER BY txid, vout
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get all unspent tze outputs: %w", err)
	}

	return outputs, nil
}

// GetTzeOutputsByType retrieves all outputs of a specific TZE type with pagination
func GetTzeOutputsByType(tzeType TzeType, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE tze_type = $1
		 ORDER BY txid, vout
		 LIMIT $2 OFFSET $3`,
		tzeType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze outputs by type: %w", err)
	}

	return outputs, nil
}

// GetTzeOutputsByMode retrieves all outputs of a specific TZE mode with pagination
func GetTzeOutputsByMode(tzeMode TzeMode, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE tze_mode = $1
		 ORDER BY txid, vout
		 LIMIT $2 OFFSET $3`,
		tzeMode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze outputs by mode: %w", err)
	}

	return outputs, nil
}

// GetTzeOutputsByTypeAndMode retrieves all outputs matching both type and mode with pagination
func GetTzeOutputsByTypeAndMode(tzeType TzeType, tzeMode TzeMode, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE tze_type = $1 AND tze_mode = $2
		 ORDER BY txid, vout
		 LIMIT $3 OFFSET $4`,
		tzeType, tzeMode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze outputs by type and mode: %w", err)
	}

	return outputs, nil
}

// GetUnspentTzeOutputsByType retrieves all unspent outputs of a specific type with pagination
func GetUnspentTzeOutputsByType(tzeType TzeType, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE tze_type = $1 AND spent_by_txid IS NULL
		 ORDER BY txid, vout
		 LIMIT $2 OFFSET $3`,
		tzeType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get unspent tze outputs by type: %w", err)
	}

	return outputs, nil
}

// GetUnspentTzeOutputsByTypeAndMode retrieves all unspent outputs matching type and mode
func GetUnspentTzeOutputsByTypeAndMode(tzeType TzeType, tzeMode TzeMode, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE tze_type = $1 AND tze_mode = $2 AND spent_by_txid IS NULL
		 ORDER BY txid, vout
		 LIMIT $3 OFFSET $4`,
		tzeType, tzeMode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get unspent tze outputs by type and mode: %w", err)
	}

	return outputs, nil
}

// GetSpentTzeOutputs retrieves all spent outputs with pagination
func GetSpentTzeOutputs(limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE spent_by_txid IS NOT NULL
		 ORDER BY spent_at_height DESC, txid, vout
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get spent tze outputs: %w", err)
	}

	return outputs, nil
}

// GetTzeOutputsByValue retrieves outputs with value greater than or equal to minimum value
func GetTzeOutputsByValue(minValue int64, limit, offset int) ([]TzeOutput, error) {
	outputs, err := postgres.PostgresQuery[TzeOutput](
		`SELECT txid, vout, value, spent_by_txid, spent_by_vin, spent_at_height,
		        tze_type, tze_mode, precondition
		 FROM tze_outputs
		 WHERE value >= $1
		 ORDER BY value DESC, txid, vout
		 LIMIT $2 OFFSET $3`,
		minValue, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tze outputs by value: %w", err)
	}

	return outputs, nil
}

// ============================================================================
// TZE STORAGE FUNCTIONS
// ============================================================================

// DBTX is an interface that both pgxpool.Pool and pgx.Tx implement
// This allows functions to work with either a connection pool or a transaction
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// StoreTzeOutput inserts or updates a TZE output in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
// If the precondition exceeds the maximum size, it will be stored as an empty byte array
func StoreTzeOutput(postgresTx DBTX, txid string, vout int, value int64, tzeType int32, tzeMode int32, precondition []byte) error {
	ctx := context.Background()

	// Validate precondition size - if it exceeds max size, store empty byte array instead
	if err := ValidatePreconditionSize(precondition); err != nil {
		log.Printf("Warning: Precondition for output %s:%d exceeds maximum size, storing empty precondition: %v", txid, vout, err)
		precondition = []byte{}
	}

	query := `
		INSERT INTO tze_outputs (txid, vout, value, tze_type, tze_mode, precondition)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (txid, vout) DO UPDATE SET
			value = EXCLUDED.value,
			tze_type = EXCLUDED.tze_type,
			tze_mode = EXCLUDED.tze_mode,
			precondition = EXCLUDED.precondition
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, txid, vout, value, tzeType, tzeMode, precondition)
	if err != nil {
		return fmt.Errorf("failed to store tze output %s:%d: %w", txid, vout, err)
	}

	return nil
}

// StoreTzeInput inserts or updates a TZE input in the database
// and marks the corresponding TZE output as spent
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreTzeInput(postgresTx DBTX, txid string, vin int, value int64, prevTxid string, prevVout int, tzeType int32, tzeMode int32, blockHeight int64) error {
	ctx := context.Background()

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	// Insert the TZE input
	inputQuery := `
		INSERT INTO tze_inputs (txid, vin, value, prev_txid, prev_vout, tze_type, tze_mode)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (txid, vin) DO UPDATE SET
			value = EXCLUDED.value,
			prev_txid = EXCLUDED.prev_txid,
			prev_vout = EXCLUDED.prev_vout,
			tze_type = EXCLUDED.tze_type,
			tze_mode = EXCLUDED.tze_mode
	`

	_, err := postgresTx.Exec(ctx, inputQuery, txid, vin, value, prevTxid, prevVout, tzeType, tzeMode)
	if err != nil {
		return fmt.Errorf("failed to store tze input %s:%d: %w", txid, vin, err)
	}

	// Mark the previous TZE output as spent
	outputQuery := `
		UPDATE tze_outputs
		SET spent_by_txid = $1,
		    spent_by_vin = $2,
		    spent_at_height = $3
		WHERE txid = $4 AND vout = $5
	`

	_, err = postgresTx.Exec(ctx, outputQuery, txid, vin, blockHeight, prevTxid, prevVout)
	if err != nil {
		return fmt.Errorf("failed to mark tze output %s:%d as spent: %w", prevTxid, prevVout, err)
	}

	return nil
}
