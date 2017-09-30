package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nedscode/nedward/common"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the currently installed version of Nedward",
	// Skip loading config
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Nedward version %v\n", common.NedwardVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
