package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "i18ngen",
	Short: "i18ngen is a code generator for i18n message and placeholders",
}

// Execute runs the root command.
func Execute() {
	// Add generate command
	rootCmd.AddCommand(NewGenerateCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
