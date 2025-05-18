package response

import (
	"net/http"

	"github.com/aykhans/bsky-feedgen/pkg/logger"
)

func Text(w http.ResponseWriter, statusCode int, content []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	if _, err := w.Write(content); err != nil {
		logger.Log.Error("Failed to write text response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func Text404(w http.ResponseWriter) {
	Text(w, 404, []byte("Not found"))
}

func Text500(w http.ResponseWriter) {
	Text(w, 500, []byte("Internal server error"))
}
