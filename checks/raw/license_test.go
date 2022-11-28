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

package raw

import (
	"testing"
)

func TestLicenseFileCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filename   string
		extensions []string
		shouldFail bool
	}{
		{
			name:       "LICENSE",
			filename:   "LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
				".adoc",
				".asc",
				".docx",
				".doc",
				".ext",
				".html",
				".markdown",
				".md",
				".rst",
				".txt",
				".xml",
			},
		},
		{
			name:       "LICENCE",
			filename:   "LICENCE",
			shouldFail: false,
			extensions: []string{
				"",
				".adoc",
				".asc",
				".docx",
				".doc",
				".ext",
				".html",
				".markdown",
				".md",
				".rst",
				".txt",
				".xml",
			},
		},
		{
			name:       "COPYING",
			filename:   "COPYING",
			shouldFail: false,
			extensions: []string{
				"",
				".adoc",
				".asc",
				".docx",
				".doc",
				".ext",
				".html",
				".markdown",
				".md",
				".rst",
				".txt",
				".xml",
				"-MIT",
			},
		},
		{
			name:       "MIT-LICENSE-MIT",
			filename:   "MIT-LICENSE-MIT",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "MIT-COPYING",
			filename:   "MIT-COPYING",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "OFL",
			filename:   "OFL",
			shouldFail: true,
			extensions: []string{
				"",
				".md",
				".textile",
			},
		},
		{
			name:       "LPPL-1.3clicences",
			filename:   "LPPL-1.3clicences",
			shouldFail: false,
			extensions: []string{
				"",
				".xml",
			},
		},
		{
			name:       "LICENSE_LPPL-1.3c",
			filename:   "LICENSE_LPPL-1.3c",
			shouldFail: false,
			extensions: []string{
				"",
				".md",
			},
		},
		{
			name:       "Artistic-1.0-Perl LICENSE",
			filename:   "Artistic-1.0-Perl LICENSE",
			shouldFail: true,
			extensions: []string{
				"",
				".md",
				".textile",
			},
		},
		{
			name:       "PATENTS",
			filename:   "PATENTS",
			shouldFail: false,
			extensions: []string{
				"",
				".txt",
			},
		},
		{
			name:       "GPL",
			filename:   "GPL",
			shouldFail: true,
			extensions: []string{
				"v1",
				"-1.0",
				"v2",
				"-2.0",
				"v3",
				"-3.0",
			},
		},
		{
			name:       ".actions/ASFLicenseHeaderMarkdown.txt",
			filename:   ".actions/ASFLicenseHeaderMarkdown.txt",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".baseline/copyright/001_apache-2.0.txt",
			filename:   ".baseline/copyright/001_apache-2.0.txt",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".copyrightignore",
			filename:   ".copyrightignore",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".dependency_license",
			filename:   ".dependency_license",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/testForLicenseHeaders.sh",
			filename:   ".github/testForLicenseHeaders.sh",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/ci_check_license.yml",
			filename:   ".github/workflows/ci_check_license.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/license-check.yaml",
			filename:   ".github/workflows/license-check.yaml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/license-check.yml",
			filename:   ".github/workflows/license-check.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/license-checker.yml",
			filename:   ".github/workflows/license-checker.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/license-eyes.yml",
			filename:   ".github/workflows/license-eyes.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/license.yml",
			filename:   ".github/workflows/license.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/licensecheck.yml",
			filename:   ".github/workflows/licensecheck.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".github/workflows/licenses.yml",
			filename:   ".github/workflows/licenses.yml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".idea/copyright/Eclipse.xml",
			filename:   ".idea/copyright/Eclipse.xml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".idea/copyright/Winery.xml",
			filename:   ".idea/copyright/Winery.xml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".licenserc.yaml",
			filename:   ".licenserc.yaml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       ".yarn/cache/spdx-license-ids-npm-3.0.5-cb028e9441-b1ceea3f87.zip",
			filename:   ".yarn/cache/spdx-license-ids-npm-3.0.5-cb028e9441-b1ceea3f87.zip",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "APACHE_LICENSETEXT.md",
			filename:   "APACHE_LICENSETEXT.md",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "COPYRIGHT",
			filename:   "COPYRIGHT",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "COPYRIGHT.txt",
			filename:   "COPYRIGHT.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "Docgen/business/features/org.polarsys.kitalpha.doc.gen.business.core.feature/license.html",
			filename:   "Docgen/business/features/org.polarsys.kitalpha.doc.gen.business.core.feature/license.html",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "Documentation/doc_infocenter/com.ibm.ism.doc/Reference/SpecialCmd/cmd_imaserver_get_licensedusage.dita",
			filename:   "Documentation/doc_infocenter/com.ibm.ism.doc/Reference/SpecialCmd/cmd_imaserver_get_licensedusage.dita",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "GPL-2.0-LICENSE",
			filename:   "GPL-2.0-LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "GPL-LICENSE",
			filename:   "GPL-LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "GPL2LICENCES",
			filename:   "GPL2LICENCES",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "GPLLICENCE",
			filename:   "GPLLICENCE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENCES",
			filename:   "LICENCES",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE,",
			filename:   "LICENSE,",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE-ASL.txt",
			filename:   "LICENSE-ASL.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE-MPL-2.0",
			filename:   "LICENSE-MPL-2.0",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "license-CC-BY-4.0",
			filename:   "license-CC-BY-4.0",
			shouldFail: false,
			extensions: []string{
				"",
				".adoc",
				".asc",
				".docx",
				".doc",
				".ext",
				".html",
				".markdown",
				".md",
				".rst",
				".txt",
				".xml",
			},
		},
		{
			name:       "LICENSE-binary",
			filename:   "LICENSE-binary",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE-junit.txt",
			filename:   "LICENSE-junit.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE.txt",
			filename:   "LICENSE.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSES.txt",
			filename:   "LICENSES.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSES/Apache-2.0.txt",
			filename:   "LICENSES/Apache-2.0.txt",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "LICENSE_RELEASE",
			filename:   "LICENSE_RELEASE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "Licenses/GPL-3.0.md",
			filename:   "Licenses/GPL-3.0.md",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "MIG_kjahdskhaskjdhk_LICENSE",
			filename:   "MIG_kjahdskhaskjdhk_LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "MIT-License",
			filename:   "MIT-License",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "MIT_.0_LICENSE",
			filename:   "MIT_.0_LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "MIT_kjahdskhaskjdhk_LICENSE",
			filename:   "MIT_kjahdskhaskjdhk_LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "UNLICENSE.md",
			filename:   "UNLICENSE.md",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "VotoSerifGX-OFL/Fonts/VotoSerifGX-VarTTF/OFL.txt",
			filename:   "VotoSerifGX-OFL/Fonts/VotoSerifGX-VarTTF/OFL.txt",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "content/dev/apply-license.md",
			filename:   "content/dev/apply-license.md",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "deprecated/3.3/features/org.eclipse.egf.portfolio.genchain.ecoretools.feature/license.html",
			filename:   "deprecated/3.3/features/org.eclipse.egf.portfolio.genchain.ecoretools.feature/license.html",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "examples-trunk/LICENSE",
			filename:   "examples-trunk/LICENSE",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "jetty-artifact-remote-resources/src/main/resources/META-INF/LICENSE",
			filename:   "jetty-artifact-remote-resources/src/main/resources/META-INF/LICENSE",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "license.md",
			filename:   "license.md",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "license.yaml",
			filename:   "license.yaml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "licenserc.yaml",
			filename:   "licenserc.yaml",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "mit-2-0-0-LICENSE",
			filename:   "mit-2-0-0-LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "mit-2.0.0-LICENSE",
			filename:   "mit-2.0.0-LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "mit.2.0.0.LICENSE",
			filename:   "mit.2.0.0.LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "mit_2_0_0_LICENSE",
			filename:   "mit_2_0_0_LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "mit_2_0_LICENSE",
			filename:   "mit_2_0_LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
		{
			name:       "python-phoenixdb/LICENSE",
			filename:   "python-phoenixdb/LICENSE",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "static/fonts/OFL.txt",
			filename:   "static/fonts/OFL.txt",
			shouldFail: true,
			extensions: []string{
				"",
			},
		},
		{
			name:       "unlicense",
			filename:   "unlicense",
			shouldFail: false,
			extensions: []string{
				"",
			},
		},
	}

	//nolint: paralleltest
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		for _, ext := range tt.extensions {
			name := tt.name + ext
			shouldFail := tt.shouldFail
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				s := TestLicense(name)
				if !s && !shouldFail {
					t.Fail()
				} else if s && shouldFail {
					t.Fail()
				}
			})
		}
	}
}
