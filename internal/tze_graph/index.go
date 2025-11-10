package tze_graph

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// IndexTzeGraph indexes TZE (Transparent Zcash Extension) graph data from a Zcash block
// This function tracks TZE inputs, outputs, and their relationships
// All TZE transactions in a block are indexed atomically in a single database transaction
func IndexTzeGraph(block *types.ZcashBlock) error {
	// Check if tze_graph module is enabled
	if !config.IsModuleEnabled("TZE_GRAPH") {
		return nil
	}

	// Count TZE transactions in this block
	tzeTransactionCount := 0
	for _, tx := range block.Tx {
		if tx.IsTZETransaction() {
			tzeTransactionCount++
		}
	}

	// If there are no TZE transactions, skip indexing
	if tzeTransactionCount == 0 {
		return nil
	}

	log.Printf("Indexing TZE graph for block %d (hash: %s, %d TZE transactions)",
		block.Height, block.Hash, tzeTransactionCount)

	ctx := context.Background()

	// Begin a database transaction for the entire block
	postgresTx, err := postgres.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin database transaction for block %d: %w", block.Height, err)
	}
	defer postgresTx.Rollback(ctx)

	// Process each transaction in the block
	for _, tx := range block.Tx {
		// Only process TZE transactions
		if !tx.IsTZETransaction() {
			continue
		}

		if err := indexTzeTransaction(postgresTx, block, &tx); err != nil {
			return fmt.Errorf("failed to index TZE transaction %s in block %d: %w",
				tx.TxID, block.Height, err)
		}
	}

	// Commit the transaction
	if err := postgresTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit database transaction for block %d: %w", block.Height, err)
	}

	log.Printf("Successfully indexed %d TZE transactions for block %d", tzeTransactionCount, block.Height)
	return nil
}

// indexTzeTransaction processes a single TZE transaction and its inputs/outputs
func indexTzeTransaction(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction) error {
	// Process TZE outputs first
	for _, vout := range tx.Vout {
		if isTzeOutput(&vout) {
			if err := indexTzeOutput(postgresTx, tx.TxID, &vout); err != nil {
				return fmt.Errorf("failed to index TZE output %d: %w", vout.N, err)
			}
		}
	}

	// Process TZE inputs
	for i, vin := range tx.Vin {
		if isTzeInput(&vin) {
			if err := indexTzeInput(postgresTx, tx.TxID, i, &vin, block.Height); err != nil {
				return fmt.Errorf("failed to index TZE input %d: %w", i, err)
			}
		}
	}

	return nil
}

// isTzeOutput checks if an output is a TZE output
func isTzeOutput(vout *types.Vout) bool {
	return vout.ScriptPubKey != nil && len(vout.ScriptPubKey.Hex) >= 2 &&
		vout.ScriptPubKey.Hex[:2] == "ff"
}

// isTzeInput checks if an input is a TZE input
func isTzeInput(vin *types.Vin) bool {
	return vin.ScriptSig != nil && len(vin.ScriptSig.Hex) >= 2 &&
		vin.ScriptSig.Hex[:2] == "ff"
}

// parseTzeData extracts TZE extension_id, mode, and data from a script byte array
// Format: ff <extension_id> <mode> <data>
// where extension_id and mode are 4 bytes each (big-endian)
func parseTzeData(scriptBytes []byte) (tzeType int32, tzeMode int32, data []byte, err error) {
	// Minimum size check: 0xff (1 byte) + extension_id (4 bytes) + mode (4 bytes) = 9 bytes
	if len(scriptBytes) < 9 {
		return 0, 0, nil, fmt.Errorf("TZE script too short: %d bytes", len(scriptBytes))
	}

	// scriptBytes[0] should be 0xff
	if scriptBytes[0] != 0xff {
		return 0, 0, nil, fmt.Errorf("TZE script does not start with 0xff marker")
	}

	// extension_id is 4 bytes (big-endian)
	tzeType = int32(uint32(scriptBytes[1])<<24 | uint32(scriptBytes[2])<<16 | uint32(scriptBytes[3])<<8 | uint32(scriptBytes[4]))
	// mode is 4 bytes (big-endian)
	tzeMode = int32(uint32(scriptBytes[5])<<24 | uint32(scriptBytes[6])<<16 | uint32(scriptBytes[7])<<8 | uint32(scriptBytes[8]))

	// Extract data - everything after the 9-byte header
	if len(scriptBytes) > 9 {
		data = scriptBytes[9:]
	} else {
		data = []byte{}
	}

	return tzeType, tzeMode, data, nil
}

// indexTzeOutput parses and stores a TZE output
func indexTzeOutput(postgresTx DBTX, txid string, vout *types.Vout) error {
	// Parse TZE data from scriptPubKey
	scriptHex := vout.ScriptPubKey.Hex

	// Decode the hex string
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return fmt.Errorf("failed to decode scriptPubKey hex: %w", err)
	}

	// Parse TZE fields
	tzeType, tzeMode, precondition, err := parseTzeData(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to parse TZE output data: %w", err)
	}

	// Store the TZE output
	err = StoreTzeOutput(
		postgresTx,
		txid,
		int(vout.N),
		vout.ValueZat,
		tzeType,
		tzeMode,
		precondition,
	)
	if err != nil {
		return fmt.Errorf("failed to store TZE output: %w", err)
	}

	return nil
}

// indexTzeInput parses and stores a TZE input
func indexTzeInput(postgresTx DBTX, txid string, vin int, input *types.Vin, blockHeight int64) error {
	// Parse TZE data from scriptSig
	scriptHex := input.ScriptSig.Hex

	// Decode the hex string
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return fmt.Errorf("failed to decode scriptSig hex: %w", err)
	}

	// Parse TZE fields (witness data is parsed but not currently stored)
	tzeType, tzeMode, _, err := parseTzeData(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to parse TZE input data: %w", err)
	}

	// Get the previous output information
	prevTxid := input.TxID
	prevVout := int(input.Vout)

	// TODO: Query the previous TZE output to get its value
	// This requires: SELECT value FROM tze_outputs WHERE txid = prevTxid AND vout = prevVout
	value := int64(0)

	// Store the TZE input
	err = StoreTzeInput(
		postgresTx,
		txid,
		vin,
		value,
		prevTxid,
		prevVout,
		tzeType,
		tzeMode,
		blockHeight,
	)
	if err != nil {
		return fmt.Errorf("failed to store TZE input: %w", err)
	}

	return nil
}
