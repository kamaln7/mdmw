// Copyright © 2018 Kamal Nasser <hello@kamal.io>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kamaln7/mdmw/mdmw"
	"github.com/kamaln7/mdmw/mdmw/storage"
	"github.com/kamaln7/mdmw/mdmw/storage/filesystem"
	"github.com/kamaln7/mdmw/mdmw/storage/spaces"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mdmw",
	Short: "A drop-in markdown middleware",
	Run:   runMdmw,
}

const (
	argListenAddress = "listenaddress"

	argStorageDriver = "storage"

	argFilesystemPath = "filesystem.path"

	argSpacesAccessKey = "spaces.auth.access"
	argSpacesSecretKey = "spaces.auth.secret"
	argSpacesSpace     = "spaces.space"
	argSpacesRegion    = "spaces.region"
	argSpacesPath      = "spaces.path"
)

// Config contains the mdmw config
type Config struct {
	ListenAddress string
	Storage       string
	Filesystem    filesystem.Config
	Spaces        spaces.Config
}

var config Config

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addStringFlag(p *string, name, shorthand, value, usage string) {
	rootCmd.Flags().StringVarP(p, name, shorthand, value, usage)
	viper.BindPFlag(name, rootCmd.Flags().Lookup(name))
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.mdmw.yaml)")

	addStringFlag(&config.ListenAddress, argListenAddress, "", "localhost:4000", "address to listen on")
	addStringFlag(&config.Storage, argStorageDriver, "", "filesystem", "storage driver to use")

	// filesystem
	addStringFlag(&config.Filesystem.Path, argFilesystemPath, "", "./files", "path to markdown files")

	// spaces
	addStringFlag(&config.Spaces.Auth.Access, argSpacesAccessKey, "", "", "DigitalOcean Spaces access key")
	addStringFlag(&config.Spaces.Auth.Secret, argSpacesSecretKey, "", "", "DigitalOcean Spaces secret key")
	addStringFlag(&config.Spaces.Space, argSpacesSpace, "", "", "DigitalOcean Spaces space name")
	addStringFlag(&config.Spaces.Path, argSpacesPath, "", "/", "DigitalOcean Spaces files path")
	addStringFlag(&config.Spaces.Region, argSpacesRegion, "", "", "DigitalOcean Spaces region")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".mdmw" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName(".mdmw")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file: ", viper.ConfigFileUsed())
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't apply config: %v", err)
		os.Exit(1)
	}
}

func runMdmw(cmd *cobra.Command, args []string) {
	var sd storage.Driver

	switch config.Storage {
	case "filesystem":
		sd = &filesystem.Driver{Config: config.Filesystem}
		break
	case "spaces":
		spaces := &spaces.Driver{
			Config: config.Spaces,
		}
		spaces.Connect()

		sd = spaces
		break
	default:
		fmt.Fprintf(os.Stderr, "storage driver %s does not exist\n", config.Storage)
		os.Exit(1)
	}

	server := &mdmw.Server{
		ListenAddress: config.ListenAddress,
		StorageDriver: sd,
	}

	server.Listen()
}
