package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponseBody struct {
	Error string `json:"error"`
}

func OKResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Headers already sent, write plain error message
		_, _ = w.Write([]byte(`{"error":"encoding failed"}`))
	}
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := ErrorResponseBody{Error: message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Headers already sent, write plain error message
		_, _ = w.Write([]byte(`{"error":"encoding failed"}`))
	}
}
