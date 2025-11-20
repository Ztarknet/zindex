package accounts

import (
	"context"
	"fmt"
	"log"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// IndexAccounts indexes account-related data from a Zcash block
// This function extracts and stores account balances, transactions, and related data
// All account updates in a block are indexed atomically in a single database transaction
func IndexAccounts(block *types.ZcashBlock) error {
	// Check if accounts module is enabled
	if !config.IsModuleEnabled("ACCOUNTS") {
		return nil
	}

	log.Printf("Indexing accounts for block %d (hash: %s, %d transactions)",
		block.Height, block.Hash, len(block.Tx))

	ctx := context.Background()

	// Begin a database transaction for the entire block
	postgresTx, err := postgres.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin database transaction for block %d: %w", block.Height, err)
	}
	defer postgresTx.Rollback(ctx)

	// Track balance changes for each address in this block
	balanceChanges := make(map[string]int64)

	// Process each transaction in the block
	for _, tx := range block.Tx {
		if err := indexAccountTransaction(postgresTx, block, &tx, balanceChanges); err != nil {
			return fmt.Errorf("failed to index accounts for transaction %s in block %d: %w",
				tx.TxID, block.Height, err)
		}
	}

	// Update account balances first (this creates accounts if they don't exist)
	for address, change := range balanceChanges {
		if err := updateAccountBalance(postgresTx, address, change); err != nil {
			return fmt.Errorf("failed to update balance for account %s in block %d: %w",
				address, block.Height, err)
		}
	}

	// Now store account transactions (accounts exist now, so FK constraint satisfied)
	for _, tx := range block.Tx {
		if err := storeAccountTransactionsForTx(postgresTx, block, &tx); err != nil {
			return fmt.Errorf("failed to store account transactions for tx %s in block %d: %w",
				tx.TxID, block.Height, err)
		}
	}

	// Commit the transaction
	if err := postgresTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit database transaction for block %d: %w", block.Height, err)
	}

	log.Printf("Successfully indexed accounts for block %d (%d addresses affected)",
		block.Height, len(balanceChanges))
	return nil
}

// indexAccountTransaction processes a single transaction and tracks balance changes
// Note: This does NOT store account_transactions - that happens after accounts are created
func indexAccountTransaction(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction, balanceChanges map[string]int64) error {
	// Skip coinbase transactions for input processing (they have no spender address)
	if !tx.IsCoinbase() {
		// TODO: Process inputs - these represent sending transactions (spending from addresses)
		// Note: In Zcash, transparent inputs (Vin) don't directly contain the sender's address.
		// The address must be looked up from the previous output (prevTxid:prevVout) being spent.
		//
		// Recommended implementation approach:
		// 1. Query transaction_outputs table from tx_graph module to get the scriptPubKey/addresses
		//    for each (vin.TxID, vin.Vout) pair
		// 2. For each sender address found:
		//    - Store account transaction with type = "send"
		//    - Subtract vout.ValueZat from balanceChanges[address]
		// 3. This requires cross-module data access (accounts depends on tx_graph)
		//    - Consider indexing order: ensure tx_graph runs before accounts
		//    - Or implement a post-processing step after both modules complete
	}

	// Process outputs - track balance changes for receiving transactions
	for _, vout := range tx.Vout {
		// Extract addresses from the output script
		if vout.ScriptPubKey != nil && len(vout.ScriptPubKey.Addresses) > 0 {
			for _, address := range vout.ScriptPubKey.Addresses {
				// Track balance change (positive for receiving)
				balanceChanges[address] += vout.ValueZat
			}
		}
	}

	return nil
}

// storeAccountTransactionsForTx stores account transaction records for a single transaction
// This should be called AFTER accounts are created to satisfy foreign key constraints
func storeAccountTransactionsForTx(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction) error {
	// Process outputs - record receiving transactions
	for _, vout := range tx.Vout {
		if vout.ScriptPubKey != nil && len(vout.ScriptPubKey.Addresses) > 0 {
			for _, address := range vout.ScriptPubKey.Addresses {
				err := StoreAccountTransaction(
					postgresTx,
					address,
					tx.TxID,
					block.Height,
					string(TxTypeReceive),
					int64(vout.Value), // positive value for receiving
				)
				if err != nil {
					return fmt.Errorf("failed to store receiving transaction for address %s: %w", address, err)
				}
			}
		}
	}

	return nil
}

// updateAccountBalance updates or creates an account with the balance change
func updateAccountBalance(postgresTx DBTX, address string, change int64) error {
	ctx := context.Background()

	// Use INSERT ... ON CONFLICT to either create or update the account
	query := `
		INSERT INTO accounts (address, balance)
		VALUES ($1, $2)
		ON CONFLICT (address) DO UPDATE SET
			balance = accounts.balance + EXCLUDED.balance
	`

	_, err := postgresTx.Exec(ctx, query, address, change)
	if err != nil {
		return fmt.Errorf("failed to update account balance for %s: %w", address, err)
	}

	return nil
}
