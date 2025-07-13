package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/alisaviation/pkg/logger"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondWithToken(w http.ResponseWriter, code int, message, token string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
		"token":   token,
	})
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}, context ...zap.Field) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log.Error("Failed to encode JSON response",
			append(context, zap.Error(err))...)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
