package cmd

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var Version = "v1.0.0"

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update app to the latest version",
	Long:  "Look for the latest version and install it",
	Args:  cobra.NoArgs,
	RunE:  updateApp,
}

func updateApp(cmd *cobra.Command, _ []string) error {
	fmt.Println("Looking for updates...")
	v, err := semver.Parse(Version)
	if err != nil {
		return fmt.Errorf("Error procesing current version: %v", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return fmt.Errorf("Error launching updater: %v", err)
	}

	latest, err := updater.UpdateSelf(v, "jpabloVega/dodeyo")
	if err != nil {
		return fmt.Errorf("Error updating: %v", err)
	}

	if latest.Version.Equals(v) {
		fmt.Println("Current version is the latest")
	} else {
		fmt.Printf("Version %v installed successfully\n", latest.Version)
	}
	return nil
}
