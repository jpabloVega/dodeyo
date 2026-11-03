package cmd

package cmd

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Check current version",
	Long:  "Check current version and if updates are availible",
	Args:  cobra.NoArgs,
	RunE:  updateApp,
}

func versionApp(cmd *cobra.Command, _ []string) error {
	v, err := semver.Parse(Version)
	if err != nil {
		return fmt.Errorf("Error procesing current version: %v", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return fmt.Errorf("Error launching updater: %v", err)
	}

	fmt.Printf("Current version: %v\n", v)

	if latest.Version.Equals(v) {
		fmt.Println("Current version is the latest\n")
	} else {
		fmt.Printf("New version availible: %v\n", latest.Version)
	}
	return nil
}
