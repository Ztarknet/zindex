package blocks

import (
	"log"
)

// ZcashBlock is a forward declaration to avoid import cycles
// The actual type is defined in internal/indexer/types.go
type ZcashBlock interface{}

// IndexBlocks indexes core block data
// This function stores essential block information and is always executed (core module)
func IndexBlocks(block ZcashBlock) error {
	// Note: Blocks module is a core module and is always enabled
	// No need to check if it's enabled in config

	// TODO: Implement blocks indexing logic
	// This should:
	// 1. Store block metadata (hash, height, time, difficulty, etc.)
	// 2. Store block relationships (previous/next block hashes)
	// 3. Store merkle roots and commitment tree roots
	// 4. Handle reorg detection by comparing previous block hash
	//
	// Reference the ZcashBlock type for accessing block data:
	//   - block.Hash, block.Height, block.Time
	//   - block.PreviousBlockHash, block.NextBlockHash
	//   - block.MerkleRoot, block.FinalSaplingRoot, block.FinalOrchardRoot
	//   - block.Difficulty, block.Size

	log.Printf("Blocks indexing called for block (not yet implemented)")
	return nil
}
