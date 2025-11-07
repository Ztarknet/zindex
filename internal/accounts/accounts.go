package accounts

import (
	"database/sql"
	"fmt"

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

	_, err := postgres.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create accounts schema: %w", err)
	}

	return nil
}

func GetAccount(address string) (*Account, error) {
	var account Account
	err := postgres.DB.QueryRow(
		`SELECT address, type, balance, tx_count, first_seen_at, last_activity_at
		 FROM accounts WHERE address = $1`,
		address,
	).Scan(&account.Address, &account.Type, &account.Balance, &account.TxCount, &account.FirstSeenAt, &account.LastActivityAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

func GetAccounts(limit, offset int) ([]Account, error) {
	rows, err := postgres.DB.Query(
		`SELECT address, type, balance, tx_count, first_seen_at, last_activity_at
		 FROM accounts ORDER BY last_activity_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.Address, &account.Type, &account.Balance, &account.TxCount, &account.FirstSeenAt, &account.LastActivityAt); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
