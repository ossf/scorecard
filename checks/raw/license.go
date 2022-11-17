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

package raw

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
)

// from checks.md
//   - files must be at the top-level directory (hence the '^' in the regex)
//   - files must be of the name like COPY[ING|RIGHT] or LICEN[SC]E(plural)
//     no preceding names or the like (again, hence the '^')
//   - a folder is also acceptable as in COPYRIGHT/ or LICENSES/ although
//     the contents of the folder are totally ignored at this time
//   - file or folder the contents are not (yet) examined
//   - it is a case insenstive check (hence the leading regex compiler option).
//
// TODO: still working on matching this pattern: GPL-2.0-LICENSE or LICENSE_RELEASE.
// TODO: integrate `(([.-_]?\d{1,2}[.-_]\d{1,2})([.-_]\d{1,3})?([.-_]))?` +.
var reLicenseFile = regexp.MustCompile(
	// generalized pattern matching here
	// COPYRIGHTs
	// UNLICENSE
	// LICENSE where LICENSE
	// may be proceeded by a licence spec (e.g., GPL as in GPLLICENSE or GPL-LICENSE)
	// may have a license spec version   (e.g., GPL_2.0_LICENSE or GPL-2.0.0-LICENSE)
	//     which can be separated by any mix/match of dot, underscore, hyphen
	// may be spelled with 'C' or 'S' (e.g., LICENCE or LICENSE)
	// may be plural (e.g., LICENCES)
	// and the extension is pretty much ignored
	`(?i)` + // case insensitive
		`(` +
		`^patent(s)|` + // patent files
		`^copy(ing|right)|` + // copyright files
		`^(un|` + // unlicence license
		`^(` + // specific license which prefix '[-_]LICEN[sc]E(s)'
		`afl|` +
		`apache|` +
		`apsl|` +
		`artistic|` +
		`(0)?bsd|` +
		`cc([0-])|` +
		`efl|` +
		`epl|` +
		`eupl|` +
		`([al])?gpl|` +
		`mit|` +
		`mpl` +
		`)` +
		`([-_])?` +
		`)?` +
		`licen[sc]e(s)?` +
		`)`,
)

// from checks.md
//   - for the most part extension are ignore, but this cannot be so, many files
//     can get caught up in the filename match (like .license.yaml), therefore
//     the need to accept a host of file extensions which could likely contain
//     license information.
/*
var reLicenseFileExts = regexp.MustCompile(`(?i)(` +
	`\.rst|` +
	`\.txt|` +
	`\.md|` +
	`\.adoc|` +
	`\.xml|` +
	`\.markdown|` +
	`\.html` +
	`)`,
)
*/

// TODO: still working on a more generalized extension.
// TODO: license.yaml still hits and it should not integrate fix.
var reLicenseFileExts = regexp.MustCompile(`(?i)(\w)`)

// Regex converted from
// https://github.com/licensee/licensee/blob/master/lib/licensee/project_files/license_file.rb
// TODO: comprehend if these are needed any longer
/*
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
		{rstr: gpl, f: nil},
	}
)
*/
// License retrieves the raw data for the License check.
func License(c *checker.CheckRequest) (checker.LicenseData, error) {
	var results checker.LicenseData
	var path string

	licensesFound, lerr := c.RepoClient.ListLicenses()
	switch {
	// repo API for licenses is supported
	// go the work and return from immediate (no searching repo).
	case lerr == nil:
		for _, v := range licensesFound {
			results.LicenseFiles = append(results.LicenseFiles,
				checker.LicenseFile{
					File: checker.File{
						Path: v.Path,
						Type: checker.FileTypeSource,
					},
					LicenseInformation: checker.License{
						Key:         v.Key,
						Name:        v.Name,
						SpdxID:      v.SPDXId,
						Attribution: checker.LicenseAttributionTypeRepo,
					},
				})
		}
		return results, nil
	// if repo API for listing licenses is not support
	// continue on using the repo search for a license file.
	case errors.Is(lerr, clients.ErrUnsupportedFeature):
		break
	// something else failed, done.
	default:
		return results, fmt.Errorf("RepoClient.ListLicenses: %w", lerr)
	}

	// no repo API for listing licenses, continue looking for files
	err := fileparser.OnAllFilesDo(c.RepoClient, isLicenseFile, &path)
	if err != nil {
		return results, fmt.Errorf("fileparser.OnAllFilesDo: %w", err)
	}

	// scorecard search stops at first candidate (isLicenseFile) license file found
	if path != "" {
		results.LicenseFiles = append(results.LicenseFiles,
			checker.LicenseFile{
				File: checker.File{
					Path: path,
					Type: checker.FileTypeSource,
				},
				LicenseInformation: checker.License{
					Key:         "",
					Name:        "",
					SpdxID:      "",
					Attribution: checker.LicenseAttributionTypeScorecard,
				},
			})
	}

	return results, nil
}

// TestLicense used for testing purposes.
func TestLicense(name string) bool {
	return checkLicense(name)
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
	proper := false
	for _, indexes := range reLicenseFile.FindAllIndex([]byte(name), -1) {
		// ah.. something matched, is match a proper file name,
		// assume it is and try to refute
		proper = true
		if len(name) != len(name[indexes[0]:indexes[1]]) {
			// the match length does not match the file name
			// so work on ignoring recognized extension(s)
			if (len(name) > indexes[1]) &&
				(!(name[indexes[1]] == '.' || name[indexes[1]] == '/' ||
					name[indexes[1]] == '_' || name[indexes[1]] == '-')) {
				// something other than an extension or path separator
				// occurred after the proper filename (sub)match must be bad
				proper = false
			} else if reLicenseFileExts.FindAllIndex([]byte(name), -1) == nil {
				// TODO: double check this
				// the index found must be at the same name[indexes[1]]
				// anything less is something like .md.license.txt
				proper = false
			}
		}
	}
	return proper
}
