package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/keep-starknet-strange/ztarknet/zindex/internal/config"
)

func ReadJsonBody[T any](r *http.Request) (*T, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	var body T
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	return &body, nil
}

func ParseQueryParam(r *http.Request, param string, defaultValue string) string {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}
	return value
}

func ParseQueryParamInt(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetDefaultPaginationLimit returns the default pagination limit from config
func GetDefaultPaginationLimit() int {
	return config.Conf.Api.Pagination.DefaultLimit
}

// NormalizePagination validates and normalizes limit and offset parameters
// Uses pagination configuration from config (max_limit and max_offset)
func NormalizePagination(limit, offset int) (int, int) {
	maxLimit := config.Conf.Api.Pagination.MaxLimit
	maxOffset := config.Conf.Api.Pagination.MaxOffset

	// Cap limit at maxLimit from config
	if limit > maxLimit {
		limit = maxLimit
	}
	if limit < 1 {
		limit = 1
	}

	// Ensure offset is non-negative and within max_offset
	if offset < 0 {
		offset = 0
	}
	if maxOffset > 0 && offset > maxOffset {
		offset = maxOffset
	}

	return limit, offset
}

// NormalizePaginationWithMax validates and normalizes limit and offset with a custom max limit
// This is kept for backward compatibility but uses config for max_offset
func NormalizePaginationWithMax(limit, offset, maxLimit int) (int, int) {
	maxOffset := config.Conf.Api.Pagination.MaxOffset

	// Cap limit at maxLimit
	if limit > maxLimit {
		limit = maxLimit
	}
	if limit < 1 {
		limit = 1
	}

	// Ensure offset is non-negative and within max_offset from config
	if offset < 0 {
		offset = 0
	}
	if maxOffset > 0 && offset > maxOffset {
		offset = maxOffset
	}

	return limit, offset
}

// ParseCommaSeparated splits a comma-separated string into a slice of trimmed strings
func ParseCommaSeparated(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
