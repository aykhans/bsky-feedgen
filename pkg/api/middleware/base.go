package middleware

import (
	"net/http"

	"github.com/aykhans/bsky-feedgen/pkg/types"
)

type ContextKey string

func GetValue[T any](r *http.Request, key ContextKey) (T, error) {
	value, ok := r.Context().Value(key).(T)
	if ok == false {
		var zero T
		return zero, types.ErrNotfound
	}

	return value, nil
}
