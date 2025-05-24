package types

import "errors"

var (
	ErrInternal = errors.New("internal error")
	ErrNotfound = errors.New("not found")
)
