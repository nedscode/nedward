package cmd

import (
	"fmt"

	"github.com/nedscode/nedward/runner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:    "run",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := nedwardClient.ServiceMap()[args[0]]
		if service == nil {
			return fmt.Errorf("service not found: %s", args[0])
		}
		r := &runner.Runner{
			Service: service,
		}
		r.NoWatch = *runFlags.noWatch
		r.WorkingDir = *runFlags.directory
		r.Logger = logger
		err := r.Run(args)
		return errors.WithStack(err)
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
	_ = runCmd.Flags().StringArrayP("tag", "t", nil, "Tags to distinguish this instance of Edward")
}
