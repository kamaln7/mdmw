package filesystem

import (
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

	if os.IsNotExist(err) {
		return nil, storage.ErrNotFound
	}

	return content, err
}
