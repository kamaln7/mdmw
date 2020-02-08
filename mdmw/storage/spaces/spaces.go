package spaces

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	cache "github.com/patrickmn/go-cache"

	"github.com/kamaln7/mdmw/mdmw/storage"
)

// Config contains the Spaces credentials and config
type Config struct {
	Auth struct {
		Access, Secret string
	}
	Region, Space, Path string
	Cache               time.Duration
}

// Driver defines a filesystem-based storage driver
type Driver struct {
	Config Config

	spaces *s3.S3
	cache  *cache.Cache
}

// ensure interface implementation
var _ storage.Driver = new(Driver)

// Connect setus up the Spaces client
func (d *Driver) Connect() {
	spacesSession := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(d.Config.Auth.Access, d.Config.Auth.Secret, ""),
		Endpoint:    aws.String(fmt.Sprintf("https://%s.digitaloceanspaces.com", d.Config.Region)),
		Region:      aws.String("us-east-1"), // Needs to be us-east-1, or it'll fail.
	})

	d.spaces = s3.New(spacesSession)

	if d.Config.Cache != 0 {
		d.cache = cache.New(d.Config.Cache, d.Config.Cache)
	}
}

func (d *Driver) Read(path string) ([]byte, error) {
	filePath := filepath.Join(d.Config.Path, path)

	if d.cache == nil {
		return d.fetchFromSpaces(filePath)
	}

	cachedRes, cached := d.cache.Get(filePath)
	if cached {
		return cachedRes.([]byte), nil
	}

	content, err := d.fetchFromSpaces(filePath)
	if err != nil {
		return nil, err
	}

	d.cache.Set(filePath, content, cache.DefaultExpiration)
	return content, nil
}

func (d *Driver) fetchFromSpaces(path string) ([]byte, error) {
	if strings.HasSuffix(path, "/") {
		// refuse to read directories
		return nil, storage.ErrNotFound
	}

	output, err := d.spaces.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.Config.Space),
		Key:    aws.String(path),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, storage.ErrNotFound
			case "InvalidAccessKeyId":
				return nil, storage.ErrForbidden
			default:
				return nil, aerr
			}
		}

		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)

	return buf.Bytes(), nil
}

func (d *Driver) List(path string) ([]storage.File, error) {
	var files []storage.File
	path = strings.TrimPrefix(path, "/")

	out, err := d.spaces.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(d.Config.Space),
		Prefix:    aws.String(path),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	prefixes := make(map[string]interface{}, 0)
	for _, cp := range out.CommonPrefixes {
		name := strings.TrimRight(*(cp.Prefix), "/")
		prefixes[name] = struct{}{}
	}

	for _, obj := range out.Contents {
		name := *obj.Key
		if _, isDir := prefixes[name]; isDir {
			continue
		}

		files = append(files, storage.File{
			Name: name,
			Path: fmt.Sprintf("%s/%s", path, name),
		})
	}

	return files, nil
}
