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

package evaluation

import (
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

func scoreLicenseCriteria(f *checker.LicenseFile,
	dl checker.DetailLogger,
) int {
	var score int
	msg := checker.LogMessage{
		Path:   "",
		Type:   checker.FileTypeNone,
		Text:   "",
		Offset: 1,
	}
	msg.Path = f.File.Path
	msg.Type = checker.FileTypeSource
	fsfOsiLicense := len(fsfOsiApprovedLicenseMap[f.LicenseInformation.SpdxID].Name) > 0
	// #1 a license file was found.
	score += 6

	// #2 the licence was found at the top-level or LICENSE/ folder.
	switch f.LicenseInformation.Attribution {
	case checker.LicenseAttributionTypeRepo, checker.LicenseAttributionTypeScorecard:
		// both repoAPI and scorecard (not using the API) follow checks.md
		// for a file to be found it must have been in the correct location
		// award location points.
		score += 3
		msg.Text = "License file found in expected location"
		dl.Info(&msg)
		// for repo attribution prepare warning if not an recognized license"
		msg.Text = "Any licence detected not an FSF or OSI recognized license"
	case checker.LicenseAttributionTypeOther:
		// TODO ascertain location found
		score += 0
		msg.Text = "License file found in unexpected location"
		dl.Warn(&msg)
		// for non repo attribution not the license detection is not supported
		msg.Text = "Detecting license content not supported"
	default:
	}

	// #3 is the license either an FSF or OSI recognized/approved license
	if fsfOsiLicense {
		score += 1
		msg.Text = "FSF or OSI recognized license"
		dl.Info(&msg)
	} else {
		// message text for this condition set above
		dl.Warn(&msg)
	}
	return score
}

// License applies the score policy for the License check.
func License(name string, dl checker.DetailLogger,
	r *checker.LicenseData,
) checker.CheckResult {
	var score int
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.LicenseFiles == nil || len(r.LicenseFiles) == 0 {
		return checker.CreateMinScoreResult(name, "license file not detected")
	}

	// TODO: although this a loop, the raw checks will only return one licence file
	// when more than one license file can be aggregated into a composite
	// score, that logic can be comprehended here.
	score = 0
	for idx := range r.LicenseFiles {
		score = scoreLicenseCriteria(&r.LicenseFiles[idx], dl)
	}

	return checker.CreateResultWithScore(name, "license file detected", score)
}

type fsfosiLicenseType struct {
	Name string
}

var fsfOsiApprovedLicenseMap = map[string]fsfosiLicenseType{
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
