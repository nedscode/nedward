package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/nedscode/nedward/common"
	"github.com/nedscode/nedward/config"
)

func autocompleteServicesAndGroups(logger *log.Logger) {
	printCommandChildren(RootCmd)

	err := config.LoadSharedConfig(getConfigPath(), common.NedwardVersion, logger)
	if err != nil {
		logger.Println("Autocomplete> Error loading config:", err)
	}
	names := append(config.GetAllGroupNames(), config.GetAllServiceNames()...)
	for _, name := range names {
		fmt.Println(name)
	}
}

func printCommandChildren(cmd *cobra.Command) {
	for _, cmd := range cmd.Commands() {
		fmt.Println(cmd.Use)
		printCommandChildren(cmd)
	}
}
