package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jpabloVega/dodeyo/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	userLicense string
)

var rootCmd = &cobra.Command{
	Use:   "dodeyo",
	Short: "Export you favorite novels from syosetsu.",
	Long: `Export your favorite novels from syosetsu
Save it locally, on your eReader or mail it to your kindle
Scan them to know the jlpt level of the vocabulary
Add furigana for an easy read`,
	RunE: GetContents,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "Juan Pablo Vega Pérez <vega_perezpablo@hotmail.com>")
	viper.SetDefault("license", "apache")

}

func initConfig() {
	err := database.StartDB()
	cobra.CheckErr(err)

	err = startCfgFile()
	cobra.CheckErr(err)

	configDir, _ := os.UserConfigDir()
	appDir := filepath.Join(configDir, "Dodeyo")

	//viper.AddConfigPath(".")
	viper.AddConfigPath(appDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".cobra")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error loading config file")
		cobra.CheckErr(err)
	}

	viper.SetDefault("defaultpath", "")
	viper.SetDefault("kindlemail", "")
	viper.SetDefault("usermail", "")
	viper.SetDefault("proxy", "")
	viper.SetDefault("license", "MIT")
	viper.SetDefault("useviper", true)

	err = database.StartDB()
	if err != nil {
		fmt.Println(err)
	}
}

func startCfgFile() error {
	configDir, _ := os.UserConfigDir()
	appDir := filepath.Join(configDir, "Dodeyo")
	cfgPath := filepath.Join(appDir, ".cobra.yaml")
	_, err := os.Stat(cfgPath)
	if errors.Is(err, os.ErrNotExist) {
		bareCfg := `defaultpath: ""
kindlemail: ""
license: MIT
usermail: ""
useviper: true
proxy: ""`
		err := os.WriteFile(cfgPath, []byte(bareCfg), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
