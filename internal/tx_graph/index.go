package tx_graph

import (
	"context"
	"fmt"
	"log"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// IndexTxGraph indexes transaction graph data from a Zcash block
// This function builds the UTXO graph by tracking transaction inputs and outputs
// All transactions in a block are indexed atomically in a single database transaction
func IndexTxGraph(block *types.ZcashBlock) error {
	// Check if tx_graph module is enabled
	if !config.IsModuleEnabled("TX_GRAPH") {
		return nil
	}

	log.Printf("Indexing transaction graph for block %d (hash: %s, %d transactions)",
		block.Height, block.Hash, len(block.Tx))

	ctx := context.Background()

	// Begin a database transaction for the entire block
	postgresTx, err := postgres.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin database transaction for block %d: %w", block.Height, err)
	}
	defer postgresTx.Rollback(ctx)

	// Process each transaction in the block
	for _, tx := range block.Tx {
		if err := indexTransaction(postgresTx, block, &tx); err != nil {
			return fmt.Errorf("failed to index transaction %s in block %d: %w",
				tx.TxID, block.Height, err)
		}
	}

	// Commit the transaction
	if err := postgresTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit database transaction for block %d: %w", block.Height, err)
	}

	log.Printf("Successfully indexed %d transactions for block %d", len(block.Tx), block.Height)
	return nil
}

// indexTransaction processes a single transaction and its inputs/outputs
func indexTransaction(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction) error {
	// Determine transaction type
	txType := determineTransactionType(tx)

	// Calculate total output value
	totalOutput := calculateTotalOutput(tx)

	// TODO: Calculate input values and fees by querying previous outputs from the database
	// This requires fetching the value of each previous output referenced by inputs.
	// Fields affected:
	//   - transactions.total_fee (set to 0 for now)
	//   - transaction_inputs.value (set to 0 for now)

	// Store the transaction
	err := StoreTransaction(
		postgresTx,
		tx.TxID,
		block.Height,
		block.Hash,
		tx.Version,
		int64(tx.LockTime),
		string(txType),
		totalOutput,
		0, // TODO: totalFee - requires calculating total_input - total_output
		tx.Size,
	)
	if err != nil {
		return fmt.Errorf("failed to store transaction: %w", err)
	}

	// Store transaction outputs
	for _, vout := range tx.Vout {
		err := StoreTransactionOutput(
			postgresTx,
			tx.TxID,
			int(vout.N),
			vout.ValueZat,
		)
		if err != nil {
			return fmt.Errorf("failed to store output %d: %w", vout.N, err)
		}
	}

	// Store transaction inputs (skip for coinbase transactions)
	if !tx.IsCoinbase() {
		for i, vin := range tx.Vin {
			// Skip if this is a coinbase input (shouldn't happen here, but be safe)
			if vin.Coinbase != "" {
				continue
			}

			// TODO: Query the previous output to get its value
			// This requires: SELECT value FROM transaction_outputs WHERE txid = vin.TxID AND vout = vin.Vout
			value := int64(0)

			err := StoreTransactionInput(
				postgresTx,
				tx.TxID,
				i,
				value,
				vin.TxID,
				int(vin.Vout),
				int64(vin.Sequence),
				block.Height,
			)
			if err != nil {
				return fmt.Errorf("failed to store input %d: %w", i, err)
			}
		}
	}

	return nil
}

// determineTransactionType determines the type of a transaction based on its properties
func determineTransactionType(tx *types.ZcashTransaction) TransactionType {
	// Check for coinbase
	if tx.IsCoinbase() {
		return TxTypeCoinbase
	}

	// Check for TZE
	if tx.IsTZETransaction() {
		return TxTypeTZE
	}

	// Determine based on shielded components
	hasTransparentInput := len(tx.Vin) > 0 && tx.Vin[0].Coinbase == ""
	hasTransparentOutput := len(tx.Vout) > 0
	hasShieldedInput := len(tx.VShieldedSpend) > 0 || len(tx.VJoinSplit) > 0 ||
		(tx.Orchard != nil && len(tx.Orchard.Actions) > 0)
	hasShieldedOutput := len(tx.VShieldedOutput) > 0 || len(tx.VJoinSplit) > 0 ||
		(tx.Orchard != nil && len(tx.Orchard.Actions) > 0)

	// Classify based on input/output types
	if hasTransparentInput && hasTransparentOutput && !hasShieldedInput && !hasShieldedOutput {
		return TxTypeT2T
	} else if hasTransparentInput && hasShieldedOutput {
		return TxTypeT2Z
	} else if hasShieldedInput && hasTransparentOutput {
		return TxTypeZ2T
	} else if hasShieldedInput && hasShieldedOutput {
		return TxTypeZ2Z
	}

	// Default to t2t if we can't determine
	return TxTypeT2T
}

// calculateTotalOutput sums up all transparent outputs in a transaction
func calculateTotalOutput(tx *types.ZcashTransaction) int64 {
	total := int64(0)
	for _, vout := range tx.Vout {
		total += vout.ValueZat
	}
	return total
}
