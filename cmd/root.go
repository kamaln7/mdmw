package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

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
	argListenAddress     = "listenaddress"
	argOutputTemplate    = "outputtemplate"
	argValidateExtension = "validateextension"
	argRootListingType   = "rootlisting"
	argRootListingTitle  = "rootlistingtitle"
	argVerbose           = "verbose"

	argStorageDriver = "storage"

	argFilesystemPath = "filesystem.path"

	argSpacesAccessKey = "spaces.auth.access"
	argSpacesSecretKey = "spaces.auth.secret"
	argSpacesSpace     = "spaces.space"
	argSpacesRegion    = "spaces.region"
	argSpacesPath      = "spaces.path"
	argSpacesCache     = "spaces.cache"

	rootListingTypes = "title-case, files, off"
)

// Config contains the mdmw config
type Config struct {
	ListenAddress                     string
	Storage                           string
	Filesystem                        filesystem.Config
	Spaces                            spaces.Config
	SpacesCacheDuration               string
	ValidateExtension, Verbose        bool
	OutputTemplate                    string
	RootListingType, RootListingTitle string
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

func addBoolFlag(p *bool, name, shorthand string, value bool, usage string) {
	rootCmd.Flags().BoolVarP(p, name, shorthand, value, usage)
	viper.BindPFlag(name, rootCmd.Flags().Lookup(name))
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.mdmw.yaml)")

	addStringFlag(&config.ListenAddress, argListenAddress, "", "localhost:4000", "address to listen on")
	addStringFlag(&config.Storage, argStorageDriver, "", "filesystem", "storage driver to use")
	addBoolFlag(&config.ValidateExtension, argValidateExtension, "", true, "validate that files have a markdown extension")
	addStringFlag(&config.OutputTemplate, argOutputTemplate, "", "", "path to HTML output template")
	addStringFlag(&config.RootListingType, argRootListingType, "", "title-case", "show a file listing at /. options: ("+rootListingTypes+")")
	addStringFlag(&config.RootListingTitle, argRootListingTitle, "", "", "the title to use for the file listing")
	addBoolFlag(&config.Verbose, argVerbose, "", false, "log all incoming requests")

	// filesystem
	addStringFlag(&config.Filesystem.Path, argFilesystemPath, "", "./files", "path to markdown files")

	// spaces
	addStringFlag(&config.Spaces.Auth.Access, argSpacesAccessKey, "", "", "DigitalOcean Spaces access key")
	addStringFlag(&config.Spaces.Auth.Secret, argSpacesSecretKey, "", "", "DigitalOcean Spaces secret key")
	addStringFlag(&config.Spaces.Space, argSpacesSpace, "", "", "DigitalOcean Spaces space name")
	addStringFlag(&config.Spaces.Path, argSpacesPath, "", "/", "DigitalOcean Spaces files path")
	addStringFlag(&config.Spaces.Region, argSpacesRegion, "", "", "DigitalOcean Spaces region")
	addStringFlag(&config.SpacesCacheDuration, argSpacesCache, "", "0", "DigitalOcean Spaces cache time")
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
		duration, err := time.ParseDuration(config.SpacesCacheDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't parse time duration %s: %v\n", config.SpacesCacheDuration, err)
			os.Exit(1)
		}
		config.Spaces.Cache = duration

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

	var rootListing int
	switch config.RootListingType {
	case "title-case":
		rootListing = mdmw.ListingTitleCase
	case "files":
		rootListing = mdmw.ListingFiles
	case "off":
		rootListing = mdmw.ListingOff
	default:
		fmt.Fprintf(os.Stderr, "unknown root listing type %s. options: (%s)\n", config.RootListingType, rootListingTypes)
		os.Exit(1)
	}

	server := &mdmw.Server{
		ListenAddress:     config.ListenAddress,
		Storage:           sd,
		ValidateExtension: config.ValidateExtension,
		RootListing:       rootListing,
		RootListingTitle:  config.RootListingTitle,
		Verbose:           config.Verbose,
	}

	tmpl := config.OutputTemplate
	if config.OutputTemplate != "" {
		source, err := ioutil.ReadFile(config.OutputTemplate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't read template file %s: %v\n", config.OutputTemplate, err)
			os.Exit(1)
		}

		tmpl = string(source)
	}

	err := server.SetOutputTemplate(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't set template: %v\n", err)
		os.Exit(1)
	}

	server.Listen()
}
