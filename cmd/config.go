package cmd

import (
	"fmt"

	"github.com/jpabloVega/dokodemoYomou/secure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	setEmail           bool
	showConfig         bool
	proxy              string
	removeProxy        bool
	newDestinationPath string
	kindleEmail        string
)

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolVar(&setEmail, "email", false, "Access the new email menu")
	configCmd.Flags().BoolVar(&showConfig, "show", false, "Show current config")
	configCmd.Flags().StringVar(&proxy, "proxy", "", "Set proxy for requests")
	configCmd.Flags().BoolVar(&removeProxy, "removeProxy", false, "Remove the current config proxy")
	configCmd.Flags().StringVar(&newDestinationPath, "defaultPath", "", "Default path to save content")
	configCmd.Flags().StringVar(&kindleEmail, "kindleEmail", "", "Set your kindle address")

}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set Default config values",
	Long:  "Set Default behaviour for the app",
	Args:  cobra.NoArgs,
	RunE:  setConfig,
}

func setConfig(_ *cobra.Command, _ []string) error {
	if proxy != "" {
		viper.Set("proxy", proxy)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	if removeProxy {
		viper.Set("proxy", "")
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	if newDestinationPath != "" {
		err := checkValidPath(newDestinationPath)
		if err != nil {
			return err
		}
	}
	if kindleEmail != "" {
		viper.Set("kindlemail", kindleEmail)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	if setEmail {
		_, _, err := secure.NewEmailPassword()
		if err != nil {
			return err
		}
	}

	if showConfig {
		defaultPath := viper.GetString("defaultPath")
		if defaultPath == "" {
			defaultPath = "<not set>"
		}
		fmt.Printf("Default path: %s\n", defaultPath)
		usermail := viper.GetString("usermail")
		if usermail == "" {
			usermail = "<not set>"
		}
		fmt.Printf("Send from mail: %s\n", usermail)
		kindlemail := viper.GetString("kindlemail")
		if kindlemail == "" {
			kindlemail = "<not set>"
		}
		fmt.Printf("Send to mail: %s\n", kindlemail)
		currProxy := viper.GetString("proxy")
		if currProxy != "" {
			fmt.Printf("Active proxy: %s\n", currProxy)
		} else {
			fmt.Println("No active proxy")
		}
	}
	return nil
}
