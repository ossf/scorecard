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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/naveensrinivasan/scorecarddata/pkg/bigquery"
	"github.com/naveensrinivasan/scorecarddata/pkg/deps"
	"github.com/spf13/cobra"
)

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Parses go.mod dependecies and fetches the data from scorecard bigquery for those repositories.",
	Long: ` This will parse the go.mod using go list and extract the github.com dependecies.
	It uses the extracted dependencies to query the bigquery scorecard table to fetch the results.
and usage of using your command. 
For example:
scorecarddata go  -m /home/sammy/go/src/github.com/naveensrinivasan/kubernetes --GOOGLE_CLOUD_PROJECT openssf
`,
	Run: func(cmd *cobra.Command, args []string) {
		modFileLocation, err := cmd.Flags().GetString("go-mod-location")
		if err != nil {
			log.Fatal(err)
		}

		projectID, err := cmd.Flags().GetString("GOOGLE_CLOUD_PROJECT")
		if err != nil {
			log.Fatal(err)
		}
		// Parse the exclusionFile
		// The expectation of the exclusion file one line with key and value separated by comma.
		// Code-Review,github.com/ossf/scorecard
		exclusion := make(map[bigquery.Key]bool)
		exclusionFile, err := cmd.Flags().GetString("exclusions-file")
		if err == nil && exclusionFile != "" {
			data, err := ioutil.ReadFile(exclusionFile)
			if err != nil {
				log.Fatal(err)
			}
			lines := string(data)
			for _, line := range strings.Split(lines, "\n") {
				cols := strings.Split(line, ",")
				if len(cols) == 2 {
					exclusion[bigquery.Key{Check: cols[0], Repoistory: cols[1]}] = true
				}
			}
		}

		checks, err := cmd.Flags().GetStringArray("scorecard_checks")
		if err != nil {
			log.Fatal(err)
		}

		d := deps.NewGolangDeps()
		deps, err := d.FetchDependencies(modFileLocation)
		if err != nil {
			log.Fatal(err)
		}

		b := bigquery.NewBigquery(projectID)
		result, err := b.FetchScorecardData(deps, checks, exclusion)
		if err != nil {
			log.Fatal(err)
		}

		j, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(j))
	},
}

func init() {
	rootCmd.AddCommand(goCmd)
	goCmd.Flags().StringP("go-mod-location", "m", "", "The directory of the go.mod location")
	goCmd.MarkFlagRequired("go-mod-location")
}
