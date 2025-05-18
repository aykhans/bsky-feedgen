package response

import (
	"encoding/json"
	"net/http"

	"github.com/aykhans/bsky-feedgen/pkg/logger"
)

type M map[string]any

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func JSON500(w http.ResponseWriter) {
	JSON(w, 500, M{"error": "Internal server error"})
}

func JSON404(w http.ResponseWriter) {
	JSON(w, 404, M{"error": "Not found"})
}
