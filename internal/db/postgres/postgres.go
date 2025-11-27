package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
)

var DB *pgxpool.Pool

// SchemaInitFunc is a function type for module schema initialization
type SchemaInitFunc func() error

// registeredModuleSchemas holds the schema initialization functions for enabled modules
var registeredModuleSchemas = make(map[string]SchemaInitFunc)

// registeredCoreSchemas holds the schema initialization functions for core schemas (always enabled)
var registeredCoreSchemas = make(map[string]SchemaInitFunc)

// RegisterModuleSchema registers a module's schema initialization function
func RegisterModuleSchema(moduleName string, initFunc SchemaInitFunc) {
	registeredModuleSchemas[moduleName] = initFunc
}

// RegisterCoreSchema registers a core schema initialization function (always initialized)
func RegisterCoreSchema(name string, initFunc SchemaInitFunc) {
	registeredCoreSchemas[name] = initFunc
}

func InitPostgres() error {
	if !config.ShouldConnectPostgres() {
		log.Println("PostgreSQL connection disabled in config")
		return nil
	}

	cfg := config.Conf.Database

	// Build connection string with all connection parameters
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d&statement_timeout=%d",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
		cfg.ConnectTimeout, cfg.StatementTimeout*1000, // statement_timeout is in milliseconds
	)

	log.Println("Connecting to PostgreSQL...")

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.ConnectionLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.ConnectionLifetime) * time.Second

	log.Printf("Database pool configured with MaxConns: %d, MinConns: %d, MaxConnLifetime: %ds, ConnectTimeout: %ds, StatementTimeout: %ds",
		cfg.MaxConnections,
		cfg.MaxIdleConnections,
		cfg.ConnectionLifetime,
		cfg.ConnectTimeout,
		cfg.StatementTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = pool
	log.Println("PostgreSQL connected successfully")

	if err := initSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

func ClosePostgres() {
	if DB != nil {
		log.Println("Closing PostgreSQL connection...")
		DB.Close()
	}
}

func initSchema() error {
	log.Println("Initializing database schema...")

	// Core schema (always initialized)
	coreSchema := `
		CREATE TABLE IF NOT EXISTS indexer_state (
			id SERIAL PRIMARY KEY,
			last_indexed_block BIGINT NOT NULL DEFAULT 0,
			last_indexed_hash VARCHAR(64),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := DB.Exec(context.Background(), coreSchema)
	if err != nil {
		return fmt.Errorf("failed to create core schema: %w", err)
	}

	// Initialize indexer_state if empty
	var count int
	err = DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM indexer_state").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check indexer_state: %w", err)
	}

	if count == 0 {
		_, err = DB.Exec(context.Background(), "INSERT INTO indexer_state (last_indexed_block) VALUES (0)")
		if err != nil {
			return fmt.Errorf("failed to initialize indexer_state: %w", err)
		}
	}

	log.Println("Core schema initialized successfully")

	// Initialize core schemas (always initialized)
	if err := initCoreSchemas(); err != nil {
		return fmt.Errorf("failed to initialize core schemas: %w", err)
	}

	// Initialize module schemas based on configuration
	if err := initModuleSchemas(); err != nil {
		return fmt.Errorf("failed to initialize module schemas: %w", err)
	}

	return nil
}

func initCoreSchemas() error {
	// Initialize registered core schemas (always enabled)
	for name, initFunc := range registeredCoreSchemas {
		log.Printf("Initializing %s schema...", name)
		if err := initFunc(); err != nil {
			return fmt.Errorf("failed to initialize %s schema: %w", name, err)
		}
		log.Printf("%s schema initialized successfully", name)
	}

	return nil
}

func initModuleSchemas() error {
	// Initialize registered module schemas based on enabled modules in configuration
	for moduleName, initFunc := range registeredModuleSchemas {
		if config.IsModuleEnabled(moduleName) {
			log.Printf("Initializing %s module schema...", moduleName)
			if err := initFunc(); err != nil {
				return fmt.Errorf("failed to initialize %s schema: %w", moduleName, err)
			}
			log.Printf("%s module schema initialized successfully", moduleName)
		} else {
			log.Printf("Skipping %s module schema initialization (module disabled)", moduleName)
		}
	}

	return nil
}

func GetLastIndexedBlock() (int64, error) {
	var lastBlock int64
	err := DB.QueryRow(context.Background(), "SELECT last_indexed_block FROM indexer_state WHERE id = 1").Scan(&lastBlock)
	if err != nil {
		return 0, fmt.Errorf("failed to get last indexed block: %w", err)
	}
	return lastBlock, nil
}

// GetLastIndexedHash returns the hash of the last indexed block
func GetLastIndexedHash() (string, error) {
	var hash string
	err := DB.QueryRow(context.Background(), "SELECT last_indexed_hash FROM indexer_state WHERE id = 1").Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("failed to get last indexed hash: %w", err)
	}
	return hash, nil
}

// GetBlockHashAtHeight returns the stored hash at a specific height
func GetBlockHashAtHeight(height int64) (string, error) {
	var hash string
	err := DB.QueryRow(context.Background(), "SELECT hash FROM blocks WHERE height = $1", height).Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("failed to get block hash at height %d: %w", height, err)
	}
	return hash, nil
}

// RollbackToHeight removes all data after the specified height
// This handles all module tables in the correct order to maintain referential integrity
func RollbackToHeight(ctx context.Context, rollbackHeight int64) error {
	tx, err := DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	log.Printf("Starting rollback to height %d", rollbackHeight)

	// Step 1: Unspend transaction outputs that were spent after rollback height
	result, err := tx.Exec(ctx, `
		UPDATE transaction_outputs
		SET spent_by_txid = NULL,
		    spent_by_vin = NULL,
		    spent_at_height = NULL
		WHERE spent_at_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to unspend transaction outputs: %w", err)
	}
	log.Printf("Unspent %d transaction outputs", result.RowsAffected())

	// Step 2: Unspend TZE outputs that were spent after rollback height
	result, err = tx.Exec(ctx, `
		UPDATE tze_outputs
		SET spent_by_txid = NULL,
		    spent_by_vin = NULL,
		    spent_at_height = NULL
		WHERE spent_at_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to unspend TZE outputs: %w", err)
	}
	log.Printf("Unspent %d TZE outputs", result.RowsAffected())

	// Step 3: Recalculate account balances for affected accounts
	result, err = tx.Exec(ctx, `
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT SUM(balance_change)
			FROM account_transactions at
			WHERE at.address = a.address
			AND at.block_height <= $1
		), 0)
		WHERE a.address IN (
			SELECT DISTINCT address
			FROM account_transactions
			WHERE block_height > $1
		)
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to recalculate account balances: %w", err)
	}
	log.Printf("Recalculated %d account balances", result.RowsAffected())

	// Step 4: Delete account transactions after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM account_transactions WHERE block_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete account transactions: %w", err)
	}
	log.Printf("Deleted %d account transactions", result.RowsAffected())

	// Step 5: Delete orphaned accounts (accounts with no remaining transactions)
	result, err = tx.Exec(ctx, `
		DELETE FROM accounts
		WHERE address NOT IN (
			SELECT DISTINCT address FROM account_transactions
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to delete orphaned accounts: %w", err)
	}
	log.Printf("Deleted %d orphaned accounts", result.RowsAffected())

	// Step 6: Delete TZE inputs for transactions after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM tze_inputs WHERE txid IN (
			SELECT txid FROM transactions WHERE block_height > $1
		)
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete TZE inputs: %w", err)
	}
	log.Printf("Deleted %d TZE inputs", result.RowsAffected())

	// Step 7: Delete TZE outputs for transactions after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM tze_outputs WHERE txid IN (
			SELECT txid FROM transactions WHERE block_height > $1
		)
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete TZE outputs: %w", err)
	}
	log.Printf("Deleted %d TZE outputs", result.RowsAffected())

	// Step 8: Delete transactions after rollback height (CASCADE deletes inputs/outputs)
	result, err = tx.Exec(ctx, `
		DELETE FROM transactions WHERE block_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete transactions: %w", err)
	}
	log.Printf("Deleted %d transactions", result.RowsAffected())

	// Step 9: Delete STARK proofs after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM stark_proofs WHERE block_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete STARK proofs: %w", err)
	}
	log.Printf("Deleted %d STARK proofs", result.RowsAffected())

	// Step 10: Delete Ztarknet facts after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM ztarknet_facts WHERE block_height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete Ztarknet facts: %w", err)
	}
	log.Printf("Deleted %d Ztarknet facts", result.RowsAffected())

	// Step 11: Delete orphaned verifiers (verifiers with no remaining proofs/facts)
	result, err = tx.Exec(ctx, `
		DELETE FROM verifiers
		WHERE verifier_id NOT IN (
			SELECT DISTINCT verifier_id FROM stark_proofs
			UNION
			SELECT DISTINCT verifier_id FROM ztarknet_facts
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to delete orphaned verifiers: %w", err)
	}
	log.Printf("Deleted %d orphaned verifiers", result.RowsAffected())

	// Step 12: Delete blocks after rollback height
	result, err = tx.Exec(ctx, `
		DELETE FROM blocks WHERE height > $1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to delete blocks: %w", err)
	}
	log.Printf("Deleted %d blocks", result.RowsAffected())

	// Step 13: Update indexer state to rollback height
	_, err = tx.Exec(ctx, `
		UPDATE indexer_state
		SET last_indexed_block = $1,
		    last_indexed_hash = (SELECT hash FROM blocks WHERE height = $1),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, rollbackHeight)
	if err != nil {
		return fmt.Errorf("failed to update indexer state: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	log.Printf("Successfully rolled back to height %d", rollbackHeight)
	return nil
}

func UpdateLastIndexedBlock(height int64, hash string) error {
	_, err := DB.Exec(
		context.Background(),
		"UPDATE indexer_state SET last_indexed_block = $1, last_indexed_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = 1",
		height, hash,
	)
	if err != nil {
		return fmt.Errorf("failed to update last indexed block: %w", err)
	}
	return nil
}

// PostgresQuery is a helper function to run a query on the Postgres database.
//
//	Generic Param:
//	  RowType - Golang struct with json tags to map the query result.
//	Params:
//	  query - Postgres query string w/ $1, $2, etc. placeholders.
//	  args - Arguments to replace the placeholders in the query.
//	Returns:
//	  []RowType - Slice of RowType structs with the query result.
//	  error - Error if the query fails.
func PostgresQuery[RowType any](query string, args ...interface{}) ([]RowType, error) {
	var result []RowType
	err := pgxscan.Select(context.Background(), DB, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Same as PostgresQuery, but only returns the first row.
func PostgresQueryOne[RowType any](query string, args ...interface{}) (*RowType, error) {
	var result RowType
	err := pgxscan.Get(context.Background(), DB, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Same as PostgresQuery, but returns the result as a Marshalled JSON byte array.
func PostgresQueryJson[RowType any](query string, args ...interface{}) ([]byte, error) {
	result, err := PostgresQuery[RowType](query, args...)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

// Same as PostgresQueryOne, but returns the result as a Marshalled JSON byte array.
func PostgresQueryOneJson[RowType any](query string, args ...interface{}) ([]byte, error) {
	result, err := PostgresQueryOne[RowType](query, args...)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
