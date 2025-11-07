package accounts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

func InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS accounts (
			address VARCHAR(95) PRIMARY KEY,
			type VARCHAR(20) NOT NULL,
			balance BIGINT DEFAULT 0,
			tx_count BIGINT DEFAULT 0,
			first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);
		CREATE INDEX IF NOT EXISTS idx_accounts_balance ON accounts(balance);
	`

	_, err := postgres.DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create accounts schema: %w", err)
	}

	return nil
}

func GetAccount(address string) (*Account, error) {
	account, err := postgres.PostgresQueryOne[Account](
		`SELECT address, type, balance, tx_count, first_seen_at, last_activity_at
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

func GetAccounts(limit, offset int) ([]Account, error) {
	accounts, err := postgres.PostgresQuery[Account](
		`SELECT address, type, balance, tx_count, first_seen_at, last_activity_at
		 FROM accounts ORDER BY last_activity_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}

	return accounts, nil
}
