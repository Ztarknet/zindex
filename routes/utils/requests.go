package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
