package starks

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// TZE constants for STARK verification
const (
	TzeTypeStarkVerify = 1 // STARK verify extension type
	TzeModeInitialize  = 0 // Initialize mode (creates verifier)
	TzeModeVerify      = 1 // Verify mode (submits proof)
)

// IndexStarks indexes STARK proof data and Ztarknet-specific data from a Zcash block
// This function extracts and stores STARK proofs, verifier data, and Ztarknet facts
// All STARK data in a block are indexed atomically in a single database transaction
func IndexStarks(block *types.ZcashBlock) error {
	// Check if starks module is enabled
	if !config.IsModuleEnabled("STARKS") {
		return nil
	}

	// Count STARK-related TZE transactions in this block
	starkTransactionCount := 0
	for _, tx := range block.Tx {
		if tx.IsTZETransaction() && hasStarkVerifyTze(&tx) {
			starkTransactionCount++
		}
	}

	// If there are no STARK transactions, skip indexing
	if starkTransactionCount == 0 {
		return nil
	}

	log.Printf("Indexing STARK data for block %d (hash: %s, %d STARK transactions)",
		block.Height, block.Hash, starkTransactionCount)

	ctx := context.Background()

	// Begin a database transaction for the entire block
	postgresTx, err := postgres.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin database transaction for block %d: %w", block.Height, err)
	}
	defer postgresTx.Rollback(ctx)

	// Process each transaction in the block
	for _, tx := range block.Tx {
		// Only process TZE transactions with STARK verify
		if !tx.IsTZETransaction() || !hasStarkVerifyTze(&tx) {
			continue
		}

		if err := indexStarkTransaction(postgresTx, block, &tx); err != nil {
			return fmt.Errorf("failed to index STARK transaction %s in block %d: %w",
				tx.TxID, block.Height, err)
		}
	}

	// Commit the transaction
	if err := postgresTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit database transaction for block %d: %w", block.Height, err)
	}

	log.Printf("Successfully indexed %d STARK transactions for block %d", starkTransactionCount, block.Height)
	return nil
}

// hasStarkVerifyTze checks if a transaction has STARK verify TZE inputs or outputs
func hasStarkVerifyTze(tx *types.ZcashTransaction) bool {
	// Check outputs for STARK verify TZE
	for _, vout := range tx.Vout {
		if isStarkVerifyOutput(&vout) {
			return true
		}
	}

	// Check inputs for STARK verify TZE
	for _, vin := range tx.Vin {
		if isStarkVerifyInput(&vin) {
			return true
		}
	}

	return false
}

// isStarkVerifyOutput checks if an output is a STARK verify TZE output
func isStarkVerifyOutput(vout *types.Vout) bool {
	if vout.ScriptPubKey == nil || len(vout.ScriptPubKey.Hex) < 18 {
		return false
	}

	// Check if it starts with "ff" (TZE marker)
	if vout.ScriptPubKey.Hex[:2] != "ff" {
		return false
	}

	// Decode and check if tze_type is 1 (STARK verify)
	scriptBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex)
	if err != nil || len(scriptBytes) < 9 {
		return false
	}

	// Parse tze_type (4 bytes, big-endian, at bytes 1-4)
	tzeType := int32(uint32(scriptBytes[1])<<24 | uint32(scriptBytes[2])<<16 | uint32(scriptBytes[3])<<8 | uint32(scriptBytes[4]))

	return tzeType == TzeTypeStarkVerify
}

// isStarkVerifyInput checks if an input is a STARK verify TZE input
func isStarkVerifyInput(vin *types.Vin) bool {
	if vin.ScriptSig == nil || len(vin.ScriptSig.Hex) < 18 {
		return false
	}

	// Check if it starts with "ff" (TZE marker)
	if vin.ScriptSig.Hex[:2] != "ff" {
		return false
	}

	// Decode and check if tze_type is 1 (STARK verify)
	scriptBytes, err := hex.DecodeString(vin.ScriptSig.Hex)
	if err != nil || len(scriptBytes) < 9 {
		return false
	}

	// Parse tze_type (4 bytes, big-endian, at bytes 1-4)
	tzeType := int32(uint32(scriptBytes[1])<<24 | uint32(scriptBytes[2])<<16 | uint32(scriptBytes[3])<<8 | uint32(scriptBytes[4]))

	return tzeType == TzeTypeStarkVerify
}

// indexStarkTransaction processes a single STARK transaction and its data
func indexStarkTransaction(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction) error {
	// Check if this transaction has STARK verify inputs
	hasStarkInput := false
	for _, vin := range tx.Vin {
		if isStarkVerifyInput(&vin) {
			hasStarkInput = true
			break
		}
	}

	// Process STARK verify inputs first (verify mode - submits proofs)
	for i, vin := range tx.Vin {
		if isStarkVerifyInput(&vin) {
			if err := indexStarkVerifyInput(postgresTx, block, tx, i, &vin); err != nil {
				return fmt.Errorf("failed to index STARK verify input %d: %w", i, err)
			}
		}
	}

	// Process STARK verify outputs
	// If hasStarkInput is false, this is initialize mode (creates new verifiers)
	// If hasStarkInput is true, this is verify mode (updates existing verifier balance)
	for _, vout := range tx.Vout {
		if isStarkVerifyOutput(&vout) {
			if err := indexStarkVerifyOutput(postgresTx, block, tx, &vout, hasStarkInput); err != nil {
				return fmt.Errorf("failed to index STARK verify output %d: %w", vout.N, err)
			}
		}
	}

	return nil
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

// indexStarkVerifyOutput parses and stores a STARK verify output
// If hasStarkInput is false, this is initialize mode (creates new verifier)
// If hasStarkInput is true, this is verify mode (updates existing verifier balance)
func indexStarkVerifyOutput(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction, vout *types.Vout, hasStarkInput bool) error {
	// Parse TZE data from scriptPubKey
	scriptBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex)
	if err != nil {
		return fmt.Errorf("failed to decode scriptPubKey hex: %w", err)
	}

	tzeType, _, precondition, err := parseTzeData(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to parse TZE output data: %w", err)
	}

	// Verify this is STARK verify type
	if tzeType != TzeTypeStarkVerify {
		return fmt.Errorf("expected STARK verify type, got tzeType=%d", tzeType)
	}

	if !hasStarkInput {
		// Initialize mode: Create a new verifier
		// Parse the precondition to get initial state
		starkPrecondition, err := parseStarkVerifyPrecondition(precondition)
		if err != nil {
			return fmt.Errorf("failed to parse STARK precondition: %w", err)
		}

		// Create a verifier ID from the transaction output reference
		verifierID := fmt.Sprintf("%s:%d", tx.TxID, vout.N)

		// TODO: verifier_name and verifier_metadata are not currently defined
		// For now, we'll use placeholder values
		verifierName := fmt.Sprintf("verifier_%s_%d", tx.TxID[:8], vout.N)
		verifierMetadata := ""

		// Store the verifier
		// TODO: Similar to accounts module, balance tracking will need to handle input values being 0
		// See accounts indexing for the TODO note about this issue
		err = StoreVerifier(postgresTx, verifierID, verifierName, verifierMetadata, vout.ValueZat)
		if err != nil {
			return fmt.Errorf("failed to store verifier: %w", err)
		}

		log.Printf("Created verifier %s (initial state: %s) in block %d", verifierID, starkPrecondition.OldState, block.Height)
	} else {
		// Verify mode: Update existing verifier balance
		// We need to find the verifier ID from one of the inputs
		var verifierID string
		for _, vin := range tx.Vin {
			if isStarkVerifyInput(&vin) {
				// Look up the verifier ID from the input
				foundVerifierID, err := getVerifierIDFromInput(postgresTx, &vin)
				if err != nil {
					return fmt.Errorf("failed to get verifier ID from input: %w", err)
				}
				verifierID = foundVerifierID
				break
			}
		}

		if verifierID == "" {
			return fmt.Errorf("no verifier ID found for verify mode output")
		}

		// Update the verifier balance
		err = UpdateVerifierBalance(postgresTx, verifierID, vout.ValueZat)
		if err != nil {
			return fmt.Errorf("failed to update verifier balance: %w", err)
		}

		log.Printf("Updated verifier %s balance to %d in block %d", verifierID, vout.ValueZat, block.Height)
	}

	return nil
}

// indexStarkVerifyInput parses and stores a STARK verify input (verify mode)
// This submits a proof to a verifier
func indexStarkVerifyInput(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction, vin int, input *types.Vin) error {
	// Parse TZE data from scriptSig (witness)
	scriptBytes, err := hex.DecodeString(input.ScriptSig.Hex)
	if err != nil {
		return fmt.Errorf("failed to decode scriptSig hex: %w", err)
	}

	tzeType, _, witness, err := parseTzeData(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to parse TZE input data: %w", err)
	}

	// Verify this is STARK verify type
	// Note: We don't check the mode here because the mode field in inputs
	// doesn't necessarily indicate verify vs initialize - that's determined
	// by whether the transaction has STARK verify inputs (checked earlier)
	if tzeType != TzeTypeStarkVerify {
		return fmt.Errorf("expected STARK verify type, got tzeType=%d", tzeType)
	}

	// Parse the witness to get proof size
	witnessData, err := parseStarkVerifyWitness(witness)
	if err != nil {
		return fmt.Errorf("failed to parse STARK witness: %w", err)
	}

	// Get the verifier ID by tracing back through the chain of verifications
	// The verifier ID is the original txid:vout that created the verifier
	verifierID, err := getVerifierIDFromInput(postgresTx, input)
	if err != nil {
		return fmt.Errorf("failed to get verifier ID: %w", err)
	}

	// Store the STARK proof
	err = StoreStarkProof(postgresTx, verifierID, tx.TxID, block.Height, witnessData.ProofSize)
	if err != nil {
		return fmt.Errorf("failed to store STARK proof: %w", err)
	}

	// If Ztarknet indexing is enabled, parse and store Ztarknet facts
	if ShouldIndexZtarknet() {
		// We need to get the precondition from the TZE output to parse Ztarknet facts
		// The precondition is in the output, and the witness is in the input
		// We need to look up the previous output to get the precondition
		if err := indexZtarknetFacts(postgresTx, block, tx, verifierID, input, witnessData.ProofSize); err != nil {
			return fmt.Errorf("failed to index Ztarknet facts: %w", err)
		}
	}

	log.Printf("Stored STARK proof for verifier %s in tx %s (proof size: %d bytes)", verifierID, tx.TxID, witnessData.ProofSize)

	return nil
}

// indexZtarknetFacts parses and stores Ztarknet-specific facts from a STARK verify transaction
func indexZtarknetFacts(postgresTx DBTX, block *types.ZcashBlock, tx *types.ZcashTransaction, verifierID string, input *types.Vin, proofSize int64) error {
	// Find the corresponding TZE output in this transaction to get the new state
	// The output will have the new state in its precondition
	var newStatePrecondition []byte
	found := false

	for _, vout := range tx.Vout {
		if isStarkVerifyOutput(&vout) {
			scriptBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex)
			if err != nil {
				continue
			}

			_, _, precondition, err := parseTzeData(scriptBytes)
			if err != nil {
				continue
			}

			newStatePrecondition = precondition
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no STARK verify output found in transaction for Ztarknet facts")
	}

	// Parse the new state from the output precondition
	newStateData, err := parseStarkVerifyPrecondition(newStatePrecondition)
	if err != nil {
		return fmt.Errorf("failed to parse new state precondition: %w", err)
	}

	// Parse the old state from the input scriptSig
	// The old state is encoded in the TZE input script
	scriptBytes, err := hex.DecodeString(input.ScriptSig.Hex)
	if err != nil {
		return fmt.Errorf("failed to decode scriptSig hex: %w", err)
	}

	_, _, witness, err := parseTzeData(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to parse TZE input data for old state: %w", err)
	}

	// The old state needs to be retrieved from the previous output's precondition
	// For now, we'll use a placeholder since we need to query the previous output
	// TODO: Query the previous TZE output to get its precondition and parse old_state
	oldState := "0000000000000000000000000000000000000000000000000000000000000000" // Placeholder

	// Parse witness to ensure we have the proof data (already done in caller, but we need it here too)
	_, err = parseStarkVerifyWitness(witness)
	if err != nil {
		return fmt.Errorf("failed to parse witness for Ztarknet facts: %w", err)
	}

	// Store the Ztarknet facts
	err = StoreZtarknetFacts(
		postgresTx,
		verifierID,
		tx.TxID,
		block.Height,
		proofSize,
		oldState,
		newStateData.NewState,
		newStateData.ProgramHash,
		newStateData.InnerProgramHash,
	)
	if err != nil {
		return fmt.Errorf("failed to store Ztarknet facts: %w", err)
	}

	log.Printf("Stored Ztarknet facts for verifier %s: %s -> %s", verifierID, oldState[:8], newStateData.NewState[:8])

	return nil
}

// getVerifierIDFromInput traces back through the chain of verifications to find the original verifier ID
// It queries the database to find either:
// 1. A verifier with verifier_id matching the previous txid:vout (if this is the first verification)
// 2. A stark_proof with txid matching the previous txid, then gets its verifier_id (if this is a subsequent verification)
func getVerifierIDFromInput(postgresTx DBTX, input *types.Vin) (string, error) {
	ctx := context.Background()
	prevTxOutputRef := fmt.Sprintf("%s:%d", input.TxID, input.Vout)

	// First, try to find a verifier with this exact ID (initialize case)
	var verifierID string
	err := postgresTx.QueryRow(ctx,
		`SELECT verifier_id FROM verifiers WHERE verifier_id = $1`,
		prevTxOutputRef,
	).Scan(&verifierID)

	if err == nil {
		// Found a verifier directly - this is the original verifier
		return verifierID, nil
	}

	// If not found as a verifier, look for a stark_proof with this txid
	// This means the previous transaction was also a verification, so we need to get its verifier_id
	err = postgresTx.QueryRow(ctx,
		`SELECT verifier_id FROM stark_proofs WHERE txid = $1 LIMIT 1`,
		input.TxID,
	).Scan(&verifierID)

	if err == nil {
		// Found a proof - return its verifier_id
		return verifierID, nil
	}

	// If we still haven't found it, this is an error
	return "", fmt.Errorf("could not find verifier ID for input %s:%d", input.TxID, input.Vout)
}

// StarkPreconditionData represents parsed STARK precondition data
type StarkPreconditionData struct {
	OldState         string
	NewState         string
	ProgramHash      string // bootloader program hash
	InnerProgramHash string // OS program hash
}

// parseStarkVerifyPrecondition parses the precondition data from a STARK verify TZE output
// Format (from JavaScript reference):
// - Skip first 4 bytes (metadata/flags) = 8 hex chars
// - root: 32 bytes (64 hex chars)
// - osProgramHash: 32 bytes (64 hex chars)
// - bootloaderProgramHash: 32 bytes (64 hex chars)
func parseStarkVerifyPrecondition(precondition []byte) (*StarkPreconditionData, error) {
	// Convert to hex string for easier parsing
	hexData := hex.EncodeToString(precondition)

	// Skip first 4 bytes (8 hex chars)
	offset := 8

	// Expected length: 96 bytes (192 hex chars) = root (32) + os_program_hash (32) + bootloader_program_hash (32)
	expectedLength := offset + 192

	// Right-pad with '0' characters (not null bytes) if needed
	if len(hexData) < expectedLength {
		paddingNeeded := expectedLength - len(hexData)
		padding := make([]byte, paddingNeeded)
		for i := 0; i < paddingNeeded; i++ {
			padding[i] = '0'
		}
		hexData = hexData + string(padding)
	}

	// Parse the fields
	oldState := hexData[offset : offset+64]
	osProgramHash := hexData[offset+64 : offset+128]
	bootloaderProgramHash := hexData[offset+128 : offset+192]

	return &StarkPreconditionData{
		OldState:         oldState,
		NewState:         oldState, // For outputs, the old state is the current state (will become new state after proof)
		ProgramHash:      bootloaderProgramHash,
		InnerProgramHash: osProgramHash,
	}, nil
}

// StarkWitnessData represents parsed STARK witness data
type StarkWitnessData struct {
	WithPedersen bool
	ProofFormat  string
	ProofData    []byte
	ProofSize    int64
}

// parseStarkVerifyWitness parses the witness data from a STARK verify TZE input
// Format (from JavaScript reference):
// - 1 byte with_pedersen
// - 1 byte proof_format (0=JSON, 1=Binary)
// - variable proof_data
func parseStarkVerifyWitness(witness []byte) (*StarkWitnessData, error) {
	if len(witness) < 2 {
		return nil, fmt.Errorf("witness too short: %d bytes", len(witness))
	}

	withPedersen := witness[0] == 1
	proofFormat := "Binary"
	if witness[1] == 0 {
		proofFormat = "JSON"
	}

	var proofData []byte
	if len(witness) > 2 {
		proofData = witness[2:]
	} else {
		proofData = []byte{}
	}

	return &StarkWitnessData{
		WithPedersen: withPedersen,
		ProofFormat:  proofFormat,
		ProofData:    proofData,
		ProofSize:    int64(len(proofData)),
	}, nil
}
