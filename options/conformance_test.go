// Copyright 2026 OpenSSF Scorecard Authors
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

package options

import (
	"errors"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/internal/conformance/catalog"
)

func conformanceOptions() *Options {
	return &Options{
		Repo:     "github.com/ossf/scorecard",
		Commit:   DefaultCommit,
		Format:   FormatJSON,
		FileMode: FileModeArchive,
		Catalog:  "osps-baseline",
		Profile:  "maturity-1",
	}
}

func TestValidateConformance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		modify      func(*Options)
		wantErr     error
		name        string
		wantInError string
	}{
		{
			name:   "valid json",
			modify: func(*Options) {},
		},
		{
			name:   "valid gemara with release version",
			modify: func(o *Options) { o.Format = FormatGemara; o.Catalog = "osps-baseline@v2026.08.28" },
		},
		{
			name:    "local directory",
			modify:  func(o *Options) { o.Repo = ""; o.Local = "." },
			wantErr: errLocalWithCatalog,
		},
		{
			name:    "commit other than HEAD",
			modify:  func(o *Options) { o.Commit = "0123456789abcdef0123456789abcdef01234567" },
			wantErr: errCommitWithCatalog,
		},
		{
			name:        "unknown catalog",
			modify:      func(o *Options) { o.Catalog = "osps" },
			wantErr:     catalog.ErrUnsupportedCatalog,
			wantInError: "osps-baseline",
		},
		{
			name:        "unknown release",
			modify:      func(o *Options) { o.Catalog = "osps-baseline@v2025.01.01" },
			wantErr:     catalog.ErrUnsupportedVersion,
			wantInError: "v2026.08.28",
		},
		{
			name:    "profile missing",
			modify:  func(o *Options) { o.Profile = "" },
			wantErr: errProfileRequired,
		},
		{
			name:        "unknown profile",
			modify:      func(o *Options) { o.Profile = "maturity1" },
			wantErr:     errUnsupportedProfile,
			wantInError: "maturity-1, maturity-2, maturity-3",
		},
		{
			name:    "profile without catalog",
			modify:  func(o *Options) { o.Catalog = ""; o.Format = FormatDefault },
			wantErr: errProfileRequiresCatalog,
		},
		{
			name:    "gemara without catalog",
			modify:  func(o *Options) { o.Catalog = ""; o.Profile = ""; o.Format = FormatGemara },
			wantErr: errGemaraRequiresCatalog,
		},
		{
			name:    "format omitted",
			modify:  func(o *Options) { o.Format = FormatDefault },
			wantErr: errCatalogRequiresFormat,
		},
		{
			name:    "probe format",
			modify:  func(o *Options) { o.Format = FormatProbe },
			wantErr: errCatalogRequiresFormat,
		},
		{
			name:    "checks",
			modify:  func(o *Options) { o.ChecksToRun = []string{"Branch-Protection"} },
			wantErr: errChecksWithCatalog,
		},
		{
			name:    "probes",
			modify:  func(o *Options) { o.ProbesToRun = []string{"requiresPRsToChangeCode"} },
			wantErr: errProbesWithCatalog,
		},
		{
			name:    "policy",
			modify:  func(o *Options) { o.PolicyFile = "policy.yml" },
			wantErr: errPolicyWithCatalog,
		},
		{
			name:    "repos",
			modify:  func(o *Options) { o.Repo = ""; o.Repos = []string{"ossf/scorecard", "ossf/allstar"} },
			wantErr: errReposWithCatalog,
		},
		{
			name:    "org",
			modify:  func(o *Options) { o.Repo = ""; o.Org = "ossf" },
			wantErr: errOrgWithCatalog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := conformanceOptions()
			tt.modify(o)
			errs := o.validateConformance()
			if tt.wantErr == nil {
				if len(errs) != 0 {
					t.Fatalf("validateConformance() = %v, want no errors", errs)
				}
				return
			}
			err := errors.Join(errs...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateConformance() = %v, want %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Errorf("error %q does not list %q", err, tt.wantInError)
			}
		})
	}
}

// TestValidateWithoutCatalog checks that the conformance options add no
// validation errors to runs without --catalog.
func TestValidateWithoutCatalog(t *testing.T) {
	t.Parallel()
	o := conformanceOptions()
	o.Catalog = ""
	o.Profile = ""
	for _, format := range []string{FormatDefault, FormatJSON, FormatProbe, FormatInToto} {
		o.Format = format
		if err := o.Validate(); err != nil {
			t.Errorf("Validate() with --format=%s = %v, want nil", format, err)
		}
	}
}

func TestValidateReportsConformanceErrors(t *testing.T) {
	t.Parallel()
	o := conformanceOptions()
	o.Profile = ""
	err := o.Validate()
	if !errors.Is(err, errValidate) || !strings.Contains(err.Error(), errProfileRequired.Error()) {
		t.Errorf("Validate() = %v, want it to report %q", err, errProfileRequired)
	}
}
