package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

func StartServer(host, port string) {
	mux := http.NewServeMux()

	// Enable base routes (always enabled)
	EnableBaseRoutes(mux)

	// Enable block routes (always enabled)
	EnableBlockRoutes(mux)

	// Enable module-specific routes based on configuration
	EnableAccountsRoutes(mux)
	EnableTxGraphRoutes(mux)
	EnableTzeGraphRoutes(mux)
	EnableStarksRoutes(mux)

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("API server listening on %s", addr)

	// Configure server with timeouts and limits from config
	server := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.Conf.Api.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(config.Conf.Api.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(config.Conf.Api.IdleTimeout) * time.Second,
		MaxHeaderBytes: config.Conf.Api.MaxHeaderBytes,
	}

	log.Printf("Server configured with ReadTimeout: %ds, WriteTimeout: %ds, IdleTimeout: %ds, MaxHeaderBytes: %d",
		config.Conf.Api.ReadTimeout,
		config.Conf.Api.WriteTimeout,
		config.Conf.Api.IdleTimeout,
		config.Conf.Api.MaxHeaderBytes)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteResultJson(w, "healthy")
}

// EnableBaseRoutes registers base routes that are always available
func EnableBaseRoutes(mux *http.ServeMux) {
	log.Println("Registering base routes")

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.SetCorsHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	// Health check endpoint
	mux.HandleFunc("/health", HealthCheck)
}

// EnableAccountsRoutes registers all accounts module routes if the module is enabled
func EnableAccountsRoutes(mux *http.ServeMux) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		log.Println("Accounts module is disabled, skipping route registration")
		return
	}

	log.Println("Registering Accounts module routes")

	// Account routes
	mux.HandleFunc("/api/v1/accounts", GetAccounts)
	mux.HandleFunc("/api/v1/accounts/account", GetAccount)
	mux.HandleFunc("/api/v1/accounts/balance-range", GetAccountsByBalanceRange)
	mux.HandleFunc("/api/v1/accounts/top-balances", GetTopAccountsByBalance)
	mux.HandleFunc("/api/v1/accounts/recent-active", GetRecentActiveAccounts)

	// Account transaction routes
	mux.HandleFunc("/api/v1/accounts/transactions", GetAccountTransactions)
	mux.HandleFunc("/api/v1/accounts/transactions/type", GetAccountTransactionsByType)
	mux.HandleFunc("/api/v1/accounts/transactions/incoming", GetAccountIncomingTransactions)
	mux.HandleFunc("/api/v1/accounts/transactions/outgoing", GetAccountOutgoingTransactions)
	mux.HandleFunc("/api/v1/accounts/transactions/block-range", GetAccountTransactionsByBlockRange)
	mux.HandleFunc("/api/v1/accounts/transactions/count", GetAccountTransactionCount)
	mux.HandleFunc("/api/v1/accounts/transactions/transaction", GetAccountTransaction)
	mux.HandleFunc("/api/v1/accounts/transactions/by-txid", GetTransactionAccounts)
}

// EnableTxGraphRoutes registers all transaction graph module routes if the module is enabled
func EnableTxGraphRoutes(mux *http.ServeMux) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		log.Println("Transaction graph module is disabled, skipping route registration")
		return
	}

	log.Println("Registering Transaction Graph module routes")

	// Transaction routes
	mux.HandleFunc("/api/v1/tx-graph/transaction", GetTransaction)
	mux.HandleFunc("/api/v1/tx-graph/transactions/by-block", GetTransactionsByBlock)
	mux.HandleFunc("/api/v1/tx-graph/transactions/by-type", GetTransactionsByType)
	mux.HandleFunc("/api/v1/tx-graph/transactions/recent", GetRecentTransactions)

	// Transaction output routes
	mux.HandleFunc("/api/v1/tx-graph/outputs", GetTransactionOutputs)
	mux.HandleFunc("/api/v1/tx-graph/outputs/output", GetTransactionOutput)
	mux.HandleFunc("/api/v1/tx-graph/outputs/unspent", GetUnspentOutputs)
	mux.HandleFunc("/api/v1/tx-graph/outputs/spenders", GetOutputSpenders)

	// Transaction input routes
	mux.HandleFunc("/api/v1/tx-graph/inputs", GetTransactionInputs)
	mux.HandleFunc("/api/v1/tx-graph/inputs/input", GetTransactionInput)
	mux.HandleFunc("/api/v1/tx-graph/inputs/sources", GetInputSources)

	// Transaction graph routes
	mux.HandleFunc("/api/v1/tx-graph/graph", GetTransactionGraph)
}

// EnableTzeGraphRoutes registers all TZE graph module routes if the module is enabled
func EnableTzeGraphRoutes(mux *http.ServeMux) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		log.Println("TZE graph module is disabled, skipping route registration")
		return
	}

	log.Println("Registering TZE Graph module routes")

	// TZE input routes
	mux.HandleFunc("/api/v1/tze-graph/inputs", GetTzeInputs)
	mux.HandleFunc("/api/v1/tze-graph/inputs/input", GetTzeInput)
	mux.HandleFunc("/api/v1/tze-graph/inputs/by-type", GetTzeInputsByType)
	mux.HandleFunc("/api/v1/tze-graph/inputs/by-mode", GetTzeInputsByMode)
	mux.HandleFunc("/api/v1/tze-graph/inputs/by-type-mode", GetTzeInputsByTypeAndMode)
	mux.HandleFunc("/api/v1/tze-graph/inputs/by-prev-output", GetTzeInputsByPrevOutput)

	// TZE output routes
	mux.HandleFunc("/api/v1/tze-graph/outputs", GetTzeOutputs)
	mux.HandleFunc("/api/v1/tze-graph/outputs/output", GetTzeOutput)
	mux.HandleFunc("/api/v1/tze-graph/outputs/unspent", GetUnspentTzeOutputs)
	mux.HandleFunc("/api/v1/tze-graph/outputs/all-unspent", GetAllUnspentTzeOutputs)
	mux.HandleFunc("/api/v1/tze-graph/outputs/by-type", GetTzeOutputsByType)
	mux.HandleFunc("/api/v1/tze-graph/outputs/by-mode", GetTzeOutputsByMode)
	mux.HandleFunc("/api/v1/tze-graph/outputs/by-type-mode", GetTzeOutputsByTypeAndMode)
	mux.HandleFunc("/api/v1/tze-graph/outputs/unspent-by-type", GetUnspentTzeOutputsByType)
	mux.HandleFunc("/api/v1/tze-graph/outputs/unspent-by-type-mode", GetUnspentTzeOutputsByTypeAndMode)
	mux.HandleFunc("/api/v1/tze-graph/outputs/spent", GetSpentTzeOutputs)
	mux.HandleFunc("/api/v1/tze-graph/outputs/by-value", GetTzeOutputsByValue)
}

// EnableStarksRoutes registers all STARK module routes if the module is enabled
func EnableStarksRoutes(mux *http.ServeMux) {
	if !config.IsModuleEnabled("STARKS") {
		log.Println("STARKS module is disabled, skipping route registration")
		return
	}

	log.Println("Registering STARKS module routes")

	// Verifier routes
	mux.HandleFunc("/api/v1/starks/verifiers/verifier", GetVerifier)
	mux.HandleFunc("/api/v1/starks/verifiers/by-name", GetVerifierByName)
	mux.HandleFunc("/api/v1/starks/verifiers", GetAllVerifiers)
	mux.HandleFunc("/api/v1/starks/verifiers/by-balance", GetVerifiersByBalance)

	// STARK proof routes
	mux.HandleFunc("/api/v1/starks/proofs/proof", GetStarkProof)
	mux.HandleFunc("/api/v1/starks/proofs/by-verifier", GetStarkProofsByVerifier)
	mux.HandleFunc("/api/v1/starks/proofs/by-transaction", GetStarkProofsByTransaction)
	mux.HandleFunc("/api/v1/starks/proofs/by-block", GetStarkProofsByBlock)
	mux.HandleFunc("/api/v1/starks/proofs/recent", GetRecentStarkProofs)
	mux.HandleFunc("/api/v1/starks/proofs/by-size", GetStarkProofsBySize)

	// Ztarknet facts routes
	mux.HandleFunc("/api/v1/starks/facts/facts", GetZtarknetFacts)
	mux.HandleFunc("/api/v1/starks/facts/by-verifier", GetZtarknetFactsByVerifier)
	mux.HandleFunc("/api/v1/starks/facts/by-transaction", GetZtarknetFactsByTransaction)
	mux.HandleFunc("/api/v1/starks/facts/by-block", GetZtarknetFactsByBlock)
	mux.HandleFunc("/api/v1/starks/facts/by-state", GetZtarknetFactsByState)
	mux.HandleFunc("/api/v1/starks/facts/by-program-hash", GetZtarknetFactsByProgramHash)
	mux.HandleFunc("/api/v1/starks/facts/by-inner-program-hash", GetZtarknetFactsByInnerProgramHash)
	mux.HandleFunc("/api/v1/starks/facts/recent", GetRecentZtarknetFacts)
	mux.HandleFunc("/api/v1/starks/facts/state-transition", GetStateTransition)
}

// EnableBlockRoutes registers all block routes (always enabled)
func EnableBlockRoutes(mux *http.ServeMux) {
	log.Println("Registering Block routes")

	// Block routes
	mux.HandleFunc("/api/v1/blocks", GetBlocks)
	mux.HandleFunc("/api/v1/blocks/block", GetBlock)
	mux.HandleFunc("/api/v1/blocks/by-hash", GetBlockByHash)
	mux.HandleFunc("/api/v1/blocks/range", GetBlocksByRange)
	mux.HandleFunc("/api/v1/blocks/timestamp-range", GetBlocksByTimestampRange)
	mux.HandleFunc("/api/v1/blocks/recent", GetRecentBlocks)
	mux.HandleFunc("/api/v1/blocks/count", GetBlockCount)
	mux.HandleFunc("/api/v1/blocks/latest", GetLatestBlock)
}
