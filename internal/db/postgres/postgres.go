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

// RegisterModuleSchema registers a module's schema initialization function
func RegisterModuleSchema(moduleName string, initFunc SchemaInitFunc) {
	registeredModuleSchemas[moduleName] = initFunc
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

		CREATE TABLE IF NOT EXISTS blocks (
			height BIGINT PRIMARY KEY,
			hash VARCHAR(64) NOT NULL UNIQUE,
			prev_hash VARCHAR(64),
			merkle_root VARCHAR(64),
			timestamp BIGINT,
			difficulty VARCHAR(64),
			nonce VARCHAR(64),
			version INT,
			tx_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash);
		CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp);
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

	// Initialize module schemas based on configuration
	if err := initModuleSchemas(); err != nil {
		return fmt.Errorf("failed to initialize module schemas: %w", err)
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

// TODO: Add reorg handling functions when needed:
// - GetLastIndexedHash() - Get last indexed block hash
// - GetBlockAtHeight(height) - Get stored block hash at specific height
// - RollbackToBlock(height) - Remove blocks after specified height

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

// StoreBlock inserts or updates a block in the database
func StoreBlock(height int64, hash string, prevHash string, block map[string]interface{}) error {
	ctx := context.Background()

	// Extract block fields
	var merkleRoot, difficulty, nonce string
	var timestamp int64
	var version, txCount int

	if val, ok := block["merkleroot"].(string); ok {
		merkleRoot = val
	}
	if val, ok := block["difficulty"].(float64); ok {
		difficulty = fmt.Sprintf("%f", val)
	} else if val, ok := block["difficulty"].(string); ok {
		difficulty = val
	}
	if val, ok := block["nonce"].(string); ok {
		nonce = val
	}
	if val, ok := block["time"].(float64); ok {
		timestamp = int64(val)
	}
	if val, ok := block["version"].(float64); ok {
		version = int(val)
	}
	if tx, ok := block["tx"].([]interface{}); ok {
		txCount = len(tx)
	}

	query := `
		INSERT INTO blocks (height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (height) DO UPDATE SET
			hash = EXCLUDED.hash,
			prev_hash = EXCLUDED.prev_hash,
			merkle_root = EXCLUDED.merkle_root,
			timestamp = EXCLUDED.timestamp,
			difficulty = EXCLUDED.difficulty,
			nonce = EXCLUDED.nonce,
			version = EXCLUDED.version,
			tx_count = EXCLUDED.tx_count
	`

	_, err := DB.Exec(ctx, query, height, hash, prevHash, merkleRoot, timestamp, difficulty, nonce, version, txCount)
	if err != nil {
		return fmt.Errorf("failed to store block %d: %w", height, err)
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
