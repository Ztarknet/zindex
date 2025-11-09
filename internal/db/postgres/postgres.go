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

	schema := `
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

		CREATE TABLE IF NOT EXISTS transactions (
			txid VARCHAR(64) PRIMARY KEY,
			block_height BIGINT REFERENCES blocks(height),
			version INT,
			locktime BIGINT,
			expiry_height BIGINT,
			size INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_transactions_block_height ON transactions(block_height);
		CREATE INDEX IF NOT EXISTS idx_transactions_expiry ON transactions(expiry_height);
	`

	_, err := DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

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

	log.Println("Database schema initialized successfully")
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
