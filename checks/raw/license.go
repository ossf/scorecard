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
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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
//     can get caught up in the filename match (like .license.yaml), therefore
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
	ciMapMutex.Lock()
	defer ciMapMutex.Unlock()
	if len(fsfOsiApprovedLicenseCiMap) == 0 {
		for key, entry := range fsfOsiApprovedLicenseMap {
			// Special case, the unlicense, in the map is
			// called 'The Unlicense' with the Spdx id 'Unlicense'.
			// For the regex's 'un' will match the [pre|suf]Spdx
			// regex group (just as it would match '0BSD'), but
			// 'un' will not "hit" in the map with key 'Unlicense'
			// so change to 'UN' for 'unlicense' for 'isLicenseFile()'
			if strings.ToUpper(key) == "UNLICENSE" {
				fsfOsiApprovedLicenseCiMap["UN"] = entry
			} else {
				// TODO: unit-test fails with race condition
				//       source seems to be strings.ToUpper()
				//       unit-test does a parallel test - could be the cause
				fsfOsiApprovedLicenseCiMap[strings.ToUpper(key)] = entry
			}
		}
	}

	licensesFound, lerr := c.RepoClient.ListLicenses()
	// TODO: lerr = clients.ErrUnsupportedFeature
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
						Approved:    len(fsfOsiApprovedLicenseCiMap[strings.ToUpper(v.SPDXId)].Name) > 0,
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

	// prepare map to index into regex named groups
	// only needs to be done once for the license check
	if len(reGroupIdxs) == 0 {
		for idx, vl := range reGroupNames {
			if vl != "" {
				reGroupIdxs[vl] = idx
			}
		}
	}

	// no repo API for listing licenses, continue looking for files
	path := checker.LicenseFile{}
	err := fileparser.OnAllFilesDo(c.RepoClient, isLicenseFile, &path)
	if err != nil {
		return results, fmt.Errorf("fileparser.OnAllFilesDo: %w", err)
	}

	// scorecard search stops at first candidate (isLicenseFile) license file found
	if path != (checker.LicenseFile{}) {
		//
		// now it is time to "map it back" in the case of the
		// Spdx Identifier for "UNLICENSE" which was mapped to "UN"
		// for the regex group match and this check.
		// grab what is needed before clobbering the Spdx Identifier
		// Aside from 'UN', these settings (Name, Key) match GH repo API
		// for when the Spdx Identifier cannot be determined.
		path.LicenseInformation.Name = fsfOsiApprovedLicenseCiMap[strings.ToUpper(path.LicenseInformation.SpdxID)].Name
		path.LicenseInformation.Key = strings.ToLower(path.LicenseInformation.SpdxID)
		if strings.ToUpper(path.LicenseInformation.SpdxID) == "UN" {
			path.LicenseInformation.SpdxID = "UNLICENSE"
			path.LicenseInformation.Key = strings.ToLower(path.LicenseInformation.SpdxID)
		} else if path.LicenseInformation.SpdxID == "" {
			path.LicenseInformation.SpdxID = "NOASSERTION"
			path.LicenseInformation.Name = "Other"
			path.LicenseInformation.Key = strings.ToLower(path.LicenseInformation.Name)
		}
		path.LicenseInformation.Approved = len(
			fsfOsiApprovedLicenseCiMap[strings.ToUpper(path.LicenseInformation.SpdxID)].Name) > 0
		path.LicenseInformation.Attribution = checker.LicenseAttributionTypeScorecard
		results.LicenseFiles = append(results.LicenseFiles, path)
	}

	return results, nil
}

// TestLicense used for testing purposes.
func TestLicense(name string) bool {
	_, ok := checkLicense(name)
	return ok
}

var isLicenseFile fileparser.DoWhileTrueOnFilename = func(name string, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("isLicenseFile requires exactly one argument: %w", errInvalidArgLength)
	}
	s, ok := args[0].(*checker.LicenseFile)
	if !ok {
		return false, fmt.Errorf("isLicenseFile requires argument of type: *checker.LicenseFile: %w", errInvalidArgType)
	}
	*s, ok = checkLicense(name)
	if ok {
		return false, nil
	}
	return true, nil
}

var reGroupIdxs = make(map[string]int)

func getSpdxID(matches []string) string {
	// try to discern an SPDX Identifier (a)
	// should both "preSpdx" and "sufSpdx" have a
	// value, preSpdx takes precedence.
	// (e.g., 0BSD-LICENSE-GPL-2.0.txt)
	// TODO: decide if that is OK or should "fail"
	if matches[reGroupIdxs["preSpdx"]] != "" {
		return matches[reGroupIdxs["preSpdx"]]
	} else if matches[reGroupIdxs["sufSpdx"]] != "" {
		return matches[reGroupIdxs["sufSpdx"]]
	}
	return ""
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
	// TODO check and if needed reject files with unrecognized extensions
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
		// TODO: fmt.Printf("FIX FAILURE C Ext is '%s' spdx is '%s' reExt is '%s'\n", ext, spdx, matches[reGroupIdxs["ext"]])
		spdx = ""
	}
	// fixes when ext incorporates part of [pre|suf]Spdx as ext
	if ext != "" && spdx != "" && strings.Contains(spdx, ext) {
		if !extensionOK(ext) {
			// TODO: fmt.Printf("FIX FAILURE B\n")
			ext = ""
		} else {
			spdx = strings.ReplaceAll(spdx, ext, "")
			// TODO: fmt.Printf("FIXED FAILURE Bb r='%s, %s'\n", spdx, ext)
		}
	} else if ext != "" && spdx != "" && ext != spdx {
		if ext != matches[reGroupIdxs["ext"]] {
			spdx = spdx + matches[reGroupIdxs["ext"]]
			// TODO: fmt.Printf("FIXED FAILURE Ba '%s'('%s') != '%s'\n", ext, matches[reGroupIdxs["ext"]], spdx)
		}
	}
	return spdx, ext
}

func getLicensePath(matches []string, val, spdx, ext string) string {
	lp := matches[reGroupIdxs["lp"]]
	// at this point what matched should equal
	// the input given that there must have
	// been an extension or either/both some
	// match to an Spdx ID... check that here
	if lp != val {
		if getFolder(matches) == "" && spdx == "" {
			// TODO: fmt.Printf("NG: in='%s' != lp='%s'\n", val, lp)
			return ""
		} else if lp+ext == val {
			// TODO: fmt.Printf("\nFIX FAILURE E in='%s' != lp='%s'\nFIX FAILURE E '%#v'\n", val, lp, matches)
			return lp + ext
		}
		// fixes failure E when an SpdxID is partially matched in regex
		// this is for names like 'LICENSE-CC-BY-SA-4.0.md'
		// which is caught here--the SpdxID 100% matched.
		// TODO: fmt.Printf("\nASSERTION FAILURE E in='%s' != lp='%s'\n", val, lp)
		return ""
	}
	if !extensionOK(ext) {
		// TODO: fmt.Printf("\nASSERTION FAILURE E in='%s' != lp='%s' ext='%s'\n", val, lp, ext)
		return ""
	}
	return lp
}

func checkLicense(lfName string) (checker.LicenseFile, bool) {
	grpMatches := reLicenseFile.FindStringSubmatch(lfName)
	if len(grpMatches) == 0 {
		// TODO: fmt.Printf ("NG: 3 '%s'\n", lfName)
		return checker.LicenseFile{}, false
	}

	// ah.. detected one of the mandory file names
	// quick check for done (the name passed is
	// detected in its entirety)
	// TODO: open/read contents to try to discern
	//       license as the name of the file has
	//       no hints.
	licensePath := grpMatches[reGroupIdxs["detectedFile"]]
	if lfName == licensePath {
		// TODO: fmt.Printf ("CK: 1 '%s'\n", licensePath)
		return (checker.LicenseFile{
			File: checker.File{
				Path: licensePath,
				Type: checker.FileTypeSource,
			},
			LicenseInformation: checker.License{
				Key:    "",
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
			Type: checker.FileTypeSource, // TODO: introduce FileTypeFolder
		},
		LicenseInformation: checker.License{
			SpdxID: licenseSpdxID,
		},
	}), true
}

type fsfOsiLicenseType struct {
	Name string
}

// case-insensitive map of fsfOsiApprovedLicenseMap.
var fsfOsiApprovedLicenseCiMap = map[string]fsfOsiLicenseType{}

// parallel testing in license_test.go causes a race
// in initializing the CiMap above, this shared mutex
// prevents that race condition in the unit-test.
var ciMapMutex = sync.Mutex{}

// (there's no magic here)
// created from the table at https://spdx.org/licenses
// 1) convert table to CSV and convert to json
// (inspired from inspired from https://stackoverflow.com/questions/17912307/u-ufeff-in-python-string)
//
//	cat SPDX-license-IDs-202211131030CST.csv | \
//	 python -c \
//	   'import csv, json, sys; print(json.dumps([dict(r) for r in csv.DictReader(sys.stdin)]))' | \
//	 sed 's/\\ufeff//g' > SPDX-license-IDs-202211131030CST.json
//
// 2) convert json to go lang compatible map
//
//	cat SPDX-license-IDs-202211131030CST.json | \
//	 jq -c '.[] | select((."FSF Free/Libre?"=="Y") or (."OSI Approved?"=="Y")) | \
//	   [."Identifier",."Full Name"]' | \
//	 sed 's/","/": \{ Name: "/;s/\["/"/;s/"\]/" },/' | \
//	 sed 's/[- ]only//g;s/[- ]or[- ]later//g' | \
//	 sort | \
//	 uniq
var fsfOsiApprovedLicenseMap = map[string]fsfOsiLicenseType{
	"0BSD":                            {Name: "BSD Zero Clause License"},
	"AAL":                             {Name: "Attribution Assurance License"},
	"AFL-1.1":                         {Name: "Academic Free License v1.1"},
	"AFL-1.2":                         {Name: "Academic Free License v1.2"},
	"AFL-2.0":                         {Name: "Academic Free License v2.0"},
	"AFL-2.1":                         {Name: "Academic Free License v2.1"},
	"AFL-3.0":                         {Name: "Academic Free License v3.0"},
	"AGPL-3.0":                        {Name: "GNU Affero General Public License v3.0"},
	"APL-1.0":                         {Name: "Adaptive Public License 1.0"},
	"APSL-1.0":                        {Name: "Apple Public Source License 1.0"},
	"APSL-1.1":                        {Name: "Apple Public Source License 1.1"},
	"APSL-1.2":                        {Name: "Apple Public Source License 1.2"},
	"APSL-2.0":                        {Name: "Apple Public Source License 2.0"},
	"Apache-1.0":                      {Name: "Apache License 1.0"},
	"Apache-1.1":                      {Name: "Apache License 1.1"},
	"Apache-2.0":                      {Name: "Apache License 2.0"},
	"Artistic-1.0":                    {Name: "Artistic License 1.0"},
	"Artistic-1.0-Perl":               {Name: "Artistic License 1.0 (Perl)"},
	"Artistic-1.0-cl8":                {Name: "Artistic License 1.0 w/clause 8"},
	"Artistic-2.0":                    {Name: "Artistic License 2.0"},
	"BSD-1-Clause":                    {Name: "BSD 1-Clause License"},
	"BSD-2-Clause":                    {Name: "BSD 2-Clause \"Simplified\" License"},
	"BSD-2-Clause-Patent":             {Name: "BSD-2-Clause Plus Patent License"},
	"BSD-3-Clause":                    {Name: "BSD 3-Clause \"New\" or \"Revised\" License"},
	"BSD-3-Clause-Clear":              {Name: "BSD 3-Clause Clear License"},
	"BSD-3-Clause-LBNL":               {Name: "Lawrence Berkeley National Labs BSD variant license"},
	"BSD-4-Clause":                    {Name: "BSD 4-Clause \"Original\" or \"Old\" License"},
	"BSL-1.0":                         {Name: "Boost Software License 1.0"},
	"BitTorrent-1.1":                  {Name: "BitTorrent Open Source License v1.1"},
	"CAL-1.0":                         {Name: "Cryptographic Autonomy License 1.0"},
	"CAL-1.0-Combined-Work-Exception": {Name: "Cryptographic Autonomy License 1.0 (Combined Work Exception)"},
	"CATOSL-1.1":                      {Name: "Computer Associates Trusted Open Source License 1.1"},
	"CC-BY-4.0":                       {Name: "Creative Commons Attribution 4.0 International"},
	"CC-BY-SA-4.0":                    {Name: "Creative Commons Attribution Share Alike 4.0 International"},
	"CC0-1.0":                         {Name: "Creative Commons Zero v1.0 Universal"},
	"CDDL-1.0":                        {Name: "Common Development and Distribution License 1.0"},
	"CECILL-2.0":                      {Name: "CeCILL Free Software License Agreement v2.0"},
	"CECILL-2.1":                      {Name: "CeCILL Free Software License Agreement v2.1"},
	"CECILL-B":                        {Name: "CeCILL-B Free Software License Agreement"},
	"CECILL-C":                        {Name: "CeCILL-C Free Software License Agreement"},
	"CERN-OHL-P-2.0":                  {Name: "CERN Open Hardware Licence Version 2 - Permissive"},
	"CERN-OHL-S-2.0":                  {Name: "CERN Open Hardware Licence Version 2 - Strongly Reciprocal"},
	"CERN-OHL-W-2.0":                  {Name: "CERN Open Hardware Licence Version 2 - Weakly Reciprocal"},
	"CNRI-Python":                     {Name: "CNRI Python License"},
	"CPAL-1.0":                        {Name: "Common Public Attribution License 1.0"},
	"CPL-1.0":                         {Name: "Common Public License 1.0"},
	"CUA-OPL-1.0":                     {Name: "CUA Office Public License v1.0"},
	"ClArtistic":                      {Name: "Clarified Artistic License"},
	"Condor-1.1":                      {Name: "Condor Public License v1.1"},
	"ECL-1.0":                         {Name: "Educational Community License v1.0"},
	"ECL-2.0":                         {Name: "Educational Community License v2.0"},
	"EFL-1.0":                         {Name: "Eiffel Forum License v1.0"},
	"EFL-2.0":                         {Name: "Eiffel Forum License v2.0"},
	"EPL-1.0":                         {Name: "Eclipse Public License 1.0"},
	"EPL-2.0":                         {Name: "Eclipse Public License 2.0"},
	"EUDatagrid":                      {Name: "EU DataGrid Software License"},
	"EUPL-1.1":                        {Name: "European Union Public License 1.1"},
	"EUPL-1.2":                        {Name: "European Union Public License 1.2"},
	"Entessa":                         {Name: "Entessa Public License v1.0"},
	"FSFAP":                           {Name: "FSF All Permissive License"},
	"FTL":                             {Name: "Freetype Project License"},
	"Fair":                            {Name: "Fair License"},
	"Frameworx-1.0":                   {Name: "Frameworx Open License 1.0"},
	"GFDL-1.1":                        {Name: "GNU Free Documentation License v1.1"},
	"GFDL-1.2":                        {Name: "GNU Free Documentation License v1.2"},
	"GFDL-1.3":                        {Name: "GNU Free Documentation License v1.3"},
	"GPL-2.0":                         {Name: "GNU General Public License v2.0"},
	"GPL-3.0":                         {Name: "GNU General Public License v3.0"},
	"HPND":                            {Name: "Historical Permission Notice and Disclaimer"},
	"IJG":                             {Name: "Independent JPEG Group License"},
	"IPA":                             {Name: "IPA Font License"},
	"IPL-1.0":                         {Name: "IBM Public License v1.0"},
	"ISC":                             {Name: "ISC License"},
	"Imlib2":                          {Name: "Imlib2 License"},
	"Intel":                           {Name: "Intel Open Source License"},
	"Jam":                             {Name: "Jam License"},
	"LGPL-2.0":                        {Name: "GNU Library General Public License v2"},
	"LGPL-2.1":                        {Name: "GNU Lesser General Public License v2.1"},
	"LGPL-3.0":                        {Name: "GNU Lesser General Public License v3.0"},
	"LPL-1.0":                         {Name: "Lucent Public License Version 1.0"},
	"LPL-1.02":                        {Name: "Lucent Public License v1.02"},
	"LPPL-1.2":                        {Name: "LaTeX Project Public License v1.2"},
	"LPPL-1.3a":                       {Name: "LaTeX Project Public License v1.3a"},
	"LPPL-1.3c":                       {Name: "LaTeX Project Public License v1.3c"},
	"LiLiQ-P-1.1":                     {Name: "Licence Libre du Québec – Permissive version 1.1"},
	"LiLiQ-R-1.1":                     {Name: "Licence Libre du Québec – Réciprocité version 1.1"},
	"LiLiQ-Rplus-1.1":                 {Name: "Licence Libre du Québec – Réciprocité forte version 1.1"},
	"MIT":                             {Name: "MIT License"},
	"MIT-0":                           {Name: "MIT No Attribution"},
	"MIT-Modern-Variant":              {Name: "MIT License Modern Variant"},
	"MPL-1.0":                         {Name: "Mozilla Public License 1.0"},
	"MPL-1.1":                         {Name: "Mozilla Public License 1.1"},
	"MPL-2.0":                         {Name: "Mozilla Public License 2.0"},
	"MPL-2.0-no-copyleft-exception":   {Name: "Mozilla Public License 2.0 (no copyleft exception)"},
	"MS-PL":                           {Name: "Microsoft Public License"},
	"MS-RL":                           {Name: "Microsoft Reciprocal License"},
	"MirOS":                           {Name: "The MirOS Licence"},
	"Motosoto":                        {Name: "Motosoto License"},
	"MulanPSL-2.0":                    {Name: "Mulan Permissive Software License, Version 2"},
	"Multics":                         {Name: "Multics License"},
	"NASA-1.3":                        {Name: "NASA Open Source Agreement 1.3"},
	"NCSA":                            {Name: "University of Illinois/NCSA Open Source License"},
	"NGPL":                            {Name: "Nethack General Public License"},
	"NOSL":                            {Name: "Netizen Open Source License"},
	"NPL-1.0":                         {Name: "Netscape Public License v1.0"},
	"NPL-1.1":                         {Name: "Netscape Public License v1.1"},
	"NPOSL-3.0":                       {Name: "Non-Profit Open Software License 3.0"},
	"NTP":                             {Name: "NTP License"},
	"Naumen":                          {Name: "Naumen Public License"},
	"Nokia":                           {Name: "Nokia Open Source License"},
	"OCLC-2.0":                        {Name: "OCLC Research Public License 2.0"},
	"ODbL-1.0":                        {Name: "Open Data Commons Open Database License v1.0"},
	"OFL-1.0":                         {Name: "SIL Open Font License 1.0"},
	"OFL-1.1":                         {Name: "SIL Open Font License 1.1"},
	"OFL-1.1-RFN":                     {Name: "SIL Open Font License 1.1 with Reserved Font Name"},
	"OFL-1.1-no-RFN":                  {Name: "SIL Open Font License 1.1 with no Reserved Font Name"},
	"OGTSL":                           {Name: "Open Group Test Suite License"},
	"OLDAP-2.3":                       {Name: "Open LDAP Public License v2.3"},
	"OLDAP-2.7":                       {Name: "Open LDAP Public License v2.7"},
	"OLDAP-2.8":                       {Name: "Open LDAP Public License v2.8"},
	"OSET-PL-2.1":                     {Name: "OSET Public License version 2.1"},
	"OSL-1.0":                         {Name: "Open Software License 1.0"},
	"OSL-1.1":                         {Name: "Open Software License 1.1"},
	"OSL-2.0":                         {Name: "Open Software License 2.0"},
	"OSL-2.1":                         {Name: "Open Software License 2.1"},
	"OSL-3.0":                         {Name: "Open Software License 3.0"},
	"OpenSSL":                         {Name: "OpenSSL License"},
	"PHP-3.0":                         {Name: "PHP License v3.0"},
	"PHP-3.01":                        {Name: "PHP License v3.01"},
	"PostgreSQL":                      {Name: "PostgreSQL License"},
	"Python-2.0":                      {Name: "Python License 2.0"},
	"QPL-1.0":                         {Name: "Q Public License 1.0"},
	"RPL-1.1":                         {Name: "Reciprocal Public License 1.1"},
	"RPL-1.5":                         {Name: "Reciprocal Public License 1.5"},
	"RPSL-1.0":                        {Name: "RealNetworks Public Source License v1.0"},
	"RSCPL":                           {Name: "Ricoh Source Code Public License"},
	"Ruby":                            {Name: "Ruby License"},
	"SGI-B-2.0":                       {Name: "SGI Free Software License B v2.0"},
	"SISSL":                           {Name: "Sun Industry Standards Source License v1.1"},
	"SMLNJ":                           {Name: "Standard ML of New Jersey License"},
	"SPL-1.0":                         {Name: "Sun Public License v1.0"},
	"SimPL-2.0":                       {Name: "Simple Public License 2.0"},
	"Sleepycat":                       {Name: "Sleepycat License"},
	"UCL-1.0":                         {Name: "Upstream Compatibility License v1.0"},
	"UPL-1.0":                         {Name: "Universal Permissive License v1.0"},
	"Unicode-DFS-2016":                {Name: "Unicode License Agreement - Data Files and Software (2016)"},
	"Unlicense":                       {Name: "The Unlicense"},
	"VSL-1.0":                         {Name: "Vovida Software License v1.0"},
	"Vim":                             {Name: "Vim License"},
	"W3C":                             {Name: "W3C Software Notice and License (2002-12-31)"},
	"WTFPL":                           {Name: "Do What The F*ck You Want To Public License"},
	"Watcom-1.0":                      {Name: "Sybase Open Watcom Public License 1.0"},
	"X11":                             {Name: "X11 License"},
	"XFree86-1.1":                     {Name: "XFree86 License 1.1"},
	"Xnet":                            {Name: "X.Net License"},
	"YPL-1.1":                         {Name: "Yahoo! Public License v1.1"},
	"ZPL-2.0":                         {Name: "Zope Public License 2.0"},
	"ZPL-2.1":                         {Name: "Zope Public License 2.1"},
	"Zend-2.0":                        {Name: "Zend License v2.0"},
	"Zimbra-1.3":                      {Name: "Zimbra Public License v1.3"},
	"Zlib":                            {Name: "zlib License"},
	"gnuplot":                         {Name: "gnuplot License"},
	"iMatix":                          {Name: "iMatix Standard Function Library Agreement"},
	"xinetd":                          {Name: "xinetd License"},
}
