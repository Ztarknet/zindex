package routes

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
	"github.com/keep-starknet-strange/ztarknet/zindex/internal/starks"
	"github.com/keep-starknet-strange/ztarknet/zindex/routes/utils"
)

// ============================================================================
// Verifier Routes
// ============================================================================

// GetVerifier retrieves a single verifier by its ID
func GetVerifier(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	verifierID := utils.ParseQueryParam(r, "verifier_id", "")
	if verifierID == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_id")
		return
	}

	verifier, err := starks.GetVerifier(verifierID)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if verifier == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Verifier not found")
		return
	}

	utils.WriteDataJson(w, verifier)
}

// GetVerifierByName retrieves a verifier by its name
func GetVerifierByName(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	verifierName := utils.ParseQueryParam(r, "verifier_name", "")
	if verifierName == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_name")
		return
	}

	verifier, err := starks.GetVerifierByName(verifierName)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if verifier == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Verifier not found")
		return
	}

	utils.WriteDataJson(w, verifier)
}

// GetAllVerifiers retrieves all verifiers with pagination
func GetAllVerifiers(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	verifiers, err := starks.GetAllVerifiers(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, verifiers)
}

// GetVerifiersByBalance retrieves verifiers sorted by balance with pagination
func GetVerifiersByBalance(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	verifiers, err := starks.GetVerifiersByBalance(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, verifiers)
}

// ============================================================================
// StarkProof Routes
// ============================================================================

// GetStarkProof retrieves a STARK proof by verifier ID and transaction ID
func GetStarkProof(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	verifierID := utils.ParseQueryParam(r, "verifier_id", "")
	if verifierID == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_id")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	proof, err := starks.GetStarkProof(verifierID, txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if proof == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARK proof not found")
		return
	}

	utils.WriteDataJson(w, proof)
}

// GetStarkProofsByVerifier retrieves all STARK proofs for a verifier with pagination
func GetStarkProofsByVerifier(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	verifierID := utils.ParseQueryParam(r, "verifier_id", "")
	if verifierID == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_id")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	proofs, err := starks.GetStarkProofsByVerifier(verifierID, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

// GetStarkProofsByTransaction retrieves all STARK proofs for a transaction
func GetStarkProofsByTransaction(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	proofs, err := starks.GetStarkProofsByTransaction(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

// GetStarkProofsByBlock retrieves all STARK proofs for a specific block
func GetStarkProofsByBlock(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	blockHeight := int64(utils.ParseQueryParamInt(r, "block_height", -1))
	if blockHeight < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: block_height")
		return
	}

	proofs, err := starks.GetStarkProofsByBlock(blockHeight)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

// GetRecentStarkProofs retrieves the most recent STARK proofs with pagination
func GetRecentStarkProofs(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	proofs, err := starks.GetRecentStarkProofs(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

// GetStarkProofsBySize retrieves STARK proofs filtered by size range with pagination
func GetStarkProofsBySize(w http.ResponseWriter, r *http.Request) {
	if !config.IsModuleEnabled("STARKS") {
		utils.WriteErrorJson(w, http.StatusNotFound, "STARKS module is disabled")
		return
	}

	minSize := int64(utils.ParseQueryParamInt(r, "min_size", 0))
	maxSize := int64(utils.ParseQueryParamInt(r, "max_size", -1))

	if maxSize < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: max_size")
		return
	}

	if minSize < 0 {
		minSize = 0
	}

	if minSize > maxSize {
		utils.WriteErrorJson(w, http.StatusBadRequest, "min_size must be less than or equal to max_size")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	proofs, err := starks.GetStarkProofsBySize(minSize, maxSize, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, proofs)
}

// ============================================================================
// ZtarknetFacts Routes
// ============================================================================

// GetZtarknetFacts retrieves Ztarknet facts by verifier ID and transaction ID
func GetZtarknetFacts(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	verifierID := utils.ParseQueryParam(r, "verifier_id", "")
	if verifierID == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_id")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	facts, err := starks.GetZtarknetFacts(verifierID, txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if facts == nil {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet facts not found")
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByVerifier retrieves all Ztarknet facts for a verifier with pagination
func GetZtarknetFactsByVerifier(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	verifierID := utils.ParseQueryParam(r, "verifier_id", "")
	if verifierID == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: verifier_id")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	facts, err := starks.GetZtarknetFactsByVerifier(verifierID, limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByTransaction retrieves all Ztarknet facts for a transaction
func GetZtarknetFactsByTransaction(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	txid := utils.ParseQueryParam(r, "txid", "")
	if txid == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: txid")
		return
	}

	facts, err := starks.GetZtarknetFactsByTransaction(txid)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByBlock retrieves all Ztarknet facts for a specific block
func GetZtarknetFactsByBlock(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	blockHeight := int64(utils.ParseQueryParamInt(r, "block_height", -1))
	if blockHeight < 0 {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid required parameter: block_height")
		return
	}

	facts, err := starks.GetZtarknetFactsByBlock(blockHeight)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByState retrieves Ztarknet facts by state hash
func GetZtarknetFactsByState(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	stateHash := utils.ParseQueryParam(r, "state_hash", "")
	if stateHash == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: state_hash")
		return
	}

	facts, err := starks.GetZtarknetFactsByState(stateHash)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByProgramHash retrieves Ztarknet facts by program hash
func GetZtarknetFactsByProgramHash(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	programHash := utils.ParseQueryParam(r, "program_hash", "")
	if programHash == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: program_hash")
		return
	}

	facts, err := starks.GetZtarknetFactsByProgramHash(programHash)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetZtarknetFactsByInnerProgramHash retrieves Ztarknet facts by inner program hash
func GetZtarknetFactsByInnerProgramHash(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	innerProgramHash := utils.ParseQueryParam(r, "inner_program_hash", "")
	if innerProgramHash == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: inner_program_hash")
		return
	}

	facts, err := starks.GetZtarknetFactsByInnerProgramHash(innerProgramHash)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetRecentZtarknetFacts retrieves the most recent Ztarknet facts with pagination
func GetRecentZtarknetFacts(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	limit := utils.ParseQueryParamInt(r, "limit", utils.GetDefaultPaginationLimit())
	offset := utils.ParseQueryParamInt(r, "offset", 0)
	limit, offset = utils.NormalizePagination(limit, offset)

	facts, err := starks.GetRecentZtarknetFacts(limit, offset)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}

// GetStateTransition retrieves the state transition from old_state to new_state
func GetStateTransition(w http.ResponseWriter, r *http.Request) {
	if !starks.ShouldIndexZtarknet() {
		utils.WriteErrorJson(w, http.StatusNotFound, "Ztarknet indexing is disabled")
		return
	}

	oldState := utils.ParseQueryParam(r, "old_state", "")
	if oldState == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: old_state")
		return
	}

	newState := utils.ParseQueryParam(r, "new_state", "")
	if newState == "" {
		utils.WriteErrorJson(w, http.StatusBadRequest, "Missing required parameter: new_state")
		return
	}

	facts, err := starks.GetStateTransition(oldState, newState)
	if err != nil {
		utils.WriteErrorJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteDataJson(w, facts)
}
