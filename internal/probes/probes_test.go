// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package probes

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func emptyImpl(r *checker.RawResults) ([]finding.Finding, string, error) {
	return nil, "", nil
}

func emptyIndependentImpl(c *checker.CheckRequest) ([]finding.Finding, string, error) {
	return nil, "", nil
}

var (
	p1 = Probe{
		Name:            "someProbe1",
		Implementation:  emptyImpl,
		RequiredRawData: []CheckName{BinaryArtifacts},
	}

	p2 = Probe{
		Name:            "someProbe2",
		Implementation:  emptyImpl,
		RequiredRawData: []CheckName{BranchProtection},
	}
)

//nolint:paralleltest // registration isn't safe for concurrent use
func Test_register(t *testing.T) {
	tests := []struct {
		name    string
		probe   Probe
		wantErr bool
	}{
		{
			name: "name is required",
			probe: Probe{
				Name:            "",
				Implementation:  emptyImpl,
				RequiredRawData: []CheckName{BinaryArtifacts},
			},
			wantErr: true,
		},
		{
			name: "implementation is required",
			probe: Probe{
				Name:            "foo",
				Implementation:  nil,
				RequiredRawData: []CheckName{BinaryArtifacts},
			},
			wantErr: true,
		},
		{
			name: "raw check data is required",
			probe: Probe{
				Name:            "foo",
				Implementation:  emptyImpl,
				RequiredRawData: []CheckName{},
			},
			wantErr: true,
		},
		{
			name: "valid registration",
			probe: Probe{
				Name:            "foo",
				Implementation:  emptyImpl,
				RequiredRawData: []CheckName{BinaryArtifacts},
			},
			wantErr: false,
		},
		{
			name: "independent probe registration",
			probe: Probe{
				Name:                      "bar",
				IndependentImplementation: emptyIndependentImpl,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := register(tt.probe)
			if err != nil != tt.wantErr {
				t.Fatalf("got err: %v, wanted err: %t", err, tt.wantErr)
			}
		})
	}
}

func setupControlledProbes(t *testing.T) {
	t.Helper()
	err := register(p1)
	if err != nil {
		t.Fatalf("unable to register someProbe1")
	}
	err = register(p2)
	if err != nil {
		t.Fatalf("unable to register someProbe2")
	}
}

//nolint:paralleltest // registration isn't safe for concurrent use
func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		probeName string
		expected  Probe
		wantErr   bool
	}{
		{
			name:      "probe is found",
			probeName: p1.Name,
			expected:  p1,
			wantErr:   false,
		},
		{
			name:      "probe not found",
			probeName: "noProbeCalledThis",
			wantErr:   true,
		},
	}
	setupControlledProbes(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p, err := Get(tt.probeName)
			if err != nil != tt.wantErr {
				t.Fatalf("got err: %v, wanted err: %t", err, tt.wantErr)
			}
			if diff := cmp.Diff(p.Name, tt.expected.Name); diff != "" {
				t.Error("probes didn't match: " + diff)
			}
		})
	}
}
