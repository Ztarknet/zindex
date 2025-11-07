package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/accounts"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

func GetAccount(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	account, err := accounts.GetAccount(address)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if account == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Account not found")
		return
	}

	utils.WriteDataJson(w, account)
}

func GetAccounts(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	accountList, err := accounts.GetAccounts(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, accountList)
}
