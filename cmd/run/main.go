package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/provider"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes"

	// Import core schemas to register their initialization functions
	_ "github.com/keep-starknet-strange/ztarknet/zindex/internal/blocks"

	// Import modules to register their schema initialization functions
	_ "github.com/keep-starknet-strange/ztarknet/zindex/internal/accounts"
	_ "github.com/keep-starknet-strange/ztarknet/zindex/internal/starks"
	_ "github.com/keep-starknet-strange/ztarknet/zindex/internal/tx_graph"
	_ "github.com/keep-starknet-strange/ztarknet/zindex/internal/tze_graph"
)

func main() {
	var (
		configPath string
		rpcURL     string
		startBlock int64
	)

	flag.StringVar(&configPath, "config", "configs/config.yaml", "Path to config file")
	flag.StringVar(&rpcURL, "rpc", "", "Zcash RPC URL (overrides config)")
	flag.Int64Var(&startBlock, "start-block", -1, "Starting block height (optional, -1 for resume)")
	flag.Parse()

	log.Println("Initializing zIndex...")

	config.InitConfig(configPath)

	if rpcURL != "" {
		log.Printf("Overriding RPC URL with: %s", rpcURL)
		config.Conf.Rpc.Url = rpcURL
	}

	log.Println("Connecting to PostgreSQL...")
	if err := postgres.InitPostgres(); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer postgres.ClosePostgres()

	log.Println("Initializing Zcash provider...")
	if err := provider.InitProvider(startBlock); err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}
	defer provider.CloseProvider()

	log.Printf("Starting API server on %s:%s...", config.Conf.Api.Host, config.Conf.Api.Port)
	go routes.StartServer(config.Conf.Api.Host, config.Conf.Api.Port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-interrupt:
			log.Println("Interrupt signal received, shutting down...")
			return
		case err := <-provider.ErrorChannel:
			log.Printf("Provider error: %v", err)
			fmt.Println("Critical error occurred, shutting down...")
			return
		}
	}
}
