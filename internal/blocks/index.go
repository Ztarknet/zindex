package blocks

import (
	"fmt"
	"log"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/types"
)

// IndexBlocks indexes core block data
// This function stores essential block information and is always executed (core module)
func IndexBlocks(block *types.ZcashBlock) error {
	// Note: Blocks module is a core module and is always enabled
	// No need to check if it's enabled in config

	log.Printf("Indexing block data %d (hash: %s)", block.Height, block.Hash)

	// Store the block using the existing storage function
	err := StoreBlock(
		block.Height,
		block.Hash,
		block.PreviousBlockHash,
		block.MerkleRoot,
		block.Time,
		block.Difficulty,
		block.Nonce,
		block.Version,
		len(block.Tx),
	)
	if err != nil {
		return fmt.Errorf("failed to store block %d: %w", block.Height, err)
	}

	log.Printf("Successfully indexed block data %d", block.Height)
	return nil
}
