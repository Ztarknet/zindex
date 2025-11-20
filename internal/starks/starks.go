package starks

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func init() {
	// Register this module's schema initialization with the postgres package
	postgres.RegisterModuleSchema("STARKS", InitSchema)
}

// InitSchema creates the starks module tables and indexes
func InitSchema() error {
	schema := `
		-- Verifiers table
		CREATE TABLE IF NOT EXISTS verifiers (
			verifier_id VARCHAR(80) PRIMARY KEY,  -- txid (64) + ":" (1) + vout (up to 10 digits)
			verifier_name VARCHAR(255) NOT NULL,
			verifier_metadata TEXT,
			balance BIGINT NOT NULL DEFAULT 0,
			first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- STARK proofs table
		CREATE TABLE IF NOT EXISTS stark_proofs (
			verifier_id VARCHAR(80) NOT NULL,  -- matches verifiers.verifier_id
			txid VARCHAR(64) NOT NULL,
			block_height BIGINT NOT NULL,
			proof_size BIGINT NOT NULL,
			PRIMARY KEY (verifier_id, txid),
			FOREIGN KEY (verifier_id) REFERENCES verifiers(verifier_id) ON DELETE CASCADE
		);

		-- Ztarknet facts table
		CREATE TABLE IF NOT EXISTS ztarknet_facts (
			verifier_id VARCHAR(80) NOT NULL,  -- matches verifiers.verifier_id
			txid VARCHAR(64) NOT NULL,
			block_height BIGINT NOT NULL,
			proof_size BIGINT NOT NULL,
			old_state VARCHAR(64) NOT NULL,
			new_state VARCHAR(64) NOT NULL,
			program_hash VARCHAR(64) NOT NULL,
			inner_program_hash VARCHAR(64) NOT NULL,
			PRIMARY KEY (verifier_id, txid),
			FOREIGN KEY (verifier_id) REFERENCES verifiers(verifier_id) ON DELETE CASCADE
		);

		-- Indexes for verifiers
		CREATE INDEX IF NOT EXISTS idx_verifiers_name ON verifiers(verifier_name);
		CREATE INDEX IF NOT EXISTS idx_verifiers_first_seen ON verifiers(first_seen_at);
		CREATE INDEX IF NOT EXISTS idx_verifiers_balance ON verifiers(balance);

		-- Indexes for stark_proofs
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_txid ON stark_proofs(txid);
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_block_height ON stark_proofs(block_height);
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_verifier ON stark_proofs(verifier_id);
		CREATE INDEX IF NOT EXISTS idx_stark_proofs_size ON stark_proofs(proof_size);

		-- Indexes for ztarknet_facts
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_txid ON ztarknet_facts(txid);
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_block_height ON ztarknet_facts(block_height);
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_verifier ON ztarknet_facts(verifier_id);
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_old_state ON ztarknet_facts(old_state);
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_new_state ON ztarknet_facts(new_state);
		CREATE INDEX IF NOT EXISTS idx_ztarknet_facts_program_hash ON ztarknet_facts(program_hash);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create starks schema: %w", err)
	}

	return nil
}

// ShouldIndexZtarknet returns whether ztarknet facts should be indexed based on configuration
func ShouldIndexZtarknet() bool {
	return config.Conf.Modules.Starks.Enabled && config.Conf.Modules.Starks.IndexZtarknet
}

// ============================================================================
// Verifier Query Functions
// ============================================================================

// GetVerifier retrieves a verifier by its ID
func GetVerifier(verifierID string) (*Verifier, error) {
	verifier, err := postgres.PostgresQueryOne[Verifier](
		`SELECT verifier_id, verifier_name, verifier_metadata, balance, first_seen_at
		 FROM verifiers WHERE verifier_id = $1`,
		verifierID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verifier: %w", err)
	}

	return verifier, nil
}

// GetVerifierByName retrieves a verifier by its name
func GetVerifierByName(verifierName string) (*Verifier, error) {
	verifier, err := postgres.PostgresQueryOne[Verifier](
		`SELECT verifier_id, verifier_name, verifier_metadata, balance, first_seen_at
		 FROM verifiers WHERE verifier_name = $1`,
		verifierName,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verifier by name: %w", err)
	}

	return verifier, nil
}

// GetAllVerifiers retrieves all verifiers with pagination
func GetAllVerifiers(limit, offset int) ([]Verifier, error) {
	verifiers, err := postgres.PostgresQuery[Verifier](
		`SELECT verifier_id, verifier_name, verifier_metadata, balance, first_seen_at
		 FROM verifiers
		 ORDER BY first_seen_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifiers: %w", err)
	}

	return verifiers, nil
}

// GetVerifiersByBalance retrieves verifiers sorted by balance
func GetVerifiersByBalance(limit, offset int) ([]Verifier, error) {
	verifiers, err := postgres.PostgresQuery[Verifier](
		`SELECT verifier_id, verifier_name, verifier_metadata, balance, first_seen_at
		 FROM verifiers
		 ORDER BY balance DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifiers by balance: %w", err)
	}

	return verifiers, nil
}

// ============================================================================
// StarkProof Query Functions
// ============================================================================

// GetStarkProof retrieves a STARK proof by verifier ID and transaction ID
func GetStarkProof(verifierID, txid string) (*StarkProof, error) {
	proof, err := postgres.PostgresQueryOne[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 WHERE verifier_id = $1 AND txid = $2`,
		verifierID, txid,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stark proof: %w", err)
	}

	return proof, nil
}

// GetStarkProofsByVerifier retrieves all STARK proofs for a verifier
func GetStarkProofsByVerifier(verifierID string, limit, offset int) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 WHERE verifier_id = $1
		 ORDER BY block_height DESC
		 LIMIT $2 OFFSET $3`,
		verifierID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stark proofs by verifier: %w", err)
	}

	return proofs, nil
}

// GetStarkProofsByTransaction retrieves all STARK proofs for a transaction
func GetStarkProofsByTransaction(txid string) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 WHERE txid = $1
		 ORDER BY verifier_id`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stark proofs by transaction: %w", err)
	}

	return proofs, nil
}

// GetStarkProofsByBlock retrieves all STARK proofs for a block
func GetStarkProofsByBlock(blockHeight int64) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 WHERE block_height = $1
		 ORDER BY txid`,
		blockHeight,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stark proofs by block: %w", err)
	}

	return proofs, nil
}

// GetRecentStarkProofs retrieves the most recent STARK proofs
func GetRecentStarkProofs(limit, offset int) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 ORDER BY block_height DESC, txid
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent stark proofs: %w", err)
	}

	return proofs, nil
}

// GetStarkProofsBySize retrieves STARK proofs filtered by size range
func GetStarkProofsBySize(minSize, maxSize int64, limit, offset int) ([]StarkProof, error) {
	proofs, err := postgres.PostgresQuery[StarkProof](
		`SELECT verifier_id, txid, block_height, proof_size
		 FROM stark_proofs
		 WHERE proof_size >= $1 AND proof_size <= $2
		 ORDER BY proof_size DESC
		 LIMIT $3 OFFSET $4`,
		minSize, maxSize, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stark proofs by size: %w", err)
	}

	return proofs, nil
}

// ============================================================================
// ZtarknetFacts Query Functions
// ============================================================================

// GetZtarknetFacts retrieves Ztarknet facts by verifier ID and transaction ID
func GetZtarknetFacts(verifierID, txid string) (*ZtarknetFacts, error) {
	facts, err := postgres.PostgresQueryOne[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE verifier_id = $1 AND txid = $2`,
		verifierID, txid,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByVerifier retrieves all Ztarknet facts for a verifier
func GetZtarknetFactsByVerifier(verifierID string, limit, offset int) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE verifier_id = $1
		 ORDER BY block_height DESC
		 LIMIT $2 OFFSET $3`,
		verifierID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by verifier: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByTransaction retrieves all Ztarknet facts for a transaction
func GetZtarknetFactsByTransaction(txid string) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE txid = $1
		 ORDER BY verifier_id`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by transaction: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByBlock retrieves all Ztarknet facts for a block
func GetZtarknetFactsByBlock(blockHeight int64) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE block_height = $1
		 ORDER BY txid`,
		blockHeight,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by block: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByState retrieves Ztarknet facts by state hash
func GetZtarknetFactsByState(stateHash string) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE old_state = $1 OR new_state = $1
		 ORDER BY block_height DESC`,
		stateHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by state: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByProgramHash retrieves Ztarknet facts by program hash
func GetZtarknetFactsByProgramHash(programHash string) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE program_hash = $1
		 ORDER BY block_height DESC`,
		programHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by program hash: %w", err)
	}

	return facts, nil
}

// GetZtarknetFactsByInnerProgramHash retrieves Ztarknet facts by inner program hash
func GetZtarknetFactsByInnerProgramHash(innerProgramHash string) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE inner_program_hash = $1
		 ORDER BY block_height DESC`,
		innerProgramHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ztarknet facts by inner program hash: %w", err)
	}

	return facts, nil
}

// GetRecentZtarknetFacts retrieves the most recent Ztarknet facts
func GetRecentZtarknetFacts(limit, offset int) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 ORDER BY block_height DESC, txid
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent ztarknet facts: %w", err)
	}

	return facts, nil
}

// GetStateTransition retrieves the state transition from old_state to new_state
func GetStateTransition(oldState, newState string) ([]ZtarknetFacts, error) {
	facts, err := postgres.PostgresQuery[ZtarknetFacts](
		`SELECT verifier_id, txid, block_height, proof_size, old_state, new_state,
		        program_hash, inner_program_hash
		 FROM ztarknet_facts
		 WHERE old_state = $1 AND new_state = $2
		 ORDER BY block_height DESC`,
		oldState, newState,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get state transition: %w", err)
	}

	return facts, nil
}

// ============================================================================
// STORAGE FUNCTIONS
// ============================================================================

// DBTX is an interface that both pgxpool.Pool and pgx.Tx implement
// This allows functions to work with either a connection pool or a transaction
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// StoreVerifier inserts or updates a verifier in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreVerifier(postgresTx DBTX, verifierID, verifierName, verifierMetadata string, balance int64) error {
	ctx := context.Background()

	query := `
		INSERT INTO verifiers (verifier_id, verifier_name, verifier_metadata, balance)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (verifier_id) DO UPDATE SET
			verifier_name = EXCLUDED.verifier_name,
			verifier_metadata = EXCLUDED.verifier_metadata,
			balance = EXCLUDED.balance
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, verifierID, verifierName, verifierMetadata, balance)
	if err != nil {
		return fmt.Errorf("failed to store verifier %s: %w", verifierID, err)
	}

	return nil
}

// UpdateVerifierBalance updates the balance of an existing verifier
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func UpdateVerifierBalance(postgresTx DBTX, verifierID string, balance int64) error {
	ctx := context.Background()

	query := `
		UPDATE verifiers
		SET balance = $2
		WHERE verifier_id = $1
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, verifierID, balance)
	if err != nil {
		return fmt.Errorf("failed to update verifier %s balance: %w", verifierID, err)
	}

	return nil
}

// StoreStarkProof inserts or updates a STARK proof in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreStarkProof(postgresTx DBTX, verifierID, txid string, blockHeight, proofSize int64) error {
	ctx := context.Background()

	query := `
		INSERT INTO stark_proofs (verifier_id, txid, block_height, proof_size)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (verifier_id, txid) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			proof_size = EXCLUDED.proof_size
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, verifierID, txid, blockHeight, proofSize)
	if err != nil {
		return fmt.Errorf("failed to store STARK proof for verifier %s, tx %s: %w", verifierID, txid, err)
	}

	return nil
}

// StoreZtarknetFacts inserts or updates Ztarknet facts in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreZtarknetFacts(postgresTx DBTX, verifierID, txid string, blockHeight, proofSize int64,
	oldState, newState, programHash, innerProgramHash string) error {
	ctx := context.Background()

	query := `
		INSERT INTO ztarknet_facts (verifier_id, txid, block_height, proof_size,
		                            old_state, new_state, program_hash, inner_program_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (verifier_id, txid) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			proof_size = EXCLUDED.proof_size,
			old_state = EXCLUDED.old_state,
			new_state = EXCLUDED.new_state,
			program_hash = EXCLUDED.program_hash,
			inner_program_hash = EXCLUDED.inner_program_hash
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, verifierID, txid, blockHeight, proofSize,
		oldState, newState, programHash, innerProgramHash)
	if err != nil {
		return fmt.Errorf("failed to store Ztarknet facts for verifier %s, tx %s: %w", verifierID, txid, err)
	}

	return nil
}

// ============================================================================
// Count Functions
// ============================================================================

// CountVerifiers returns the total count of verifiers with optional filters
func CountVerifiers() (int64, error) {
	var count int64
	err := postgres.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM verifiers`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count verifiers: %w", err)
	}
	return count, nil
}

// CountStarkProofs returns the total count of stark proofs with optional filters
func CountStarkProofs(verifierID string, blockHeight int64) (int64, error) {
	var query string
	var args []interface{}

	if verifierID != "" && blockHeight > 0 {
		query = `SELECT COUNT(*) FROM stark_proofs WHERE verifier_id = $1 AND block_height = $2`
		args = []interface{}{verifierID, blockHeight}
	} else if verifierID != "" {
		query = `SELECT COUNT(*) FROM stark_proofs WHERE verifier_id = $1`
		args = []interface{}{verifierID}
	} else if blockHeight > 0 {
		query = `SELECT COUNT(*) FROM stark_proofs WHERE block_height = $1`
		args = []interface{}{blockHeight}
	} else {
		query = `SELECT COUNT(*) FROM stark_proofs`
		args = []interface{}{}
	}

	var count int64
	err := postgres.DB.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count stark proofs: %w", err)
	}

	return count, nil
}

// CountZtarknetFacts returns the total count of ztarknet facts with optional filters
func CountZtarknetFacts(verifierID string, blockHeight int64) (int64, error) {
	var query string
	var args []interface{}

	if verifierID != "" && blockHeight > 0 {
		query = `SELECT COUNT(*) FROM ztarknet_facts WHERE verifier_id = $1 AND block_height = $2`
		args = []interface{}{verifierID, blockHeight}
	} else if verifierID != "" {
		query = `SELECT COUNT(*) FROM ztarknet_facts WHERE verifier_id = $1`
		args = []interface{}{verifierID}
	} else if blockHeight > 0 {
		query = `SELECT COUNT(*) FROM ztarknet_facts WHERE block_height = $1`
		args = []interface{}{blockHeight}
	} else {
		query = `SELECT COUNT(*) FROM ztarknet_facts`
		args = []interface{}{}
	}

	var count int64
	err := postgres.DB.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ztarknet facts: %w", err)
	}

	return count, nil
}

// SumStarkProofSizesByVerifier returns the sum of all proof sizes for a given verifier
func SumStarkProofSizesByVerifier(verifierID string) (int64, error) {
	var sum int64
	err := postgres.DB.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(proof_size), 0) FROM stark_proofs WHERE verifier_id = $1`,
		verifierID,
	).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("failed to sum proof sizes for verifier %s: %w", verifierID, err)
	}
	return sum, nil
}
