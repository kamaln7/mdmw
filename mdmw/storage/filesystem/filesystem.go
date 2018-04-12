package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kamaln7/mdmw/mdmw/storage"
)

// Driver defines a filesystem-based storage driver
type Driver struct {
	Path string
}

// ensure interface implementation
var _ storage.Driver = new(Driver)

func (d *Driver) Read(path string) ([]byte, error) {
	filePath := filepath.Join(d.Path, path)

	content, err := ioutil.ReadFile(filePath)

	if os.IsNotExist(err) {
		return nil, storage.ErrNotFound
	}

	return content, err
}
