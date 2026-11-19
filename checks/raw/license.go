// Copyright 2021 OpenSSF Scorecard Authors
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
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

// from checks.md
//   - files must be at the top-level directory (hence the '^' in the regex).
//   - files must be of the name like COPY[ING|RIGHT] or LICEN[SC]E(plural)
//     no preceding names or the like (again, hence the '^').
//   - a folder is also acceptable as in COPYRIGHT/ or LICENSES/.
//
// TODO: although the contents of the folder are totally ignored at this time.
// TODO: file or folder the contents are not (yet) examined.
//   - it is a case insenstive check (hence the leading regex compiler option).
var reLicenseFile = regexp.MustCompile(
	// generalized pattern matching to detect a file named:
	// PATENTs
	// COPYRIGHTs
	// LICENSEs
	// where
	//   the detected file:
	//   - must be at the top-level directory
	//   - may be preceded or suffixed by a SPDX Identifier
	//   - may be preceded or suffixed by a pre- or -suf separator
	//   - may have a file extension denoted by a leading dot/period
	// notes
	//   1. when a suffix of '/' for the detected file is found (a folder)
	//      the SPDX Identifier will be sensed from the first filename found
	//      (e.g., LICENSE/GPL-3.0.md will set SPDX Identifier as 'GPL-3.0')
	//      warning: retrieval of filenames is non-deterministic.
	//      TODO: return a list of possible SPDX Identifiers from a folder
	//   2. an SPDX ID is loosely pattern matched by a sequence of alpha
	//      numerics followed by a separator (hyphen, dot, or underscore)
	//      followed by some form of a version number (major, minor, but
	//      no patch--with each number being no more than 2 digits) where
	//      that version number may be proceeded by a letter.
	//   3. the function 'validateSpdxIDAndExt()' in this check works
	//      to disambiguate SPDX Identifiers and allowed file extensions
	//      that cannot be cleanly (or regularly) discerned by regex,
	//      that is why this is a generalized regex pattern.
	`(?i)` + // case insensitive
		`^(?P<lp>` + // must be at the top level (i.e., no leading dot/period or path separator)
		// (opt) preceded SPDX ID (e.g., 'GPL-2.0' as in GPL-2.0-LICENCE)
		`(?P<preSpdx>([0-9A-Za-z]+)((([-_.])[[:digit:]]{1,2}[A-Za-z]{0,1}){0,5})(?P<preSpdxExt>(([_-])?[0-9A-Za-z.])*))?` +
		// (opt) separator before the detected file (e.g., '-' as in GPL-2.0-LICENCE)
		`(?P<pre>([-_]))?` +
		// mandatory file names to detect (e.g., 'LICENCE' as in GPL-2.0-LICENCE)
		`(?P<detectedFile>(patent(s)?|copy(ing|right)|LICEN[SC]E(S)?))` +
		// (opt) separator after the detected file (e.g., '_' as in LICENSE_Apache-1.1)
		`(?P<suf>([-_./]))?` +
		// (opt) suffixed SPDX ID (e.g., 'Apache-1.1' as in LICENSE_Apache-1.1)
		`(?P<sufSpdx>([0-9A-Za-z]+)((([-_.])[[:digit:]]{1,2}[A-Za-z]{0,1}){0,5})(?P<sufSpdxExt>(([_-])?[0-9A-Za-z.])*))?` +
		// (opt) file name extension (e.g., '.md' as in LICENSES.md)
		`(?P<ext>([.]?[A-Za-z]+))?` +
		`)` +
		``,
)

var reGroupNames = reLicenseFile.SubexpNames()

// from checks.md
//   - for the most part extension are ignore, but this cannot be so, many files
//     can get caught up in the filename match (like 'license.yaml'), therefore
//     the need to accept a host of file extensions which could likely contain
//     license information.
var reLicenseFileExts = regexp.MustCompile(
	`(?i)` +
		`(` +
		`\.adoc|` +
		`\.asc|` +
		`\.doc(x)?|` +
		`\.ext|` +
		`\.html|` +
		`\.markdown|` +
		`\.md|` +
		`\.rst|` +
		`\.txt|` +
		`\.xml` +
		`)`,
)

// License retrieves the raw data for the License check.
func License(c *checker.CheckRequest) (checker.LicenseData, error) {
	var results checker.LicenseData

	// prepare case insensitive map to map approved licenses matched in repo.
	setCiMap()

	licensesFound, lerr := c.RepoClient.ListLicenses()
	switch {
	// repo API for licenses is supported
	// go the work and return from immediate (no searching repo).
	case lerr == nil:
		// licenses API may be supported, but platform might not detect license same way we do
		// fallback to our local file logic if no licenses were found or if all returned licenses are unasserted (NOASSERTION)
		hasAssertedLicense := false
		for _, v := range licensesFound {
			if v.SPDXId != "" && !strings.EqualFold(v.SPDXId, "NOASSERTION") {
				hasAssertedLicense = true
				break
			}
		}
		if len(licensesFound) > 0 && hasAssertedLicense {
			for _, v := range licensesFound {
				results.LicenseFiles = append(results.LicenseFiles,
					checker.LicenseFile{
						File: checker.File{
							Path: v.Path,
							Type: finding.FileTypeSource,
						},
						LicenseInformation: checker.License{
							Approved:    len(fsfOsiApprovedLicenseCiMap[strings.ToUpper(v.SPDXId)].Name) > 0,
							Name:        v.Name,
							SpdxID:      v.SPDXId,
							Attribution: checker.LicenseAttributionTypeAPI,
						},
					})
			}
			return results, nil
		}
		// If len(licensesFound) == 0 or only NOASSERTION was returned, fallback to local file logic.
		break
	// if repo API for listing licenses is not support
	// continue on using the repo search for a license file.
	case errors.Is(lerr, clients.ErrUnsupportedFeature):
		break
	// something else failed, done.
	default:
		return results, fmt.Errorf("RepoClient.ListLicenses: %w", lerr)
	}

	// prepare map to index into regex named groups
	// only needs to be done once for the non-repo
	// license check.
	setGroupIdxsMap()

	// no repo API for listing licenses, continue looking for files
	path := checker.LicenseFile{}
	err := fileparser.OnAllFilesDo(c.RepoClient, isLicenseFile, &path)
	if err != nil {
		return results, fmt.Errorf("fileparser.OnAllFilesDo: %w", err)
	}

	// scorecard search stops at first candidate (isLicenseFile) license file found
	if path != (checker.LicenseFile{}) {
		path.LicenseInformation.Name = fsfOsiApprovedLicenseCiMap[strings.ToUpper(path.LicenseInformation.SpdxID)].Name
		// these settings (Name, Key) match GH repo API
		// for when the Spdx Identifier cannot be determined.
		if path.LicenseInformation.SpdxID == "" {
			path.LicenseInformation.SpdxID = "NOASSERTION"
			path.LicenseInformation.Name = "Other"
		}
		path.LicenseInformation.Approved = len(
			fsfOsiApprovedLicenseCiMap[strings.ToUpper(path.LicenseInformation.SpdxID)].Name) > 0
		path.LicenseInformation.Attribution = checker.LicenseAttributionTypeHeuristics
		results.LicenseFiles = append(results.LicenseFiles, path)
	} else if len(licensesFound) > 0 {
		// If local file detection did not locate a license file, preserve the API's NOASSERTION
		// finding so repositories with an unclassified license still receive credit for having one.
		for _, v := range licensesFound {
			results.LicenseFiles = append(results.LicenseFiles,
				checker.LicenseFile{
					File: checker.File{
						Path: v.Path,
						Type: finding.FileTypeSource,
					},
					LicenseInformation: checker.License{
						Approved:    false,
						Name:        v.Name,
						SpdxID:      v.SPDXId,
						Attribution: checker.LicenseAttributionTypeAPI,
					},
				})
		}
	}

	return results, nil
}

// TestLicense used for testing purposes.
func TestLicense(name string) bool {
	setCiMap()
	setGroupIdxsMap()

	_, ok := checkLicense(name)
	return ok
}

func isApprovedLicense(spdxID string) bool {
	if spdxID == "" {
		return false
	}
	setCiMap()
	return len(fsfOsiApprovedLicenseCiMap[strings.ToUpper(spdxID)].Name) > 0
}

var isLicenseFile fileparser.DoWhileTrueOnFilename = func(name string, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("isLicenseFile requires exactly one argument: %w", errInvalidArgLength)
	}
	s, ok := args[0].(*checker.LicenseFile)
	if !ok {
		return false, fmt.Errorf("isLicenseFile requires argument of type: *checker.LicenseFile: %w", errInvalidArgType)
	}
	matched, ok := checkLicense(name)
	if !ok {
		return true, nil
	}

	// If this license has an approved SPDX ID, choose it and stop searching.
	if isApprovedLicense(matched.LicenseInformation.SpdxID) {
		*s = matched
		return false, nil
	}

	// If we haven't found any candidate yet, or if previous candidate had no SpdxID
	// but the current candidate does, update candidate.
	if s.File.Path == "" || (s.LicenseInformation.SpdxID == "" && matched.LicenseInformation.SpdxID != "") {
		*s = matched
	}

	return true, nil
}

// direct mapping of regex groups by name to index number.
var reGroupIdxs = make(map[string]int)

// case-insensitive map of fsfOsiApprovedLicenseMap.
var fsfOsiApprovedLicenseCiMap = map[string]fsfOsiLicenseType{}

// parallel testing in license_test.go causes a race
// in initializing the CiMap and reGroupIdxs with values, this shared mutex
// prevents that race condition in the unit-test.
var (
	ciMapMutex        = sync.Mutex{}
	reGrpIdxsMapMutex = sync.Mutex{}
)

func setCiMap() {
	ciMapMutex.Lock()
	defer ciMapMutex.Unlock()
	if len(fsfOsiApprovedLicenseCiMap) == 0 {
		for key, entry := range fsfOsiApprovedLicenseMap {
			fsfOsiApprovedLicenseCiMap[strings.ToUpper(key)] = entry
		}
	}
}

func setGroupIdxsMap() {
	reGrpIdxsMapMutex.Lock()
	defer reGrpIdxsMapMutex.Unlock()
	if len(reGroupIdxs) == 0 {
		for idx, vl := range reGroupNames {
			if vl != "" {
				reGroupIdxs[vl] = idx
			}
		}
	}
}

func getSpdxID(matches []string) string {
	// try to discern an SPDX Identifier (a)
	// should both "preSpdx" and "sufSpdx" have a
	// value, preSpdx takes precedence.
	// (e.g., 0BSD-LICENSE-GPL-2.0.txt)
	// TODO: decide if that is OK or should "fail"
	var id string
	if matches[reGroupIdxs["preSpdx"]] != "" {
		id = matches[reGroupIdxs["preSpdx"]]
	} else if matches[reGroupIdxs["sufSpdx"]] != "" {
		id = matches[reGroupIdxs["sufSpdx"]]
	}
	// Special case, the unlicense, in the map is
	// called 'The Unlicense' with the Spdx id 'Unlicense'.
	// For the regex's 'un' will match the [pre|suf]Spdx
	// regex group (just as it would match '0BSD'), but
	// 'un' will not "hit" in the map with key 'Unlicense'
	if strings.EqualFold(id, "UN") {
		id = "UNLICENSE"
	}

	return id
}

func getExt(filename string, matches []string) string {
	ext := filepath.Ext(filename)
	if ext != "" && strings.Contains(ext, matches[reGroupIdxs["detectedFile"]]) {
		// fixes when ext incorporates part of detectedFile as ext
		return ""
	}
	return ext
}

func getFolder(matches []string) string {
	if matches[reGroupIdxs["suf"]] == "/" {
		return matches[reGroupIdxs["detectedFile"]] + matches[reGroupIdxs["suf"]]
	}
	return ""
}

func extensionOK(ext string) bool {
	if ext == "" {
		return true
	}
	return len(reLicenseFileExts.FindStringSubmatch(ext)) != 0
}

func validateSpdxIDAndExt(matches []string, spdx, ext string) (string, string) {
	if spdx == "" {
		return spdx, ext
	}

	// fixes when [pre|suf]Spdx consumes ext
	if ext != "" && strings.Contains(ext, spdx) {
		spdx = ""
	}
	// fixes when ext incorporates part of [pre|suf]Spdx as ext
	// while allowing spdx's that appear as extensions which
	// start with a digit.
	if ext != "" && spdx != "" && strings.Contains(spdx, ext) {
		spdxDigitInExt := regexp.MustCompile(`\.[[:digit:]]`)
		if !extensionOK(ext) && len(spdxDigitInExt.FindStringSubmatch(spdx)) > 0 {
			ext = ""
		} else {
			spdx = strings.ReplaceAll(spdx, ext, "")
		}
	} else if ext != "" && spdx != "" && ext != spdx {
		if ext != matches[reGroupIdxs["ext"]] {
			spdx += matches[reGroupIdxs["ext"]]
		}
	}
	return spdx, ext
}

func getLicensePath(matches []string, val, spdx, ext string) string {
	lp := matches[reGroupIdxs["lp"]]
	// at this point what matched should equal
	// the input given that there must have
	// been an extension or either/both some
	// match to an Spdx ID, check that here.
	if lp != val {
		if getFolder(matches) == "" && spdx == "" {
			return ""
		} else if lp+ext == val {
			return lp + ext
		}
		return ""
	}
	// last chance to reject file extensions which are not allowed
	if !extensionOK(ext) {
		return ""
	}
	return lp
}

func checkLicense(lfName string) (checker.LicenseFile, bool) {
	grpMatches := reLicenseFile.FindStringSubmatch(lfName)
	if len(grpMatches) == 0 {
		return checker.LicenseFile{}, false
	}

	// Detected one of the mandatory file names
	// quick check for done (the name passed is
	// detected in its entirety)
	// TODO: open/read contents to try to discern
	//       license as the name of the file has
	//       no hints.
	licensePath := grpMatches[reGroupIdxs["detectedFile"]]
	if lfName == licensePath {
		return (checker.LicenseFile{
			File: checker.File{
				Path: licensePath,
				Type: finding.FileTypeSource,
			},
			LicenseInformation: checker.License{
				Name:   "",
				SpdxID: "",
			},
		}), true
	}
	// there is more in the file name,
	// match might yield additional hints.
	//   a. have an (or more) Spdx Identifier, and/or
	licenseSpdxID := getSpdxID(grpMatches)
	//   b. have an extension
	licenseExt := getExt(lfName, grpMatches)
	//   c. or, the detected file could be a folder
	// TODO: licenseFolder := getFolder(grpMatches)
	// deconflict any overlap in group matches for SpdxID and any extension
	licenseSpdxID, licenseExt = validateSpdxIDAndExt(grpMatches, licenseSpdxID, licenseExt)
	// reset licensePath based on validated matches
	licensePath = getLicensePath(grpMatches, lfName, licenseSpdxID, licenseExt)
	if licensePath == "" {
		return checker.LicenseFile{}, false
	}

	return (checker.LicenseFile{
		File: checker.File{
			Path: licensePath,
			Type: finding.FileTypeSource, // TODO: introduce FileTypeFolder with licenseFolder
		},
		LicenseInformation: checker.License{
			SpdxID: licenseSpdxID,
		},
	}), true
}

type fsfOsiLicenseType struct {
	Name string
}

// The SPDX license list map fsfOsiApprovedLicenseMap is generated in license_map.go
// from a pinned SPDX release via checks/raw/internal/generate/main.go.
