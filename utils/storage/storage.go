package storage

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrDatabaseError = errors.New("storage: database error")
)