package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v1.0.0"

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update app to the newest version",
	Long:  "Look for the newest version and install it, if current is the latest version does nothing",
	Args:  cobra.NoArgs,
	Run:   updateApp,
}

func updateApp(cmd *cobra.Command, _ []string) {
	fmt.Printf("Current version: %v\n", Version)
	fmt.Println("Looking for updates...")
}
