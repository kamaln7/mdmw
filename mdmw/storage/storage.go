package storage

import (
	"errors"
)

// Driver defines a storage driver
type Driver interface {
	Read(path string) ([]byte, error)
	List(path string) ([]File, error)
}

type File struct {
	Name string
	Path string
}

var (
	ErrNotFound  = errors.New("file not found")
	ErrForbidden = errors.New("could not read file")
)
