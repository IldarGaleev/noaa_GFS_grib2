package storage

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrDatabaseError = errors.New("storage: database error")
	ErrBatchSize     = errors.New("storage: too long batch size")
)
