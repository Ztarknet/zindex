package routes

import (
	"net/http"
	"strconv"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/starks"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

func GetProof(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	idStr := utils.ParseQueryParam(r, "id", "")
	if idStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid proof id")
		return
	}

	proof, err := starks.GetProof(id)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if proof == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Proof not found")
		return
	}

	utils.WriteDataJson(w, proof)
}

func GetProofsByTransaction(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	proofs, err := starks.GetProofsByTransaction(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

func GetProofStats(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	stats, err := starks.GetProofStats()
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, stats)
}

func GetUnverifiedProofs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	if limit > 100 {
		limit = 100
	}

	proofs, err := starks.GetUnverifiedProofs(limit)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}
