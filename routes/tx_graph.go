package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tx_graph"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

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
	if depth > 10 {
		depth = 10
	}

	edges, err := tx_graph.GetTransactionGraph(txid, depth)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, edges)
}
