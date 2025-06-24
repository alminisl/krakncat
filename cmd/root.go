package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "krakn",
	Short: "krakncat CLI tool for managing GitHub accounts",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip migration check for help commands and migrate command itself
		if cmd.Name() == "help" || cmd.Name() == "migrate" || cmd.Parent() != nil && cmd.Parent().Name() == "help" {
			return
		}
		
		// Run migration check
		if err := checkAndOfferMigration(); err != nil {
			// Don't fail the command if migration fails, just warn
			// This ensures the tool still works even if migration has issues
		}
	},
}

func Execute() error {
	return RootCmd.Execute()
}
