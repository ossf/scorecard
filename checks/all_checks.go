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
	"os"

	"github.com/ossf/scorecard/v4/checker"
)

// allChecks is the list of all registered security checks
var allChecks = checker.CheckNameToFnMap{}

func getAll(overrideExperimental bool) checker.CheckNameToFnMap {
	// need to make a copy or caller could mutate original map
	possibleChecks := checker.CheckNameToFnMap{}
	for k, v := range allChecks {
		possibleChecks[k] = v
	}

	if overrideExperimental {
		return possibleChecks
	}

	// TODO: remove this check when v6 is released
	if _, v6 := os.LookupEnv("SCORECARD_V6"); !v6 {
		delete(possibleChecks, CheckWebHooks)
	}

	return possibleChecks
}

// GetAllForEnvironment returns the full list of checks, given any environment variable
// constraints.
func GetAllForEnvironment() checker.CheckNameToFnMap {
	return getAll(false /*overrideExperimental*/)
}

// GetAll returns the full list of checks, regardless of
// environment variable constraints.
func GetAll() checker.CheckNameToFnMap {
	return getAll(true /*overrideExperimental*/)
}

func registerCheck(name string, fn checker.CheckFn, supportedRequestTypes []checker.RequestType) error {
	if name == "" {
		return errInternalNameCannotBeEmpty
	}
	if fn == nil {
		return errInternalCheckFuncCannotBeNil
	}
	allChecks[name] = checker.Check{
		Fn:                    fn,
		SupportedRequestTypes: supportedRequestTypes,
	}
	return nil
}
