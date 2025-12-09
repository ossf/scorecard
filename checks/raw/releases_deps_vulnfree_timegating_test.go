// Copyright 2025 OpenSSF Scorecard Authors
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
	"context"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/clients"
)

// mockOSVClientForTimeGating is a mock OSV client that returns predefined vulnerability records.
type mockOSVClientForTimeGating struct {
	vulns map[string]*clients.OSVVuln
}

func (m *mockOSVClientForTimeGating) QueryBatch(ctx context.Context, queries []clients.OSVQuery) ([][]string, error) {
	return nil, nil
}

func (m *mockOSVClientForTimeGating) GetVuln(ctx context.Context, id string) (*clients.OSVVuln, error) {
	if vuln, ok := m.vulns[id]; ok {
		return vuln, nil
	}
	return &clients.OSVVuln{ID: id}, nil
}

// TestFilterVulnsByPublishTime tests the time-gating logic for vulnerability filtering.
func TestFilterVulnsByPublishTime(t *testing.T) {
	t.Parallel()

	// Test timestamps
	releaseTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	beforeRelease := time.Date(2023, 1, 10, 10, 0, 0, 0, time.UTC)
	atRelease := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	afterRelease := time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC)

	//nolint:govet // fieldalignment: test struct optimized for readability
	tests := []struct {
		mockVulns   map[string]*clients.OSVVuln
		vulnIDs     []string
		expectedIDs []string
		releaseTime time.Time
		name        string
		description string
	}{
		{
			name:        "all vulnerabilities published before release",
			vulnIDs:     []string{"GHSA-1111", "GHSA-2222", "GHSA-3333"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-1111": {ID: "GHSA-1111", Published: &beforeRelease},
				"GHSA-2222": {ID: "GHSA-2222", Published: &beforeRelease},
				"GHSA-3333": {ID: "GHSA-3333", Published: &beforeRelease},
			},
			expectedIDs: []string{"GHSA-1111", "GHSA-2222", "GHSA-3333"},
			description: "All vulnerabilities published before release should be included",
		},
		{
			name:        "vulnerability published exactly at release time",
			vulnIDs:     []string{"GHSA-1111"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-1111": {ID: "GHSA-1111", Published: &atRelease},
			},
			expectedIDs: []string{"GHSA-1111"},
			description: "Vulnerability published at exact release time should be included",
		},
		{
			name:        "all vulnerabilities published after release",
			vulnIDs:     []string{"GHSA-1111", "GHSA-2222"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-1111": {ID: "GHSA-1111", Published: &afterRelease},
				"GHSA-2222": {ID: "GHSA-2222", Published: &afterRelease},
			},
			expectedIDs: []string{},
			description: "Vulnerabilities published after release should be excluded",
		},
		{
			name:        "mixed vulnerabilities - before, at, and after release",
			vulnIDs:     []string{"GHSA-BEFORE", "GHSA-AT", "GHSA-AFTER"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-BEFORE": {ID: "GHSA-BEFORE", Published: &beforeRelease},
				"GHSA-AT":     {ID: "GHSA-AT", Published: &atRelease},
				"GHSA-AFTER":  {ID: "GHSA-AFTER", Published: &afterRelease},
			},
			expectedIDs: []string{"GHSA-BEFORE", "GHSA-AT"},
			description: "Only vulnerabilities published before/at release should be included",
		},
		{
			name:        "vulnerability with nil published timestamp",
			vulnIDs:     []string{"GHSA-1111", "GHSA-NO-TIME"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-1111":    {ID: "GHSA-1111", Published: &beforeRelease},
				"GHSA-NO-TIME": {ID: "GHSA-NO-TIME", Published: nil},
			},
			expectedIDs: []string{"GHSA-1111"},
			description: "Vulnerability with nil published time should be excluded",
		},
		{
			name:        "empty vulnerability list",
			vulnIDs:     []string{},
			releaseTime: releaseTime,
			mockVulns:   map[string]*clients.OSVVuln{},
			expectedIDs: []string{},
			description: "Empty input should return empty output",
		},
		{
			name:        "vulnerability one second before release",
			vulnIDs:     []string{"GHSA-EDGE"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-EDGE": {
					ID:        "GHSA-EDGE",
					Published: timePtr(releaseTime.Add(-1 * time.Second)),
				},
			},
			expectedIDs: []string{"GHSA-EDGE"},
			description: "Vulnerability published one second before release should be included",
		},
		{
			name:        "vulnerability one second after release",
			vulnIDs:     []string{"GHSA-EDGE"},
			releaseTime: releaseTime,
			mockVulns: map[string]*clients.OSVVuln{
				"GHSA-EDGE": {
					ID:        "GHSA-EDGE",
					Published: timePtr(releaseTime.Add(1 * time.Second)),
				},
			},
			expectedIDs: []string{},
			description: "Vulnerability published one second after release should be excluded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockOSV := &mockOSVClientForTimeGating{vulns: tt.mockVulns}
			ctx := context.Background()

			result := filterVulnsByPublishTime(ctx, mockOSV, tt.vulnIDs, tt.releaseTime)

			// Check length
			if len(result) != len(tt.expectedIDs) {
				t.Errorf("%s: got %d vulnerabilities, want %d\nGot: %v\nWant: %v",
					tt.description, len(result), len(tt.expectedIDs), result, tt.expectedIDs)
				return
			}

			// Check each ID is present (order doesn't matter for this test)
			resultMap := make(map[string]bool)
			for _, id := range result {
				resultMap[id] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !resultMap[expectedID] {
					t.Errorf("%s: expected vulnerability %q not found in result %v",
						tt.description, expectedID, result)
				}
			}
		})
	}
}

// TestFilterVulnsByPublishTime_RealWorldScenario tests a realistic scenario.
func TestFilterVulnsByPublishTime_RealWorldScenario(t *testing.T) {
	t.Parallel()

	// Scenario: Project released v2.0.0 on March 1, 2023
	releaseDate := time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC)

	// Vulnerabilities timeline:
	// - CVE-2022-1234: discovered in December 2022 (before release)
	// - CVE-2023-5678: discovered in February 2023 (before release)
	// - CVE-2023-9999: discovered in July 2023 (after release)
	// - GHSA-ABCD: discovered in March 2024 (way after release)

	dec2022 := time.Date(2022, 12, 15, 0, 0, 0, 0, time.UTC)
	feb2023 := time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)
	jul2023 := time.Date(2023, 7, 20, 0, 0, 0, 0, time.UTC)
	mar2024 := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)

	mockOSV := &mockOSVClientForTimeGating{
		vulns: map[string]*clients.OSVVuln{
			"CVE-2022-1234": {ID: "CVE-2022-1234", Published: &dec2022},
			"CVE-2023-5678": {ID: "CVE-2023-5678", Published: &feb2023},
			"CVE-2023-9999": {ID: "CVE-2023-9999", Published: &jul2023},
			"GHSA-ABCD":     {ID: "GHSA-ABCD", Published: &mar2024},
		},
	}

	vulnIDs := []string{"CVE-2022-1234", "CVE-2023-5678", "CVE-2023-9999", "GHSA-ABCD"}
	ctx := context.Background()

	result := filterVulnsByPublishTime(ctx, mockOSV, vulnIDs, releaseDate)

	// Should only include the two vulnerabilities known at release time
	expectedCount := 2
	if len(result) != expectedCount {
		t.Errorf("Real-world scenario: got %d vulnerabilities, want %d\nGot: %v",
			len(result), expectedCount, result)
	}

	// Verify the correct vulnerabilities are included
	found2022 := false
	found2023Feb := false
	for _, id := range result {
		if id == "CVE-2022-1234" {
			found2022 = true
		}
		if id == "CVE-2023-5678" {
			found2023Feb = true
		}
		if id == "CVE-2023-9999" || id == "GHSA-ABCD" {
			t.Errorf("Real-world scenario: vulnerability %q from after release should not be included", id)
		}
	}

	if !found2022 {
		t.Error("Real-world scenario: CVE-2022-1234 (Dec 2022) should be included")
	}
	if !found2023Feb {
		t.Error("Real-world scenario: CVE-2023-5678 (Feb 2023) should be included")
	}
}

// timePtr is a helper to get a pointer to a time.Time.
func timePtr(t time.Time) *time.Time {
	return &t
}
