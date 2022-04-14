// Copyright OpenSSF Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard-action/signing"
)

func main() {
	action, err := entrypoint.New()
	if err != nil {
		log.Fatalf("creating scorecard entrypoint: %v", err)
	}

	if err := action.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}

	if os.Getenv(options.EnvInputPublishResults) == "true" { //nolint
		sarifOutputFile := os.Getenv(options.EnvInputResultsFile)
		// Get sarif results from file.
		sarifPayload, err := ioutil.ReadFile(sarifOutputFile)
		if err != nil {
			log.Fatalf("error reading from sarif output file: %v", err)
		}

		// Sign sarif results.
		if err = signing.SignScorecardResult(sarifOutputFile); err != nil {
			log.Fatalf("error signing scorecard sarif results: %v", err)
		}

		// Get json results by re-running scorecard.
		jsonPayload, err := signing.GetJSONScorecardResults()
		if err != nil {
			log.Fatalf("error generating json scorecard results: %v", err)
		}

		// Sign json results.
		if err = signing.SignScorecardResult("results.json"); err != nil {
			log.Fatalf("error signing scorecard json results: %v", err)
		}

		// Processes sarif & json results.
		repoName := os.Getenv(options.EnvGithubRepository)
		repoRef := os.Getenv(options.EnvGithubRef)
		if err := signing.ProcessSignature(sarifPayload, jsonPayload, repoName, repoRef); err != nil {
			log.Fatalf("error processing signature: %v", err)
		}
	}
}
