package indexer

// This file is kept for backward compatibility
// All types have been moved to internal/types package
// Re-export them here to avoid breaking existing code

import "github.com/keep-starknet-strange/ztarknet/zindex/internal/types"

// Re-export all types from internal/types
type ZcashBlock = types.ZcashBlock
type ChainSupply = types.ChainSupply
type ValuePool = types.ValuePool
type CommitTrees = types.CommitTrees
type TreeInfo = types.TreeInfo
type ZcashTransaction = types.ZcashTransaction
type Vin = types.Vin
type ScriptSig = types.ScriptSig
type Vout = types.Vout
type ScriptPubKey = types.ScriptPubKey
type ShieldedSpend = types.ShieldedSpend
type ShieldedOutput = types.ShieldedOutput
type JoinSplit = types.JoinSplit
type OrchardBundle = types.OrchardBundle
type OrchardAction = types.OrchardAction
