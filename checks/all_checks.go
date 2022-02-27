// Copyright 2020 Security Scorecard Authors
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

// Package checks defines all Scorecard checks.
package checks

import (
	"github.com/ossf/scorecard/v4/checker"
)

// AllChecks is the list of all security checks that will be run.
var AllChecks = checker.CheckNameToFnMap{}

// GetAll returns the full list of checks, given any environment variable
// constraints.
// TODO(checks): Is this actually necessary given `AllChecks` exists?
func GetAll() checker.CheckNameToFnMap {
	possibleChecks := AllChecks
	return possibleChecks
}

func registerCheck(name string, fn checker.CheckFn, supportedRequestTypes []checker.RequestType) error {
	if name == "" {
		return errInternalNameCannotBeEmpty
	}
	if fn == nil {
		return errInternalCheckFuncCannotBeNil
	}
	AllChecks[name] = checker.Check{
		Fn:                    fn,
		SupportedRequestTypes: supportedRequestTypes,
	}
	return nil
}
