package blocks

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func init() {
	// Register this as a core schema (always initialized)
	postgres.RegisterCoreSchema("blocks", InitSchema)
}

// InitSchema creates the blocks table and indexes
// This is part of the core schema and is always initialized
func InitSchema() error {
	schema := `
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

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create blocks schema: %w", err)
	}

	return nil
}

// StoreBlock inserts or updates a block in the database
func StoreBlock(height int64, hash string, prevHash string, merkleRoot string, timestamp int64, difficulty float64, nonce string, version int, txCount int) error {
	ctx := context.Background()

	// Convert difficulty to string for storage
	difficultyStr := fmt.Sprintf("%f", difficulty)

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

	_, err := postgres.DB.Exec(ctx, query, height, hash, prevHash, merkleRoot, timestamp, difficultyStr, nonce, version, txCount)
	if err != nil {
		return fmt.Errorf("failed to store block %d: %w", height, err)
	}

	return nil
}

// GetBlock retrieves a block by its height
func GetBlock(height int64) (*Block, error) {
	block, err := postgres.PostgresQueryOne[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks WHERE height = $1`,
		height,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return block, nil
}

// GetBlockByHash retrieves a block by its hash
func GetBlockByHash(hash string) (*Block, error) {
	block, err := postgres.PostgresQueryOne[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks WHERE hash = $1`,
		hash,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	return block, nil
}

// GetBlocks retrieves blocks with pagination
func GetBlocks(limit, offset int) ([]Block, error) {
	blocks, err := postgres.PostgresQuery[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks
		 ORDER BY height DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}

	return blocks, nil
}

// GetBlocksByRange retrieves blocks within a height range
func GetBlocksByRange(fromHeight, toHeight int64, limit, offset int) ([]Block, error) {
	blocks, err := postgres.PostgresQuery[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks
		 WHERE height >= $1 AND height <= $2
		 ORDER BY height DESC
		 LIMIT $3 OFFSET $4`,
		fromHeight, toHeight, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks by range: %w", err)
	}

	return blocks, nil
}

// GetBlocksByTimestampRange retrieves blocks within a timestamp range
func GetBlocksByTimestampRange(fromTimestamp, toTimestamp int64, limit, offset int) ([]Block, error) {
	blocks, err := postgres.PostgresQuery[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks
		 WHERE timestamp >= $1 AND timestamp <= $2
		 ORDER BY timestamp DESC
		 LIMIT $3 OFFSET $4`,
		fromTimestamp, toTimestamp, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks by timestamp range: %w", err)
	}

	return blocks, nil
}

// GetRecentBlocks retrieves the most recent blocks
func GetRecentBlocks(limit int) ([]Block, error) {
	blocks, err := postgres.PostgresQuery[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks
		 ORDER BY height DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent blocks: %w", err)
	}

	return blocks, nil
}

// GetBlockCount returns the total number of blocks
func GetBlockCount() (int64, error) {
	type result struct {
		Count int64 `db:"count"`
	}

	res, err := postgres.PostgresQueryOne[result](
		`SELECT COUNT(*) as count FROM blocks`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}

	return res.Count, nil
}

// GetLatestBlock retrieves the most recent block
func GetLatestBlock() (*Block, error) {
	block, err := postgres.PostgresQueryOne[Block](
		`SELECT height, hash, prev_hash, merkle_root, timestamp, difficulty, nonce, version, tx_count, created_at
		 FROM blocks
		 ORDER BY height DESC
		 LIMIT 1`,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return block, nil
}
