package tze_graph

// TzeInput represents a TZE input in a transaction
type TzeInput struct {
	TxID     string `json:"txid" db:"txid"`
	Vin      int    `json:"vin" db:"vin"`
	Value    int64  `json:"value" db:"value"`
	PrevTxID string `json:"prev_txid" db:"prev_txid"`
	PrevVout int    `json:"prev_vout" db:"prev_vout"`
	TzeType  int32  `json:"tze_type" db:"tze_type"` // 4-byte extension_id (0=demo, 1=stark_verify)
	TzeMode  int32  `json:"tze_mode" db:"tze_mode"` // 4-byte mode (demo: 0=open, 1=close; stark_verify: 0=initialize, 1=verify)
}

// TzeOutput represents a TZE output in a transaction
type TzeOutput struct {
	TxID          string  `json:"txid" db:"txid"`
	Vout          int     `json:"vout" db:"vout"`
	Value         int64   `json:"value" db:"value"`
	SpentByTxID   *string `json:"spent_by_txid,omitempty" db:"spent_by_txid"`
	SpentByVin    *int    `json:"spent_by_vin,omitempty" db:"spent_by_vin"`
	SpentAtHeight *int64  `json:"spent_at_height,omitempty" db:"spent_at_height"`
	TzeType       int32   `json:"tze_type" db:"tze_type"`         // 4-byte extension_id (0=demo, 1=stark_verify)
	TzeMode       int32   `json:"tze_mode" db:"tze_mode"`         // 4-byte mode (demo: 0=open, 1=close; stark_verify: 0=initialize, 1=verify)
	Precondition  []byte  `json:"precondition" db:"precondition"` // TZE precondition data
}

// TzeType represents the type of TZE transaction (4-byte extension_id)
type TzeType int32

const (
	TzeTypeDemo        TzeType = 0
	TzeTypeStarkVerify TzeType = 1
)

// String returns the string representation of TzeType
func (t TzeType) String() string {
	switch t {
	case TzeTypeDemo:
		return "demo"
	case TzeTypeStarkVerify:
		return "stark_verify"
	default:
		return "unknown"
	}
}

// ParseTzeType converts a string to TzeType
func ParseTzeType(s string) (TzeType, bool) {
	switch s {
	case "demo":
		return TzeTypeDemo, true
	case "stark_verify":
		return TzeTypeStarkVerify, true
	default:
		return -1, false
	}
}

// TzeMode represents the mode of TZE operation (4-byte mode field)
// Note: The meaning of mode values depends on the TzeType
// For demo: 0=open, 1=close
// For stark_verify: 0=initialize, 1=verify
type TzeMode int32

const (
	// Demo modes
	TzeModeOpen  TzeMode = 0
	TzeModeClose TzeMode = 1

	// StarkVerify modes
	TzeModeInitialize TzeMode = 0
	TzeModeVerify     TzeMode = 1
)

// String returns the string representation of TzeMode for a given TzeType
func (m TzeMode) String(tzeType TzeType) string {
	switch tzeType {
	case TzeTypeDemo:
		switch m {
		case TzeModeOpen:
			return "open"
		case TzeModeClose:
			return "close"
		default:
			return "unknown"
		}
	case TzeTypeStarkVerify:
		switch m {
		case TzeModeInitialize:
			return "initialize"
		case TzeModeVerify:
			return "verify"
		default:
			return "unknown"
		}
	default:
		return "unknown"
	}
}

// ParseTzeMode converts a string to TzeMode for a given TzeType
func ParseTzeMode(s string, tzeType TzeType) (TzeMode, bool) {
	switch tzeType {
	case TzeTypeDemo:
		switch s {
		case "open":
			return TzeModeOpen, true
		case "close":
			return TzeModeClose, true
		default:
			return -1, false
		}
	case TzeTypeStarkVerify:
		switch s {
		case "initialize":
			return TzeModeInitialize, true
		case "verify":
			return TzeModeVerify, true
		default:
			return -1, false
		}
	default:
		return -1, false
	}
}
