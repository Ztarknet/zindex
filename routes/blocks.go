package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/blocks"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

// GetBlock retrieves a single block by height
func GetBlock(w http.ResponseWriter, r *http.Request) {
	height := int64(utils.ParseQueryParamInt(r, "height", -1))
	if height < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: height")
		return
	}

	block, err := blocks.GetBlock(height)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if block == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Block not found")
		return
	}

	utils.WriteDataJson(w, block)
}

// GetBlockByHash retrieves a single block by hash
func GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	hash := utils.ParseQueryParam(r, "hash", "")
	if hash == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: hash")
		return
	}

	block, err := blocks.GetBlockByHash(hash)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if block == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Block not found")
		return
	}

	utils.WriteDataJson(w, block)
}

// GetBlocks retrieves blocks with pagination
func GetBlocks(w http.ResponseWriter, r *http.Request) {
	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	blockList, err := blocks.GetBlocks(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, blockList)
}

// GetBlocksByRange retrieves blocks within a height range
func GetBlocksByRange(w http.ResponseWriter, r *http.Request) {
	fromHeight := int64(utils.ParseQueryParamInt(r, "from_height", -1))
	if fromHeight < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: from_height")
		return
	}

	toHeight := int64(utils.ParseQueryParamInt(r, "to_height", -1))
	if toHeight < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: to_height")
		return
	}

	if fromHeight > toHeight {
		utils.WriteErrorJson(w, http.StatusBadRequest, "from_height must be less than or equal to to_height")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	blockList, err := blocks.GetBlocksByRange(fromHeight, toHeight, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, blockList)
}

// GetBlocksByTimestampRange retrieves blocks within a timestamp range
func GetBlocksByTimestampRange(w http.ResponseWriter, r *http.Request) {
	fromTimestamp := int64(utils.ParseQueryParamInt(r, "from_timestamp", -1))
	if fromTimestamp < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: from_timestamp")
		return
	}

	toTimestamp := int64(utils.ParseQueryParamInt(r, "to_timestamp", -1))
	if toTimestamp < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: to_timestamp")
		return
	}

	if fromTimestamp > toTimestamp {
		utils.WriteErrorJson(w, http.StatusBadRequest, "from_timestamp must be less than or equal to to_timestamp")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	blockList, err := blocks.GetBlocksByTimestampRange(fromTimestamp, toTimestamp, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, blockList)
}

// GetRecentBlocks retrieves the most recent blocks
func GetRecentBlocks(w http.ResponseWriter, r *http.Request) {
	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	limit, _ = utils.NormalizePagination(limit, 0)

	blockList, err := blocks.GetRecentBlocks(limit)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, blockList)
}

// GetBlockCount returns the total number of blocks
func GetBlockCount(w http.ResponseWriter, r *http.Request) {
	count, err := blocks.GetBlockCount()
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, map[string]int64{"count": count})
}

// GetLatestBlock retrieves the most recent block
func GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	block, err := blocks.GetLatestBlock()
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if block == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "No blocks found")
		return
	}

	utils.WriteDataJson(w, block)
}
