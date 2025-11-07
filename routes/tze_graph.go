package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tze_graph"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

func GetTZETransaction(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	tx, err := tze_graph.GetTZETransaction(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if tx == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE transaction not found")
		return
	}

	utils.WriteDataJson(w, tx)
}

func GetTZETransactionsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeType := utils.ParseQueryParam(r, "type", "")
	if tzeType == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	txs, err := tze_graph.GetTZETransactionsByType(tzeType, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

func GetTZEWitnesses(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	witnesses, err := tze_graph.GetTZEWitnesses(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, witnesses)
}
