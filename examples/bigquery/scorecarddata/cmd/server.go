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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/naveensrinivasan/scorecarddata/pkg/bigquery"
	"github.com/spf13/cobra"
)

type Request struct {
	Repositories []string `json:"repositories"`
	Checks       []string `json:"checks"`
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTP Server which listens in 8080 port",
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func Run() {
	router := gin.Default()
	router.POST("/go", func(c *gin.Context) {
		var json Request
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		bq := bigquery.NewBigquery("openssf")
		result, err := bq.FetchScorecardData(json.Repositories, json.Checks, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"result": result})
	})
	router.Run("127.0.0.1:8080")
}
