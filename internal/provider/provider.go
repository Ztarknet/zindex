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
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/indexer"
)

var (
	client       *http.Client
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

	// Create RPC client wrapper for the indexer
	rpcClient := &rpcClientWrapper{}

	// Start the indexer
	_, ErrorChannel = indexer.Start(startBlock, rpcClient)

	return nil
}

func CloseProvider() {
	log.Println("Stopping provider...")
	indexer.Stop()
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
	// Use verbosity 2 to get full transaction details
	result, err := makeRPCCall("getblock", []interface{}{hash, 2})
	if err != nil {
		return nil, err
	}

	var block map[string]interface{}
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return block, nil
}

// rpcClientWrapper implements the indexer.RpcClient interface
// It wraps the provider's RPC functions for use by the indexer
type rpcClientWrapper struct{}

func (w *rpcClientWrapper) GetBlockHash(height int64) (string, error) {
	return GetBlockHash(height)
}

func (w *rpcClientWrapper) GetBlock(hash string) (map[string]interface{}, error) {
	return GetBlock(hash)
}

func (w *rpcClientWrapper) GetBlockCount() (int64, error) {
	return GetBlockCount()
}
