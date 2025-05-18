package middleware

import (
	"context"
	"net/http"
)

const UserDIDKey ContextKey = "user_did"

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// TODO: Add auth verification
		ctx := context.WithValue(r.Context(), UserDIDKey, "")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
