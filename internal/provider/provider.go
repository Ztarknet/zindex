package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/db/postgres"
)

var (
	client       *http.Client
	stopChan     chan struct{}
	ErrorChannel chan error
)

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func InitProvider(startBlock int64) error {
	log.Println("Initializing Zcash provider...")

	client = &http.Client{
		Timeout: time.Duration(config.Conf.Rpc.Timeout) * time.Second,
	}

	stopChan = make(chan struct{})
	ErrorChannel = make(chan error, 1)

	var indexStartBlock int64
	if startBlock >= 0 {
		indexStartBlock = startBlock
		log.Printf("Starting indexer from specified block: %d", startBlock)
	} else {
		lastBlock, err := postgres.GetLastIndexedBlock()
		if err != nil {
			log.Printf("Failed to get last indexed block, starting from config: %v", err)
			indexStartBlock = config.Conf.Indexer.StartBlock
		} else {
			indexStartBlock = lastBlock + 1
			log.Printf("Resuming indexer from block: %d", indexStartBlock)
		}
	}

	go startIndexing(indexStartBlock)

	return nil
}

func CloseProvider() {
	log.Println("Stopping provider...")
	close(stopChan)
}

func startIndexing(startBlock int64) {
	currentBlock := startBlock
	pollInterval := time.Duration(config.Conf.Indexer.PollInterval) * time.Second

	log.Printf("Starting indexing loop from block %d", currentBlock)

	for {
		select {
		case <-stopChan:
			log.Println("Indexing stopped")
			return
		default:
			blockCount, err := GetBlockCount()
			if err != nil {
				log.Printf("Failed to get block count: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			if currentBlock > blockCount {
				time.Sleep(pollInterval)
				continue
			}

			batchEnd := currentBlock + int64(config.Conf.Indexer.BatchSize)
			if batchEnd > blockCount {
				batchEnd = blockCount
			}

			log.Printf("Indexing blocks %d to %d (chain height: %d)", currentBlock, batchEnd, blockCount)

			for height := currentBlock; height <= batchEnd; height++ {
				select {
				case <-stopChan:
					return
				default:
					if err := indexBlock(height); err != nil {
						log.Printf("Error indexing block %d: %v", height, err)
						ErrorChannel <- err
						return
					}
				}
			}

			currentBlock = batchEnd + 1

			if currentBlock > blockCount {
				time.Sleep(pollInterval)
			}
		}
	}
}

func indexBlock(height int64) error {
	blockHash, err := GetBlockHash(height)
	if err != nil {
		return fmt.Errorf("failed to get block hash: %w", err)
	}

	block, err := GetBlock(blockHash)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	// TODO: Implement reorg detection and handling
	// Config options available: config.Conf.Indexer.EnableReorgHandling, config.Conf.Indexer.MaxReorgDepth
	// When implemented, should:
	//   1. Compare block's previousblockhash with stored hash at height-1
	//   2. If mismatch detected, search backward up to max_reorg_depth to find common ancestor
	//   3. Rollback database to common ancestor
	//   4. Continue indexing from rollback point

	// Extract previous block hash from block data
	var prevHash string
	if val, ok := block["previousblockhash"].(string); ok {
		prevHash = val
	}

	// Store block data
	if err := postgres.StoreBlock(height, blockHash, prevHash, block); err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}

	// Update indexer state
	if err := postgres.UpdateLastIndexedBlock(height, blockHash); err != nil {
		return fmt.Errorf("failed to update last indexed block: %w", err)
	}

	log.Printf("Indexed block %d: %s", height, blockHash)

	return nil
}

func makeRPCCall(method string, params []interface{}) (json.RawMessage, error) {
	request := RPCRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	maxAttempts := config.Conf.Rpc.RetryAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			retryDelay := time.Duration(config.Conf.Rpc.RetryDelay) * time.Second
			log.Printf("Retrying RPC call to %s (attempt %d/%d) after %v", method, attempt+1, maxAttempts, retryDelay)
			time.Sleep(retryDelay)
		}

		req, err := http.NewRequest("POST", config.Conf.Rpc.Url, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		var rpcResp RPCResponse
		if err := json.Unmarshal(body, &rpcResp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		if rpcResp.Error != nil {
			lastErr = fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
			continue
		}

		return rpcResp.Result, nil
	}

	return nil, fmt.Errorf("RPC call failed after %d attempts: %w", maxAttempts, lastErr)
}

func GetBlockCount() (int64, error) {
	result, err := makeRPCCall("getblockcount", []interface{}{})
	if err != nil {
		return 0, err
	}

	var count int64
	if err := json.Unmarshal(result, &count); err != nil {
		return 0, fmt.Errorf("failed to unmarshal block count: %w", err)
	}

	return count, nil
}

func GetBlockHash(height int64) (string, error) {
	result, err := makeRPCCall("getblockhash", []interface{}{height})
	if err != nil {
		return "", err
	}

	var hash string
	if err := json.Unmarshal(result, &hash); err != nil {
		return "", fmt.Errorf("failed to unmarshal block hash: %w", err)
	}

	return hash, nil
}

func GetBlock(hash string) (map[string]interface{}, error) {
	result, err := makeRPCCall("getblock", []interface{}{hash, 1})
	if err != nil {
		return nil, err
	}

	var block map[string]interface{}
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return block, nil
}
