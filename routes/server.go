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

	mux.HandleFunc("/health", HealthCheck)
	mux.HandleFunc("/api/v1/account", GetAccount)
	mux.HandleFunc("/api/v1/accounts", GetAccounts)
	mux.HandleFunc("/api/v1/transaction", GetTransaction)
	mux.HandleFunc("/api/v1/transaction/graph", GetTransactionGraph)
	mux.HandleFunc("/api/v1/tze/transaction", GetTZETransaction)
	mux.HandleFunc("/api/v1/tze/transactions", GetTZETransactionsByType)
	mux.HandleFunc("/api/v1/tze/witnesses", GetTZEWitnesses)
	mux.HandleFunc("/api/v1/proof", GetProof)
	mux.HandleFunc("/api/v1/proof/transaction", GetProofsByTransaction)
	mux.HandleFunc("/api/v1/proof/stats", GetProofStats)
	mux.HandleFunc("/api/v1/proof/unverified", GetUnverifiedProofs)

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
