// Copyright 2021 Security Scorecard Authors
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

package checks

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckDangerousWorkflow is the exported name for Dangerous-Workflow check.
const CheckDangerousWorkflow = "Dangerous-Workflow"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckDangerousWorkflow, DangerousWorkflow, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// DangerousWorkflow  will check the repository contains Dangerous-Workflow.
func DangerousWorkflow(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.DangerousWorkflow(c.RepoClient)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckDangerousWorkflow, e)
	}

	// Return raw results.
	if c.RawResults != nil {
		c.RawResults.DangerousWorkflowResults = rawData
	}

	j, _ := json.Marshal(rawData)
	fmt.Println(string(j))
	os.Exit(0)
	// TODO: use the helper API to populate c.RawResults.DangerousWorkflowData.*
	// X := helper.SecretsInPullRequests(&rawData)

	// Return the score evaluation.
	// return evaluation.DangerousWorkflow(CheckDangerousWorkflow, c.Dlogger,
	//&c.RawResults.DangerousWorkflowResults)
	return checker.CheckResult{}
}
