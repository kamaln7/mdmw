package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kamaln7/mdmw/mdmw/storage"
)

// Config contains the filesystem config
type Config struct {
	Path string
}

// Driver defines a filesystem-based storage driver
type Driver struct {
	Config Config
}

// ensure interface implementation
var _ storage.Driver = new(Driver)

func (d *Driver) Read(path string) ([]byte, error) {
	filePath := filepath.Join(d.Config.Path, path)

	content, err := ioutil.ReadFile(filePath)

	switch {
	case os.IsNotExist(err):
		return nil, storage.ErrNotFound
	case os.IsPermission(err):
		return nil, storage.ErrForbidden
	default:
		return content, err
	}
}

func (d *Driver) List(path string) ([]storage.File, error) {
	var files []storage.File

	filePath := filepath.Join(d.Config.Path, path)

	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		file := storage.File{
			Name: info.Name(),
			Path: fmt.Sprintf("%s/%s", path, info.Name()),
		}

		files = append(files, file)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
