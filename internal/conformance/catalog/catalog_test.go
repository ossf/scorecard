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

package catalog

import (
	"errors"
	"strings"
	"testing"
)

func TestLookup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr     error
		value       string
		wantVersion string
		wantInError string
	}{
		{value: "osps-baseline", wantVersion: "v2026.08.28"},
		{value: "osps-baseline@v2026.08.28", wantVersion: "v2026.08.28"},
		{value: "osps", wantErr: ErrUnsupportedCatalog, wantInError: "osps-baseline"},
		{value: "", wantErr: ErrUnsupportedCatalog, wantInError: "osps-baseline"},
		{value: "osps-baseline@v2025.01.01", wantErr: ErrUnsupportedVersion, wantInError: "v2026.08.28"},
		{value: "osps-baseline@", wantErr: ErrUnsupportedVersion, wantInError: "v2026.08.28"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			got, err := Lookup(tt.value)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Lookup(%q) error = %v, want %v", tt.value, err, tt.wantErr)
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantInError) {
					t.Errorf("error %q does not list %q", err, tt.wantInError)
				}
				return
			}
			if got.Version != tt.wantVersion {
				t.Errorf("Lookup(%q) version = %s, want %s", tt.value, got.Version, tt.wantVersion)
			}
		})
	}
}

func TestIDs(t *testing.T) {
	t.Parallel()
	ids := IDs()
	if len(ids) != 1 || ids[0] != "osps-baseline" {
		t.Errorf("IDs() = %v, want [osps-baseline]", ids)
	}
}
