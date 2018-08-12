package storage

import "errors"

// Driver defines a storage driver
type Driver interface {
	Read(path string) ([]byte, error)
}

var (
	ErrNotFound  = errors.New("file not found")
	ErrForbidden = errors.New("could not read file")
)
