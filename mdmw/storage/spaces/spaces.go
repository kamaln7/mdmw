package spaces

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/kamaln7/mdmw/mdmw/storage"
)

// Config contains the Spaces credentials and config
type Config struct {
	Auth struct {
		Access, Secret string
	}
	Region, Space string
	Path          string
}

// Driver defines a filesystem-based storage driver
type Driver struct {
	Config Config

	spaces *s3.S3
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
}

func (d *Driver) Read(path string) ([]byte, error) {
	filePath := filepath.Join(d.Config.Path, path)

	output, err := d.spaces.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.Config.Space),
		Key:    aws.String(filePath),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, storage.ErrNotFound
		}

		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)

	return buf.Bytes(), nil
}
