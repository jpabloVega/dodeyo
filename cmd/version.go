package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Current version",
	Long:  "Check the version currently installed",
	Args:  cobra.NoArgs,
	Run:   versionApp,
}

func versionApp(cmd *cobra.Command, _ []string) {
	fmt.Printf("Current version: %v\n", Version)
}
