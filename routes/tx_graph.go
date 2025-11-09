package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tx_graph"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

// GetTransaction retrieves a single transaction by txid
func GetTransaction(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	tx, err := tx_graph.GetTransaction(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if tx == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction not found")
		return
	}

	utils.WriteDataJson(w, tx)
}

// GetTransactionsByBlock retrieves all transactions in a specific block
func GetTransactionsByBlock(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	blockHeight := int64(utils.ParseQueryParamInt(r, "block_height", -1))
	if blockHeight < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: block_height")
		return
	}

	txs, err := tx_graph.GetTransactionsByBlock(blockHeight)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetTransactionsByType retrieves transactions filtered by type with pagination
func GetTransactionsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txType := utils.ParseQueryParam(r, "type", "")
	if txType == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Validate transaction type
	validTypes := map[string]bool{
		string(tx_graph.TxTypeCoinbase): true,
		string(tx_graph.TxTypeTZE):      true,
		string(tx_graph.TxTypeT2T):      true,
		string(tx_graph.TxTypeT2Z):      true,
		string(tx_graph.TxTypeZ2T):      true,
		string(tx_graph.TxTypeZ2Z):      true,
	}
	if !validTypes[txType] {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid transaction type. Must be one of: coinbase, tze, t2t, t2z, z2t, z2z")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := tx_graph.GetTransactionsByType(txType, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetRecentTransactions retrieves the most recent transactions with pagination
func GetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := tx_graph.GetRecentTransactions(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetTransactionOutputs retrieves all outputs for a transaction
func GetTransactionOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	outputs, err := tx_graph.GetTransactionOutputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, outputs)
}

// GetTransactionOutput retrieves a specific output by txid and vout
func GetTransactionOutput(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	vout := utils.ParseQueryParamInt(r, "vout", -1)
	if vout < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: vout")
		return
	}

	output, err := tx_graph.GetTransactionOutput(txid, vout)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if output == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction output not found")
		return
	}

	utils.WriteDataJson(w, output)
}

// GetUnspentOutputs retrieves all unspent outputs for a transaction
func GetUnspentOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	outputs, err := tx_graph.GetUnspentOutputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, outputs)
}

// GetTransactionInputs retrieves all inputs for a transaction
func GetTransactionInputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	inputs, err := tx_graph.GetTransactionInputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, inputs)
}

// GetTransactionInput retrieves a specific input by txid and vin
func GetTransactionInput(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	vin := utils.ParseQueryParamInt(r, "vin", -1)
	if vin < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: vin")
		return
	}

	input, err := tx_graph.GetTransactionInput(txid, vin)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if input == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction input not found")
		return
	}

	utils.WriteDataJson(w, input)
}

// GetOutputSpenders retrieves all transactions that spent outputs from a given transaction
func GetOutputSpenders(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	outputs, err := tx_graph.GetOutputSpenders(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, outputs)
}

// GetInputSources retrieves all transactions that provided inputs to a given transaction
func GetInputSources(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	inputs, err := tx_graph.GetInputSources(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, inputs)
}

// GetTransactionGraph builds a graph of connected transactions up to a specified depth
func GetTransactionGraph(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TX_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Transaction graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	depth := utils.ParseQueryParamInt(r, "depth", 3)

	// Validate depth range
	if depth < 1 {
		depth = 1
	}
	// Cap depth at configured max_graph_depth to prevent excessive recursion
	maxDepth := config.Conf.Modules.TxGraph.MaxGraphDepth
	if depth > maxDepth {
		depth = maxDepth
	}

	txids, err := tx_graph.GetTransactionGraph(txid, depth)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txids)
}
