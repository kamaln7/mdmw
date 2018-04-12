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

	"github.com/kamaln7/mdmw/mdmw"
	"github.com/kamaln7/mdmw/mdmw/storage"
	"github.com/kamaln7/mdmw/mdmw/storage/filesystem"
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
	argListenAddress  = "listen.address"
	argStorageDriver  = "storage.driver"
	argFilesystemPath = "filesystem.path"
)

// args
var (
	listenAddress  string
	storageDriver  string
	filesystemPath string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.mdmw.yaml)")

	rootCmd.Flags().StringVarP(&listenAddress, argListenAddress, "", "localhost:4000", "address to listen on")
	viper.BindPFlag(argListenAddress, rootCmd.Flags().Lookup(argListenAddress))

	rootCmd.Flags().StringVarP(&storageDriver, argStorageDriver, "", "filesystem", "storage driver to use")
	viper.BindPFlag(argStorageDriver, rootCmd.Flags().Lookup(argStorageDriver))

	rootCmd.Flags().StringVarP(&filesystemPath, argFilesystemPath, "", "./files", "path to markdown files")
	viper.BindPFlag(argFilesystemPath, rootCmd.Flags().Lookup(argFilesystemPath))
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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	}
}

func runMdmw(cmd *cobra.Command, args []string) {
	var sd storage.Driver

	switch storageDriver {
	case "filesystem":
		sd = &filesystem.Driver{Path: filesystemPath}
		break
	default:
		fmt.Fprintf(os.Stderr, "storage driver %s does not exist\n", storageDriver)
		os.Exit(1)
	}

	server := &mdmw.Server{
		ListenAddress: listenAddress,
		StorageDriver: sd,
	}

	server.Listen()
}
