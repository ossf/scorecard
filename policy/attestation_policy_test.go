package policy

import (
	"encoding/json"
	"errors"
	"testing"
)

func (a AttestationPolicy) ToJson() string {
	jsonbytes, err := json.Marshal(a)

	if err != nil {
		return ""
	}

	return string(jsonbytes)
}

func TestAttestationPolicyRead(t *testing.T) {
	t.Parallel()

	// nolint
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
				PreventBinaryArtifacts:      true,
				AllowedBinaryArtifacts:      []string{},
				EnsureNoVulnerabilities:     true,
				EnsurePinnedDependencies:    true,
				AllowedUnpinnedDependencies: []Dependency{},
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
			if p.ToJson() != tt.result.ToJson() {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
