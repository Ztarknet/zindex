package reorg

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// RpcClient interface defines the methods required for reorg detection
type RpcClient interface {
	GetBlockHash(height int64) (string, error)
}

// ReorgError is returned when a reorg is detected and handled
// It signals that indexing should restart from a new height
type ReorgError struct {
	NewStartHeight int64
	ReorgDepth     int
}

func (e *ReorgError) Error() string {
	return fmt.Sprintf("reorg detected: rolled back %d blocks, restart from height %d", e.ReorgDepth, e.NewStartHeight)
}

// IsReorgError checks if an error is a ReorgError
func IsReorgError(err error) bool {
	var reorgErr *ReorgError
	return errors.As(err, &reorgErr)
}

// GetReorgError extracts the ReorgError from an error
func GetReorgError(err error) *ReorgError {
	var reorgErr *ReorgError
	if errors.As(err, &reorgErr) {
		return reorgErr
	}
	return nil
}

// DetectReorg checks if the incoming block's previousblockhash matches our stored hash
// Returns true if a reorg is detected (hashes don't match)
func DetectReorg(incomingBlock *types.ZcashBlock) (bool, error) {
	// Skip reorg detection for genesis block
	if incomingBlock.Height == 0 {
		return false, nil
	}

	prevHeight := incomingBlock.Height - 1

	// Get our stored hash for the previous block
	storedHash, err := postgres.GetBlockHashAtHeight(prevHeight)
	if err != nil {
		// If we don't have the previous block stored, we can't detect a reorg
		// This happens on first run or if there's a gap in our data
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No stored block at height %d, skipping reorg detection", prevHeight)
			return false, nil
		}
		return false, fmt.Errorf("failed to get stored hash at height %d: %w", prevHeight, err)
	}

	// Compare with the incoming block's previous block hash
	if storedHash != incomingBlock.PreviousBlockHash {
		log.Printf("REORG DETECTED: stored hash at height %d is %s, but incoming block claims previous hash is %s",
			prevHeight, storedHash, incomingBlock.PreviousBlockHash)
		return true, nil
	}

	return false, nil
}

// FindCommonAncestor walks back from the given height to find where our chain
// and the node's chain have a common block (same hash at same height)
// Returns the height of the common ancestor, or an error if not found within maxDepth
func FindCommonAncestor(currentHeight int64, rpcClient RpcClient, maxDepth int) (int64, error) {
	log.Printf("Searching for common ancestor from height %d (max depth: %d)", currentHeight, maxDepth)

	for depth := 1; depth <= maxDepth; depth++ {
		checkHeight := currentHeight - int64(depth)

		// Can't go below genesis
		if checkHeight < 0 {
			return 0, fmt.Errorf("reached genesis block without finding common ancestor")
		}

		// Get our stored hash at this height
		storedHash, err := postgres.GetBlockHashAtHeight(checkHeight)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// We don't have this block, so our common ancestor must be even earlier
				// or this is our starting point
				log.Printf("No stored block at height %d, this may be our starting point", checkHeight)
				continue
			}
			return 0, fmt.Errorf("failed to get stored hash at height %d: %w", checkHeight, err)
		}

		// Get the chain's hash at this height
		chainHash, err := rpcClient.GetBlockHash(checkHeight)
		if err != nil {
			return 0, fmt.Errorf("failed to get chain hash at height %d: %w", checkHeight, err)
		}

		// Check if they match
		if storedHash == chainHash {
			log.Printf("Found common ancestor at height %d (hash: %s)", checkHeight, storedHash)
			return checkHeight, nil
		}

		log.Printf("Hash mismatch at height %d: stored=%s, chain=%s", checkHeight, storedHash, chainHash)
	}

	return 0, fmt.Errorf("no common ancestor found within %d blocks - reorg too deep", maxDepth)
}

// HandleReorg orchestrates the full reorg handling process:
// 1. Find the common ancestor
// 2. Rollback the database to that point
// 3. Return the new starting height for re-indexing
func HandleReorg(currentHeight int64, rpcClient RpcClient) (*ReorgError, error) {
	maxDepth := config.Conf.Indexer.MaxReorgDepth
	if maxDepth <= 0 {
		maxDepth = 8 // Default to 8 if not configured
	}

	log.Printf("Handling reorg at height %d with max depth %d", currentHeight, maxDepth)

	// Find the common ancestor
	commonAncestor, err := FindCommonAncestor(currentHeight-1, rpcClient, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to find common ancestor: %w", err)
	}

	reorgDepth := int(currentHeight - 1 - commonAncestor)
	log.Printf("Reorg depth: %d blocks (from height %d to %d)", reorgDepth, currentHeight-1, commonAncestor)

	// Rollback the database
	ctx := context.Background()
	if err := postgres.RollbackToHeight(ctx, commonAncestor); err != nil {
		return nil, fmt.Errorf("failed to rollback to height %d: %w", commonAncestor, err)
	}

	// Return the reorg error with the new start height
	return &ReorgError{
		NewStartHeight: commonAncestor + 1,
		ReorgDepth:     reorgDepth,
	}, nil
}

// CheckAndHandleReorg is a convenience function that combines detection and handling
// Returns nil if no reorg detected, or a ReorgError if reorg was handled
func CheckAndHandleReorg(block *types.ZcashBlock, rpcClient RpcClient) error {
	// Check if reorg handling is enabled
	if !config.Conf.Indexer.EnableReorgHandling {
		return nil
	}

	// Detect reorg
	isReorg, err := DetectReorg(block)
	if err != nil {
		return fmt.Errorf("reorg detection failed: %w", err)
	}

	if !isReorg {
		return nil
	}

	// Handle the reorg
	reorgErr, err := HandleReorg(block.Height, rpcClient)
	if err != nil {
		return fmt.Errorf("reorg handling failed: %w", err)
	}

	return reorgErr
}
