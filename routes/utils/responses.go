package utils

import (
	"encoding/json"
	"net/http"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
)

type DataResponse struct {
	Data interface{} `json:"data"`
}

type ResultResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteDataJson(w http.ResponseWriter, data interface{}) {
	SetCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := DataResponse{Data: data}
	json.NewEncoder(w).Encode(response)
}

func WriteResultJson(w http.ResponseWriter, result string) {
	SetCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := ResultResponse{Result: result}
	json.NewEncoder(w).Encode(response)
}

func WriteErrorJson(w http.ResponseWriter, statusCode int, errorMsg string) {
	SetCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: errorMsg}
	json.NewEncoder(w).Encode(response)
}

func BasicErrorJson(errorMsg string) ErrorResponse {
	return ErrorResponse{Error: errorMsg}
}

func SetCorsHeaders(w http.ResponseWriter) {
	if len(config.Conf.Api.Cors.AllowedOrigins) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", config.Conf.Api.Cors.AllowedOrigins[0])
	}

	if len(config.Conf.Api.Cors.AllowedMethods) > 0 {
		methods := ""
		for i, method := range config.Conf.Api.Cors.AllowedMethods {
			if i > 0 {
				methods += ", "
			}
			methods += method
		}
		w.Header().Set("Access-Control-Allow-Methods", methods)
	}

	if len(config.Conf.Api.Cors.AllowedHeaders) > 0 {
		headers := ""
		for i, header := range config.Conf.Api.Cors.AllowedHeaders {
			if i > 0 {
				headers += ", "
			}
			headers += header
		}
		w.Header().Set("Access-Control-Allow-Headers", headers)
	}
}
