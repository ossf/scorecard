package policy

import (
	"encoding/json"
	"errors"
	"testing"

	sce "github.com/ossf/scorecard/v4/errors"
)

func (a AttestationPolicy) ToJSON() string {
	jsonbytes, err := json.Marshal(a)
	if err != nil {
		return ""
	}

	return string(jsonbytes)
}

func TestAttestationPolicyRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		name     string
		filename string
		result   AttestationPolicy
	}{
		{
			name:     "default attestation policy with everything on",
			filename: "./testdata/policy-binauthz.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   []string{},
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       true,
			},
		},
		{
			name:     "invalid attestation policy",
			filename: "./testdata/policy-binauthz-invalid.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-allowlist.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   []string{"/a/b/c", "d"},
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       true,
			},
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-missingparam.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   nil,
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       false,
			},
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := ParseAttestationPolicyFromFile(tt.filename)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if p.ToJSON() != tt.result.ToJSON() {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
