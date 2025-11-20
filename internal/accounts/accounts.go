package accounts

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
	postgres.RegisterModuleSchema("ACCOUNTS", InitSchema)
}

// InitSchema creates the account tables and indexes
func InitSchema() error {
	schema := `
		-- Accounts table
		CREATE TABLE IF NOT EXISTS accounts (
			address VARCHAR(255) PRIMARY KEY,
			balance BIGINT NOT NULL DEFAULT 0,
			first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Account transactions table
		CREATE TABLE IF NOT EXISTS account_transactions (
			address VARCHAR(255) NOT NULL,
			txid VARCHAR(64) NOT NULL,
			block_height BIGINT NOT NULL,
			type VARCHAR(10) NOT NULL,
			balance_change BIGINT NOT NULL DEFAULT 0,
			PRIMARY KEY (address, txid),
			FOREIGN KEY (address) REFERENCES accounts(address) ON DELETE CASCADE
		);

		-- Indexes for accounts
		CREATE INDEX IF NOT EXISTS idx_accounts_balance ON accounts(balance);
		CREATE INDEX IF NOT EXISTS idx_accounts_first_seen_at ON accounts(first_seen_at);

		-- Indexes for account transactions
		CREATE INDEX IF NOT EXISTS idx_account_txs_address ON account_transactions(address);
		CREATE INDEX IF NOT EXISTS idx_account_txs_txid ON account_transactions(txid);
		CREATE INDEX IF NOT EXISTS idx_account_txs_block_height ON account_transactions(block_height);
		CREATE INDEX IF NOT EXISTS idx_account_txs_type ON account_transactions(type);
		CREATE INDEX IF NOT EXISTS idx_account_txs_address_block ON account_transactions(address, block_height DESC);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create account schema: %w", err)
	}

	return nil
}

// GetAccount retrieves an account by its address
func GetAccount(address string) (*Account, error) {
	account, err := postgres.PostgresQueryOne[Account](
		`SELECT address, balance, first_seen_at
		 FROM accounts WHERE address = $1`,
		address,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// GetAccounts retrieves accounts with pagination
func GetAccounts(limit, offset int) ([]Account, error) {
	accounts, err := postgres.PostgresQuery[Account](
		`SELECT address, balance, first_seen_at
		 FROM accounts
		 ORDER BY balance DESC, first_seen_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return accounts, nil
}

// GetAccountsByBalanceRange retrieves accounts within a balance range
func GetAccountsByBalanceRange(minBalance, maxBalance int64, limit, offset int) ([]Account, error) {
	accounts, err := postgres.PostgresQuery[Account](
		`SELECT address, balance, first_seen_at
		 FROM accounts
		 WHERE balance >= $1 AND balance <= $2
		 ORDER BY balance DESC, first_seen_at DESC
		 LIMIT $3 OFFSET $4`,
		minBalance, maxBalance, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by balance range: %w", err)
	}

	return accounts, nil
}

// GetTopAccountsByBalance retrieves accounts with highest balances
func GetTopAccountsByBalance(limit int) ([]Account, error) {
	accounts, err := postgres.PostgresQuery[Account](
		`SELECT address, balance, first_seen_at
		 FROM accounts
		 ORDER BY balance DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get top accounts: %w", err)
	}

	return accounts, nil
}

// GetAccountTransactions retrieves all transactions for an account
func GetAccountTransactions(address string, limit, offset int) ([]AccountTransaction, error) {
	txs, err := postgres.PostgresQuery[AccountTransaction](
		`SELECT address, txid, block_height, type, balance_change
		 FROM account_transactions
		 WHERE address = $1
		 ORDER BY block_height DESC
		 LIMIT $2 OFFSET $3`,
		address, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account transactions: %w", err)
	}

	return txs, nil
}

// GetAccountTransactionsByType retrieves transactions for an account filtered by type
func GetAccountTransactionsByType(address string, txType string, limit, offset int) ([]AccountTransaction, error) {
	txs, err := postgres.PostgresQuery[AccountTransaction](
		`SELECT address, txid, block_height, type, balance_change
		 FROM account_transactions
		 WHERE address = $1 AND type = $2
		 ORDER BY block_height DESC
		 LIMIT $3 OFFSET $4`,
		address, txType, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account transactions by type: %w", err)
	}

	return txs, nil
}

// GetAccountReceivingTransactions retrieves receiving transactions for an account
func GetAccountReceivingTransactions(address string, limit, offset int) ([]AccountTransaction, error) {
	return GetAccountTransactionsByType(address, string(TxTypeReceive), limit, offset)
}

// GetAccountSendingTransactions retrieves sending transactions for an account
func GetAccountSendingTransactions(address string, limit, offset int) ([]AccountTransaction, error) {
	return GetAccountTransactionsByType(address, string(TxTypeSend), limit, offset)
}

// GetAccountTransactionsByBlockRange retrieves transactions for an account within a block range
func GetAccountTransactionsByBlockRange(address string, fromBlock, toBlock int64, limit, offset int) ([]AccountTransaction, error) {
	txs, err := postgres.PostgresQuery[AccountTransaction](
		`SELECT address, txid, block_height, type, balance_change
		 FROM account_transactions
		 WHERE address = $1 AND block_height >= $2 AND block_height <= $3
		 ORDER BY block_height DESC
		 LIMIT $4 OFFSET $5`,
		address, fromBlock, toBlock, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account transactions by block range: %w", err)
	}

	return txs, nil
}

// GetAccountTransactionCount returns the total number of transactions for an account
func GetAccountTransactionCount(address string) (int64, error) {
	type result struct {
		Count int64 `db:"count"`
	}

	res, err := postgres.PostgresQueryOne[result](
		`SELECT COUNT(*) as count FROM account_transactions WHERE address = $1`,
		address,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get account transaction count: %w", err)
	}

	return res.Count, nil
}

// GetAccountTransaction retrieves a specific transaction for an account
func GetAccountTransaction(address, txid string) (*AccountTransaction, error) {
	tx, err := postgres.PostgresQueryOne[AccountTransaction](
		`SELECT address, txid, block_height, type, balance_change
		 FROM account_transactions
		 WHERE address = $1 AND txid = $2`,
		address, txid,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account transaction: %w", err)
	}

	return tx, nil
}

// GetTransactionAccounts retrieves all accounts associated with a transaction
func GetTransactionAccounts(txid string) ([]AccountTransaction, error) {
	txs, err := postgres.PostgresQuery[AccountTransaction](
		`SELECT address, txid, block_height, type, balance_change
		 FROM account_transactions
		 WHERE txid = $1
		 ORDER BY type, address`,
		txid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction accounts: %w", err)
	}

	return txs, nil
}

// GetRecentActiveAccounts retrieves accounts with recent transaction activity
func GetRecentActiveAccounts(limit int) ([]Account, error) {
	accounts, err := postgres.PostgresQuery[Account](
		`SELECT a.address, a.balance, a.first_seen_at
		 FROM accounts a
		 WHERE a.address IN (
			SELECT DISTINCT ON (address) address
			FROM account_transactions
			ORDER BY address, block_height DESC
		 )
		 ORDER BY (
			SELECT MAX(block_height)
			FROM account_transactions at
			WHERE at.address = a.address
		 ) DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent active accounts: %w", err)
	}

	return accounts, nil
}

// StoreAccountTransaction inserts or updates an account transaction in the database
// If postgresTx is provided, it will be used; otherwise a standalone query is executed
func StoreAccountTransaction(postgresTx DBTX, address string, txid string, blockHeight int64, txType string, balanceChange int64) error {
	ctx := context.Background()

	query := `
		INSERT INTO account_transactions (address, txid, block_height, type, balance_change)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (address, txid) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			type = EXCLUDED.type,
			balance_change = EXCLUDED.balance_change
	`

	if postgresTx == nil {
		postgresTx = postgres.DB
	}

	_, err := postgresTx.Exec(ctx, query, address, txid, blockHeight, txType, balanceChange)
	if err != nil {
		return fmt.Errorf("failed to store account transaction %s for address %s: %w", txid, address, err)
	}

	return nil
}

// CountAccounts returns the total count of accounts
func CountAccounts() (int64, error) {
	var count int64
	err := postgres.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM accounts`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count accounts: %w", err)
	}
	return count, nil
}

// CountAccountTransactions returns the total count of account transactions with optional filters
func CountAccountTransactions(address string, txType string) (int64, error) {
	var query string
	var args []interface{}

	if address != "" && txType != "" {
		query = `SELECT COUNT(*) FROM account_transactions WHERE address = $1 AND type = $2`
		args = []interface{}{address, txType}
	} else if address != "" {
		query = `SELECT COUNT(*) FROM account_transactions WHERE address = $1`
		args = []interface{}{address}
	} else if txType != "" {
		query = `SELECT COUNT(*) FROM account_transactions WHERE type = $1`
		args = []interface{}{txType}
	} else {
		query = `SELECT COUNT(*) FROM account_transactions`
		args = []interface{}{}
	}

	var count int64
	err := postgres.DB.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count account transactions: %w", err)
	}

	return count, nil
}
