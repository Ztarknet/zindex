package utils

import (
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
)

func NonProductionMiddleware(w http.ResponseWriter, r *http.Request) bool {
	if config.Conf.Api.Production {
		WriteErrorJson(w, http.StatusNotImplemented, "This endpoint is not available in production mode")
		return true
	}
	return false
}

func AuthMiddleware(w http.ResponseWriter, r *http.Request) bool {
	return false
}

func AdminMiddleware(w http.ResponseWriter, r *http.Request) bool {
	if !config.Conf.Api.Admin {
		WriteErrorJson(w, http.StatusUnauthorized, "Admin access required")
		return true
	}
	return false
}
