package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/accounts"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/blocks"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/reorg"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/starks"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tx_graph"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tze_graph"
)

// RpcClient interface defines the methods required to fetch block data from RPC
type RpcClient interface {
	GetBlockHash(height int64) (string, error)
	GetBlock(hash string) (map[string]interface{}, error)
	GetBlockCount() (int64, error)
}

const (
	// maxIndexRetries is the maximum number of times to retry indexing a block after rollback
	maxIndexRetries = 3
)

var (
	stopChan     chan struct{}
	errorChannel chan error
)

// IndexBlock fetches and indexes a single block at the specified height
// This is the main entry point for indexing a block and coordinates all module indexing
func IndexBlock(height int64, rpcClient RpcClient) error {
	log.Printf("Indexing block at height %d", height)

	// Fetch block hash
	blockHash, err := rpcClient.GetBlockHash(height)
	if err != nil {
		return fmt.Errorf("failed to get block hash for height %d: %w", height, err)
	}

	// Fetch block data
	rawBlock, err := rpcClient.GetBlock(blockHash)
	if err != nil {
		return fmt.Errorf("failed to get block %s: %w", blockHash, err)
	}

	// Parse block into ZcashBlock structure
	block, err := parseBlock(rawBlock)
	if err != nil {
		return fmt.Errorf("failed to parse block %d: %w", height, err)
	}

	// Verify block height matches expected height
	if block.Height != height {
		return fmt.Errorf("block height mismatch: expected %d, got %d", height, block.Height)
	}

	// Reorg detection and handling
	// If enabled, compare block's previousblockhash with stored hash at height-1
	// If mismatch detected, rollback to common ancestor and return ReorgError
	if err := reorg.CheckAndHandleReorg(block, rpcClient); err != nil {
		return err // This may be a ReorgError which will be handled by the indexing loop
	}

	// Index block data in each enabled module
	// Order matters: blocks should be indexed first, then modules that depend on blocks
	if err := indexModules(block); err != nil {
		return fmt.Errorf("failed to index modules for block %d: %w", height, err)
	}

	// Update indexer state with the new last indexed block
	if err := postgres.UpdateLastIndexedBlock(height, blockHash); err != nil {
		return fmt.Errorf("failed to update last indexed block: %w", err)
	}

	log.Printf("Successfully indexed block %d: %s", height, blockHash)
	return nil
}

// parseBlock converts the raw RPC response map into a strongly-typed ZcashBlock structure
func parseBlock(rawBlock map[string]interface{}) (*types.ZcashBlock, error) {
	// Marshal the map back to JSON
	jsonData, err := json.Marshal(rawBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal block data: %w", err)
	}

	// Unmarshal into our ZcashBlock type
	var block types.ZcashBlock
	if err := json.Unmarshal(jsonData, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block data: %w", err)
	}

	return &block, nil
}

// indexModules calls the indexing function for each enabled module
// This function orchestrates the indexing across all modules
func indexModules(block *types.ZcashBlock) error {
	// Always index blocks (core module)
	if err := blocks.IndexBlocks(block); err != nil {
		return fmt.Errorf("failed to index blocks module: %w", err)
	}

	// Index accounts module (if enabled)
	if err := accounts.IndexAccounts(block); err != nil {
		return fmt.Errorf("failed to index accounts module: %w", err)
	}

	// Index transaction graph module (if enabled)
	if err := tx_graph.IndexTxGraph(block); err != nil {
		return fmt.Errorf("failed to index tx_graph module: %w", err)
	}

	// Index TZE graph module (if enabled)
	if err := tze_graph.IndexTzeGraph(block); err != nil {
		return fmt.Errorf("failed to index tze_graph module: %w", err)
	}

	// Index STARK module (if enabled)
	// This includes both STARK proofs and Ztarknet-specific data
	if err := starks.IndexStarks(block); err != nil {
		return fmt.Errorf("failed to index starks module: %w", err)
	}

	return nil
}

// GetLastIndexedBlock retrieves the last successfully indexed block height
// This is used to resume indexing from the correct position
func GetLastIndexedBlock() (int64, error) {
	return postgres.GetLastIndexedBlock()
}

// Start begins the indexing process from the specified start block
// If startBlock is -1, it will resume from the last indexed block
// This function runs in a goroutine and returns channels for stopping and error reporting
func Start(startBlock int64, rpcClient RpcClient) (chan struct{}, chan error) {
	stopChan = make(chan struct{})
	errorChannel = make(chan error, 1)

	// Determine starting block height
	var indexStartBlock int64
	if startBlock >= 0 {
		indexStartBlock = startBlock
		log.Printf("Starting indexer from specified block: %d", startBlock)
	} else {
		lastBlock, err := GetLastIndexedBlock()
		if err != nil {
			log.Printf("Failed to get last indexed block, starting from config: %v", err)
			indexStartBlock = config.Conf.Indexer.StartBlock
		} else {
			indexStartBlock = lastBlock + 1
			log.Printf("Resuming indexer from block: %d", indexStartBlock)
		}
	}

	// Start indexing loop in goroutine
	go startIndexingLoop(indexStartBlock, rpcClient)

	return stopChan, errorChannel
}

// Stop signals the indexing loop to stop
func Stop() {
	if stopChan != nil {
		log.Println("Stopping indexer...")
		close(stopChan)
	}
}

// startIndexingLoop is the main indexing loop that continuously processes blocks
func startIndexingLoop(startBlock int64, rpcClient RpcClient) {
	currentBlock := startBlock
	pollInterval := time.Duration(config.Conf.Indexer.PollInterval) * time.Second
	retryCount := 0 // Track retries for the current block

	log.Printf("Starting indexing loop from block %d", currentBlock)

	for {
		select {
		case <-stopChan:
			log.Println("Indexing stopped")
			return
		default:
			// Get current blockchain height
			blockCount, err := rpcClient.GetBlockCount()
			if err != nil {
				log.Printf("Failed to get block count: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			// Wait if we're caught up
			if currentBlock > blockCount {
				time.Sleep(pollInterval)
				continue
			}

			// Calculate batch end
			batchEnd := currentBlock + int64(config.Conf.Indexer.BatchSize)
			if batchEnd > blockCount {
				batchEnd = blockCount
			}

			log.Printf("Indexing blocks %d to %d (chain height: %d)", currentBlock, batchEnd, blockCount)

			// Index batch of blocks
			for height := currentBlock; height <= batchEnd; height++ {
				select {
				case <-stopChan:
					return
				default:
					if err := IndexBlock(height, rpcClient); err != nil {
						// Check if this is a reorg error - if so, restart from the new height
						if reorgErr := reorg.GetReorgError(err); reorgErr != nil {
							log.Printf("Reorg handled: %s", reorgErr.Error())
							currentBlock = reorgErr.NewStartHeight
							retryCount = 0 // Reset retry count after reorg
							break          // Exit the inner loop to restart from new height
						}

						// Non-reorg error - attempt rollback and retry
						log.Printf("Error indexing block %d: %v", height, err)
						retryCount++

						if retryCount > maxIndexRetries {
							log.Printf("Max retries (%d) exceeded for block %d, stopping indexer", maxIndexRetries, height)
							errorChannel <- fmt.Errorf("max retries exceeded for block %d: %w", height, err)
							return
						}

						// Rollback to previous block and retry
						rollbackHeight := height - 1
						if rollbackHeight < 0 {
							rollbackHeight = 0
						}

						log.Printf("Rolling back to block %d and retrying (attempt %d/%d)", rollbackHeight, retryCount, maxIndexRetries)

						ctx := context.Background()
						if rollbackErr := postgres.RollbackToHeight(ctx, rollbackHeight); rollbackErr != nil {
							log.Printf("Failed to rollback to height %d: %v", rollbackHeight, rollbackErr)
							errorChannel <- fmt.Errorf("failed to rollback after indexing error: %w", rollbackErr)
							return
						}

						// Set current block to retry from the rollback height + 1
						currentBlock = rollbackHeight + 1
						break // Exit inner loop to restart from the rollback point
					}

					// Success - reset retry count
					retryCount = 0
				}
			}

			// Only advance to next batch if we completed the current one without reorg or error
			if currentBlock <= batchEnd {
				currentBlock = batchEnd + 1
			}

			// Sleep if we're caught up
			if currentBlock > blockCount {
				time.Sleep(pollInterval)
			}
		}
	}
}
