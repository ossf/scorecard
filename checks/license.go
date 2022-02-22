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

package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
)

type check func(str string, extCheck []string) bool

type checks struct {
	rstr string // regex string
	f    check
	p    []string
}

// CheckLicense is the registered name for License.
const CheckLicense = "License"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
	}
	if err := registerCheck(CheckLicense, LicenseCheck, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

const (
	copying      = "copy(ing|right)"
	license      = "(un)?licen[sc]e"
	preferredExt = "*\\.(md|markdown|html)$"
	anyExt       = ".[^./]"
	ofl          = "ofl"
	patents      = "patents"
)

// Regex converted from
// https://github.com/licensee/licensee/blob/master/lib/licensee/project_files/license_file.rb
var (
	extensions  = []string{"xml", "go", "gemspec"}
	regexChecks = []checks{
		{rstr: copying, f: nil},
		{rstr: license, f: nil},
		{rstr: license + preferredExt, f: nil},
		{rstr: copying + preferredExt, f: nil},
		{rstr: copying + anyExt, f: nil},
		{rstr: ofl, f: nil},
		{rstr: ofl + preferredExt, f: nil},
		{rstr: patents, f: nil},
		{rstr: license, f: extensionMatch, p: []string{"spdx", "header"}},
		{rstr: license + "[-_][^.]*", f: extensionMatch, p: extensions},
		{rstr: copying + "[-_][^.]*", f: extensionMatch, p: extensions},
		{rstr: "\\w+[-_]" + license + "[^.]*", f: extensionMatch, p: extensions},
		{rstr: "\\w+[-_]" + copying + "[^.]*", f: extensionMatch, p: extensions},
		{rstr: ofl, f: extensionMatch, p: extensions},
	}
)

// ExtensionMatch to check for matching extension.
func extensionMatch(f string, exts []string) bool {
	s := strings.Split(f, ".")

	if len(s) <= 1 {
		return false
	}

	fext := s[len(s)-1]

	found := false
	for _, ext := range exts {
		if ext == fext {
			found = true
			break
		}
	}

	return found
}

// TestLicenseCheck used for testing purposes.
func testLicenseCheck(name string) bool {
	return checkLicense(name)
}

// LicenseCheck runs LicenseCheck check.
func LicenseCheck(c *checker.CheckRequest) checker.CheckResult {
	var s string

	err := fileparser.OnAllFilesDo(c.RepoClient, isLicenseFile, &s)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckLicense, err)
	}
	if s != "" {
		c.Dlogger.Info(&checker.LogMessage{
			Path:   s,
			Type:   checker.FileTypeSource,
			Offset: 1,
		})
		return checker.CreateMaxScoreResult(CheckLicense, "license file detected")
	}
	return checker.CreateMinScoreResult(CheckLicense, "license file not detected")
}

var isLicenseFile fileparser.DoWhileTrueOnFilename = func(name string, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("isLicenseFile requires exactly one argument: %w", errInvalidArgLength)
	}
	s, ok := args[0].(*string)
	if !ok {
		return false, fmt.Errorf("isLicenseFile requires argument of type: *string: %w", errInvalidArgType)
	}
	if checkLicense(name) {
		if s != nil {
			*s = name
		}
		return false, nil
	}
	return true, nil
}

// CheckLicense to check whether the name parameter fulfill license file criteria.
func checkLicense(name string) bool {
	for _, check := range regexChecks {
		rg := regexp.MustCompile(check.rstr)

		nameLower := strings.ToLower(name)
		t := rg.MatchString(nameLower)
		if t {
			extFound := true

			// check extension calling f function.
			// f function will always be func extensionMatch(..)
			if check.f != nil {
				extFound = check.f(nameLower, check.p)
			}

			return extFound
		}
	}
	return false
}
