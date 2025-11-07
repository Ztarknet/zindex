package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
)

var DB *sql.DB

func InitPostgres() error {
	if !config.ShouldConnectPostgres() {
		log.Println("PostgreSQL connection disabled in config")
		return nil
	}

	cfg := config.Conf.Database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	log.Println("Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
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

	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM indexer_state").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check indexer_state: %w", err)
	}

	if count == 0 {
		_, err = DB.Exec("INSERT INTO indexer_state (last_indexed_block) VALUES (0)")
		if err != nil {
			return fmt.Errorf("failed to initialize indexer_state: %w", err)
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}

func GetLastIndexedBlock() (int64, error) {
	var lastBlock int64
	err := DB.QueryRow("SELECT last_indexed_block FROM indexer_state WHERE id = 1").Scan(&lastBlock)
	if err != nil {
		return 0, fmt.Errorf("failed to get last indexed block: %w", err)
	}
	return lastBlock, nil
}

func UpdateLastIndexedBlock(height int64, hash string) error {
	_, err := DB.Exec(
		"UPDATE indexer_state SET last_indexed_block = $1, last_indexed_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = 1",
		height, hash,
	)
	if err != nil {
		return fmt.Errorf("failed to update last indexed block: %w", err)
	}
	return nil
}
