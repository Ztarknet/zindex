package indexer

// ZcashBlock represents a complete Zcash block with all transactions and metadata
type ZcashBlock struct {
	// Block identification and metadata
	Hash          string `json:"hash"`
	Confirmations int64  `json:"confirmations"`
	Size          int64  `json:"size"`
	Height        int64  `json:"height"`
	Version       int    `json:"version"`

	// Merkle and commitment roots
	MerkleRoot       string `json:"merkleroot"`
	BlockCommitments string `json:"blockcommitments"`
	FinalSaplingRoot string `json:"finalsaplingroot"`
	FinalOrchardRoot string `json:"finalorchardroot,omitempty"`

	// Transactions
	Tx []ZcashTransaction `json:"tx"`

	// Timing and difficulty
	Time       int64   `json:"time"`
	Nonce      string  `json:"nonce,omitempty"`
	Bits       string  `json:"bits"`
	Difficulty float64 `json:"difficulty"`

	// Chain links
	PreviousBlockHash string `json:"previousblockhash,omitempty"`
	NextBlockHash     string `json:"nextblockhash,omitempty"`

	// Chain supply and value pools (optional, depends on node config)
	ChainSupply *ChainSupply `json:"chainSupply,omitempty"`
	ValuePools  []ValuePool  `json:"valuePools,omitempty"`
	Trees       *CommitTrees `json:"trees,omitempty"`
}

// ChainSupply represents the total chain supply information
type ChainSupply struct {
	Monitored     bool    `json:"monitored"`
	ChainValue    float64 `json:"chainValue"`
	ChainValueZat int64   `json:"chainValueZat"`
	ValueDelta    float64 `json:"valueDelta"`
	ValueDeltaZat int64   `json:"valueDeltaZat"`
}

// ValuePool represents a value pool (transparent, sprout, sapling, orchard)
type ValuePool struct {
	ID            string  `json:"id"`
	Monitored     bool    `json:"monitored"`
	ChainValue    float64 `json:"chainValue"`
	ChainValueZat int64   `json:"chainValueZat"`
	ValueDelta    float64 `json:"valueDelta"`
	ValueDeltaZat int64   `json:"valueDeltaZat"`
}

// CommitTrees represents the commitment tree information
type CommitTrees struct {
	Sapling *TreeInfo `json:"sapling,omitempty"`
	Orchard *TreeInfo `json:"orchard,omitempty"`
}

// TreeInfo represents commitment tree subtree information
type TreeInfo struct {
	Size int64 `json:"size"`
}

// ZcashTransaction represents a complete Zcash transaction
type ZcashTransaction struct {
	// Transaction identification
	TxID       string `json:"txid"`
	AuthDigest string `json:"authdigest,omitempty"`
	Hex        string `json:"hex"`
	Size       int    `json:"size"`
	Version    int    `json:"version"`
	LockTime   uint32 `json:"locktime"`

	// Transaction flags
	Overwintered   bool   `json:"overwintered"`
	VersionGroupID string `json:"versiongroupid,omitempty"`
	ExpiryHeight   int64  `json:"expiryheight,omitempty"`

	// Inputs and outputs
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`

	// Shielded components (Sapling)
	VShieldedSpend  []ShieldedSpend  `json:"vShieldedSpend"`
	VShieldedOutput []ShieldedOutput `json:"vShieldedOutput"`
	ValueBalance    float64          `json:"valueBalance"`
	ValueBalanceZat int64            `json:"valueBalanceZat"`

	// Sprout JoinSplits
	VJoinSplit []JoinSplit `json:"vjoinsplit"`

	// Orchard actions
	Orchard *OrchardBundle `json:"orchard,omitempty"`

	// Block context
	InActiveChain bool   `json:"in_active_chain,omitempty"`
	Height        int64  `json:"height,omitempty"`
	Confirmations int64  `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	BlockHash     string `json:"blockhash,omitempty"`
	BlockTime     int64  `json:"blocktime,omitempty"`
}

// Vin represents a transaction input
type Vin struct {
	// For coinbase transactions
	Coinbase string `json:"coinbase,omitempty"`

	// For regular transactions
	TxID      string     `json:"txid,omitempty"`
	Vout      uint32     `json:"vout,omitempty"`
	ScriptSig *ScriptSig `json:"scriptSig,omitempty"`

	// Common fields
	Sequence uint32 `json:"sequence"`
}

// ScriptSig represents the signature script
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// Vout represents a transaction output
type Vout struct {
	Value        float64       `json:"value"`
	ValueZat     int64         `json:"valueZat"`
	N            uint32        `json:"n"`
	ScriptPubKey *ScriptPubKey `json:"scriptPubKey"`
}

// ScriptPubKey represents the public key script
type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

// ShieldedSpend represents a Sapling spend
type ShieldedSpend struct {
	CV           string `json:"cv"`
	Anchor       string `json:"anchor"`
	Nullifier    string `json:"nullifier"`
	Rk           string `json:"rk"`
	Proof        string `json:"proof"`
	SpendAuthSig string `json:"spendAuthSig"`
}

// ShieldedOutput represents a Sapling output
type ShieldedOutput struct {
	CV            string `json:"cv"`
	Cmu           string `json:"cmu"`
	EphemeralKey  string `json:"ephemeralKey"`
	EncCiphertext string `json:"encCiphertext"`
	OutCiphertext string `json:"outCiphertext"`
	Proof         string `json:"proof"`
}

// JoinSplit represents a Sprout JoinSplit
type JoinSplit struct {
	VPubOld       float64  `json:"vpub_old"`
	VPubNew       float64  `json:"vpub_new"`
	Anchor        string   `json:"anchor"`
	Nullifiers    []string `json:"nullifiers"`
	Commitments   []string `json:"commitments"`
	OneTimePubKey string   `json:"onetimePubKey"`
	RandomSeed    string   `json:"randomSeed"`
	Macs          []string `json:"macs"`
	Proof         string   `json:"proof"`
	Ciphertexts   []string `json:"ciphertexts"`
}

// OrchardBundle represents Orchard actions in a transaction
type OrchardBundle struct {
	Actions         []OrchardAction `json:"actions"`
	ValueBalance    float64         `json:"valueBalance"`
	ValueBalanceZat int64           `json:"valueBalanceZat"`
	Flags           string          `json:"flags,omitempty"`
	Anchor          string          `json:"anchor,omitempty"`
	Proof           string          `json:"proof,omitempty"`
	BindingSig      string          `json:"bindingSig,omitempty"`
}

// OrchardAction represents a single Orchard action
type OrchardAction struct {
	CV            string `json:"cv"`
	Nullifier     string `json:"nullifier"`
	Rk            string `json:"rk"`
	Cmx           string `json:"cmx"`
	EphemeralKey  string `json:"ephemeralKey"`
	EncCiphertext string `json:"encCiphertext"`
	OutCiphertext string `json:"outCiphertext"`
}

// IsTZETransaction checks if a transaction is a TZE transaction
// TZE transactions are identified by version 65535 (0xFFFF) and versiongroupid "ffffffff"
func (tx *ZcashTransaction) IsTZETransaction() bool {
	return tx.Version == 65535 && tx.VersionGroupID == "ffffffff"
}

// IsCoinbase checks if a transaction is a coinbase transaction
func (tx *ZcashTransaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && tx.Vin[0].Coinbase != ""
}

// HasTZEOutputs checks if any of the transaction outputs are TZE outputs
// TZE outputs are identified by scriptPubKey hex starting with "ff"
func (tx *ZcashTransaction) HasTZEOutputs() bool {
	for _, vout := range tx.Vout {
		if vout.ScriptPubKey != nil && len(vout.ScriptPubKey.Hex) >= 2 &&
			vout.ScriptPubKey.Hex[:2] == "ff" {
			return true
		}
	}
	return false
}
