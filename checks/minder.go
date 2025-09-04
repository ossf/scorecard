// Copyright 2025 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/minder"
)

func init() {
	IngestToType := map[string][]checker.RequestType{
		"rest": nil,
		"git":  {checker.CommitBased, checker.FileBased},
	}

	for _, rule := range minder.AllRules {
		rulefunc := minder.CheckRule(rule)
		err := registerCheck(rule.GetName(), rulefunc, IngestToType[rule.GetDef().GetIngest().GetType()])
		if err != nil {
			panic("Failed to register minder rule: " + rule.GetName() + ": " + err.Error())
		}
	}
}
