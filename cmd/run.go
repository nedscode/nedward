package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/nedscode/nedward/common"
	"github.com/nedscode/nedward/config"
	"github.com/nedscode/nedward/runner"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:    "run",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPathFromWorkingDirectory()
		if err != nil {
			return errors.WithStack(err)
		}
		cfg, err := config.LoadConfig(configPath, common.NedwardVersion, logger)
		if err != nil {
			return errors.WithMessage(err, configPath)
		}

		r := &runner.Runner{
			Service: cfg.ServiceMap[args[0]],
		}
		r.NoWatch = *runFlags.noWatch
		r.WorkingDir = *runFlags.directory
		r.Logger = logger
		r.Run(args)
		return nil
	},
}

var runFlags struct {
	noWatch   *bool
	directory *string
}

func init() {
	RootCmd.AddCommand(runCmd)

	runFlags.noWatch = runCmd.Flags().Bool("no-watch", false, "Disable autorestart")
	runFlags.directory = runCmd.Flags().StringP("directory", "d", "", "Working directory")
}
