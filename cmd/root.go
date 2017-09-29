package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/nedscode/nedward/common"
	"github.com/nedscode/nedward/config"
	"github.com/nedscode/nedward/nedward"
	"github.com/nedscode/nedward/home"
	"github.com/nedscode/nedward/services"
	"github.com/nedscode/nedward/updates"
)

var cfgFile string

var nedwardClient *nedward.Client

var logger *log.Logger

var checkUpdateChan chan interface{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "nedward",
	Short: "A tool for managing local instances of microservices",
	Long: `Nedward is a tool to simplify your microservices development workflow.
Build, start and manage service instances with a single command.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Begin logging
		logger.Printf("=== Nedward v%v ===\n", common.NedwardVersion)
		logger.Printf("Args: %v\n", os.Args)

		nedwardClient = nedward.NewClient()
		// Set the default config path
		if configPath != "" {
			nedwardClient.Config = configPath
		} else {
			nedwardClient.Config = getConfigPath()
		}
		// Set service checks to restart the client on sudo as needed
		nedwardClient.ServiceChecks = func(sgs []services.ServiceOrGroup) error {
			return errors.WithStack(sudoIfNeeded(sgs))
		}
		nedwardClient.Logger = logger
		// Populate the Nedward executable with this binary
		nedwardClient.NedwardExecutable = os.Args[0]

		command := cmd.Use

		if command != "generate" {
			err := config.LoadSharedConfig(nedwardClient.Config, common.NedwardVersion, logger)
			if err != nil {
				return errors.WithStack(err)
			}
			err = os.Chdir(config.GetBasePath())
			if err != nil {
				return errors.WithStack(err)
			}
		} else {
			config.InitEmptyConfig()
		}

		if command != "stop" {
			// Check for legacy pidfiles and error out if any are found
			for _, service := range config.GetServiceMap() {
				if _, err := os.Stat(service.GetPidPathLegacy()); !os.IsNotExist(err) {
					return errors.New("one or more services were started with an older version of Nedward. Please run `nedward stop` to stop these instances")
				}
			}
		}

		if command != "run" {
			checkUpdateChan = make(chan interface{})
			go checkUpdateAvailable(checkUpdateChan)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		defer logger.Printf("=== Exiting ===\n")
		if checkUpdateChan != nil { //&& !didAutoComplete { // TODO: skip when autocompleting
			updateAvailable, ok := (<-checkUpdateChan).(bool)
			if ok && updateAvailable {
				latestVersion := (<-checkUpdateChan).(string)
				fmt.Printf("A new version of Nedward is available (%v), update with:\n\tgo get -u github.com/nedscode/nedward\n", latestVersion)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := home.NedwardConfig.Initialize(); err != nil {
		fmt.Printf("%+v", err)
	}

	logger = log.New(&lumberjack.Logger{
		Filename:   filepath.Join(home.NedwardConfig.NedwardLogDir, "nedward.log"),
		MaxSize:    50, // megabytes
		MaxBackups: 30,
		MaxAge:     1, //days
	}, fmt.Sprintf("%v >", os.Args), log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	for _, arg := range os.Args {
		if arg == "--generate-bash-completion" {
			autocompleteServicesAndGroups(logger)
			return
		}
	}

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var configPath string

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Use service configuration file at `PATH`")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra-start" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra-start")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// getConfigPath identifies the location of nedward.json, if any exists
func getConfigPath() string {
	var pathOptions []string

	// Config file in Nedward Config dir
	pathOptions = append(pathOptions, filepath.Join(home.NedwardConfig.Dir, "nedward.json"))

	// Config file in current working directory
	wd, err := os.Getwd()
	if err == nil {
		pathOptions = append(pathOptions, filepath.Join(wd, "nedward.json"))
	}
	for path.Dir(wd) != wd {
		wd = path.Dir(wd)
		pathOptions = append(pathOptions, filepath.Join(wd, "nedward.json"))
	}

	for _, path := range pathOptions {
		_, err := os.Stat(path)
		if err != nil {
			continue
		}
		absfp, absErr := filepath.Abs(path)
		if absErr != nil {
			fmt.Println("Error getting config file: ", absErr)
			return ""
		}
		return absfp
	}

	return ""
}

func checkUpdateAvailable(checkUpdateChan chan interface{}) {
	defer close(checkUpdateChan)
	updateAvailable, latestVersion, err := updates.UpdateAvailable("github.com/nedscode/nedward", common.NedwardVersion, filepath.Join(home.NedwardConfig.Dir, ".updatecache"), logger)
	if err != nil {
		logger.Println("Error checking for updates:", err)
		return
	}

	checkUpdateChan <- updateAvailable
	if updateAvailable {
		checkUpdateChan <- latestVersion
	}
}
