/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scorecarddata",
	Short: "scorecarddata uses the scorecard bigquery to query for data.",
	Long:  `scorecarddata uses the scorecard bigquery to fetch results for dependecies.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	projectID := ""
	ex := ""
	checks := []string{"Code-Review", "Branch-Protection", "Pinned-Dependencies", "Dependency-Update-Tool", "Fuzzing"}
	rootCmd.PersistentFlags().StringVar(&projectID, "GOOGLE_CLOUD_PROJECT", "",
		"The ENV variable that will be used in the BigQuery for querying.")
	rootCmd.MarkPersistentFlagRequired("GOOGLE_CLOUD_PROJECT")
	rootCmd.PersistentFlags().StringArray("scorecard_checks", checks,
		"The scorecard checks to filter by.Example CI-Tests,Binary-Artifacts etc.https://github.com/ossf/scorecard/blob/main/docs/checks.md")

	rootCmd.PersistentFlags().String("exclusions-file", ex, "A file with exclusions comma separated by check and value. Example Code-Review,github.com/ossf/scorecard")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.scorecarddata.yaml)")

	rootCmd.AddCommand()
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".scorecarddata" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".scorecarddata")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
