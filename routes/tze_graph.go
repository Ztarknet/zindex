package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/tze_graph"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

// ============================================================================
// TZE INPUT ROUTES
// ============================================================================

// GetTzeInputs retrieves all inputs for a transaction
func GetTzeInputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	inputs, err := tze_graph.GetTzeInputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, inputs)
}

// GetTzeInput retrieves a specific input by txid and vin
func GetTzeInput(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
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

	input, err := tze_graph.GetTzeInput(txid, vin)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if input == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE input not found")
		return
	}

	utils.WriteDataJson(w, input)
}

// GetTzeInputsByType retrieves all inputs of a specific TZE type with pagination
func GetTzeInputsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, inputs)
}

// GetTzeInputsByMode retrieves all inputs of a specific TZE mode with pagination
// Note: mode values have different meanings depending on type context
func GetTzeInputsByMode(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	modeInt := int16(utils.ParseQueryParamInt(r, "mode", -1))
	if modeInt < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: mode (must be 0 or 1)")
		return
	}

	// Validate mode value (0 or 1 are valid)
	if modeInt > 1 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid mode value. Must be 0 or 1")
		return
	}

	tzeMode := tze_graph.TzeMode(modeInt)

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, inputs)
}

// GetTzeInputsByTypeAndMode retrieves all inputs matching both type and mode with pagination
func GetTzeInputsByTypeAndMode(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	modeStr := utils.ParseQueryParam(r, "mode", "")
	if modeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: mode")
		return
	}

	// Parse and validate TZE mode based on type
	tzeMode, ok := tze_graph.ParseTzeMode(modeStr, tzeType)
	if !ok {
		var validModes string
		if tzeType == tze_graph.TzeTypeDemo {
			validModes = "open, close"
		} else {
			validModes = "initialize, verify"
		}
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE mode for type "+tzeTypeStr+". Must be one of: "+validModes)
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, inputs)
}

// GetTzeInputsByPrevOutput retrieves all inputs spending a specific previous output
func GetTzeInputsByPrevOutput(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	prevTxid := utils.ParseQueryParam(r, "prev_txid", "")
	if prevTxid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: prev_txid")
		return
	}

	prevVout := utils.ParseQueryParamInt(r, "prev_vout", -1)
	if prevVout < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: prev_vout")
		return
	}

	inputs, err := tze_graph.GetTzeInputsByPrevOutput(prevTxid, prevVout)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, inputs)
}

// ============================================================================
// TZE OUTPUT ROUTES
// ============================================================================

// GetTzeOutputs retrieves all outputs for a transaction
func GetTzeOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	outputs, err := tze_graph.GetTzeOutputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, outputs)
}

// GetTzeOutput retrieves a specific output by txid and vout
func GetTzeOutput(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
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

	output, err := tze_graph.GetTzeOutput(txid, vout)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if output == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE output not found")
		return
	}

	utils.WriteDataJson(w, output)
}

// GetUnspentTzeOutputs retrieves all unspent outputs for a transaction
func GetUnspentTzeOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	outputs, err := tze_graph.GetUnspentTzeOutputs(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, outputs)
}

// GetAllUnspentTzeOutputs retrieves all unspent TZE outputs with pagination
func GetAllUnspentTzeOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetTzeOutputsByType retrieves all outputs of a specific TZE type with pagination
func GetTzeOutputsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetTzeOutputsByMode retrieves all outputs of a specific TZE mode with pagination
// Note: mode values have different meanings depending on type context
func GetTzeOutputsByMode(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	modeInt := int16(utils.ParseQueryParamInt(r, "mode", -1))
	if modeInt < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: mode (must be 0 or 1)")
		return
	}

	// Validate mode value (0 or 1 are valid)
	if modeInt > 1 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid mode value. Must be 0 or 1")
		return
	}

	tzeMode := tze_graph.TzeMode(modeInt)

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetTzeOutputsByTypeAndMode retrieves all outputs matching both type and mode with pagination
func GetTzeOutputsByTypeAndMode(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	modeStr := utils.ParseQueryParam(r, "mode", "")
	if modeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: mode")
		return
	}

	// Parse and validate TZE mode based on type
	tzeMode, ok := tze_graph.ParseTzeMode(modeStr, tzeType)
	if !ok {
		var validModes string
		if tzeType == tze_graph.TzeTypeDemo {
			validModes = "open, close"
		} else {
			validModes = "initialize, verify"
		}
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE mode for type "+tzeTypeStr+". Must be one of: "+validModes)
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetUnspentTzeOutputsByType retrieves all unspent outputs of a specific type with pagination
func GetUnspentTzeOutputsByType(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetUnspentTzeOutputsByTypeAndMode retrieves all unspent outputs matching type and mode
func GetUnspentTzeOutputsByTypeAndMode(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	tzeTypeStr := utils.ParseQueryParam(r, "type", "")
	if tzeTypeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: type")
		return
	}

	// Parse and validate TZE type
	tzeType, ok := tze_graph.ParseTzeType(tzeTypeStr)
	if !ok {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE type. Must be one of: demo, stark_verify")
		return
	}

	modeStr := utils.ParseQueryParam(r, "mode", "")
	if modeStr == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: mode")
		return
	}

	// Parse and validate TZE mode based on type
	tzeMode, ok := tze_graph.ParseTzeMode(modeStr, tzeType)
	if !ok {
		var validModes string
		if tzeType == tze_graph.TzeTypeDemo {
			validModes = "open, close"
		} else {
			validModes = "initialize, verify"
		}
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid TZE mode for type "+tzeTypeStr+". Must be one of: "+validModes)
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetSpentTzeOutputs retrieves all spent outputs with pagination
func GetSpentTzeOutputs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}

// GetTzeOutputsByValue retrieves outputs with value greater than or equal to minimum value
func GetTzeOutputsByValue(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("TZE_GRAPH") {
		utils.WriteErrorJson(w, http.StatusNotFound, "TZE graph module is disabled")
		return
	}

	minValue := int64(utils.ParseQueryParamInt(r, "min_value", 0))
	if minValue < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Invalid parameter: min_value must be non-negative")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", 50)
	offset := utils.ParseQueryParamInt(r, "offset", 0)


	utils.WriteDataJson(w, outputs)
}
