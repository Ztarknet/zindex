package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/accounts"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

// GetAccount retrieves a single account by address
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

// GetAccounts retrieves all accounts with pagination
func GetAccounts(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	accountList, err := accounts.GetAccounts(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, accountList)
}

// GetAccountsByBalanceRange retrieves accounts within a specified balance range
func GetAccountsByBalanceRange(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	minBalance := int64(utils.ParseQueryParamInt(r, "min_balance", 0))
	maxBalance := int64(utils.ParseQueryParamInt(r, "max_balance", -1))

	if maxBalance < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: max_balance")
		return
	}

	if minBalance < 0 {
		minBalance = 0
	}

	if minBalance > maxBalance {
		utils.WriteErrorJson(w, http.StatusBadRequest, "min_balance must be less than or equal to max_balance")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	accountList, err := accounts.GetAccountsByBalanceRange(minBalance, maxBalance, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, accountList)
}

// GetTopAccountsByBalance retrieves accounts with the highest balances
func GetTopAccountsByBalance(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	limit, _ = utils.NormalizePagination(limit, 0)

	accountList, err := accounts.GetTopAccountsByBalance(limit)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, accountList)
}

// GetAccountTransactions retrieves all transactions for a specific account
func GetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := accounts.GetAccountTransactions(address, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetAccountTransactionsByType retrieves transactions for an account filtered by type
func GetAccountTransactionsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	txType := utils.ParseQueryParam(r, "type", "")
	if txType == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Validate transaction type
	validTypes := map[string]bool{
		string(accounts.TxTypeReceive): true,
		string(accounts.TxTypeSend):    true,
	}
	if !validTypes[txType] {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid transaction type. Must be one of: receive, send")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := accounts.GetAccountTransactionsByType(address, txType, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetAccountReceivingTransactions retrieves receiving transactions for an account
func GetAccountReceivingTransactions(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := accounts.GetAccountReceivingTransactions(address, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetAccountSendingTransactions retrieves sending transactions for an account
func GetAccountSendingTransactions(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := accounts.GetAccountSendingTransactions(address, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetAccountTransactionsByBlockRange retrieves transactions for an account within a block range
func GetAccountTransactionsByBlockRange(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	fromBlock := int64(utils.ParseQueryParamInt(r, "from_block", -1))
	if fromBlock < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: from_block")
		return
	}

	toBlock := int64(utils.ParseQueryParamInt(r, "to_block", -1))
	if toBlock < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: to_block")
		return
	}

	if fromBlock > toBlock {
		utils.WriteErrorJson(w, http.StatusBadRequest, "from_block must be less than or equal to to_block")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	txs, err := accounts.GetAccountTransactionsByBlockRange(address, fromBlock, toBlock, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetAccountTransactionCount returns the total number of transactions for an account
func GetAccountTransactionCount(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	count, err := accounts.GetAccountTransactionCount(address)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, map[string]int64{"count": count})
}

// GetAccountTransaction retrieves a specific transaction for an account
func GetAccountTransaction(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	address := utils.ParseQueryParam(r, "address", "")
	if address == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: address")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	tx, err := accounts.GetAccountTransaction(address, txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if tx == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Account transaction not found")
		return
	}

	utils.WriteDataJson(w, tx)
}

// GetTransactionAccounts retrieves all accounts associated with a transaction
func GetTransactionAccounts(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	txs, err := accounts.GetTransactionAccounts(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, txs)
}

// GetRecentActiveAccounts retrieves accounts with recent transaction activity
func GetRecentActiveAccounts(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("ACCOUNTS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "Accounts module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	limit, _ = utils.NormalizePagination(limit, 0)

	accountList, err := accounts.GetRecentActiveAccounts(limit)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, accountList)
}
